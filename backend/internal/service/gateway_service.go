package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"lingshu/backend/internal/billing"
	redisstore "lingshu/backend/internal/redis"
	"lingshu/backend/internal/repository"
	"lingshu/backend/internal/upstream"
)

type GatewayService struct {
	repo          repository.GatewayRepository
	frozen        redisstore.FrozenStore
	defaultTokens int
}

type ChatRequest struct {
	Model          string `json:"model"`
	MaxTokens      int64  `json:"max_tokens"`
	Stream         bool   `json:"stream"`
	N              int64  `json:"n"`
	User           string `json:"user"`
	SessionID      string `json:"session_id"`
	ConversationID string `json:"conversation_id"`
}

type GatewayPrincipal struct {
	UserID           string
	APIKeyID         string
	Balance          string
	RPMLimit         int
	ConcurrencyLimit int
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrNoHealthyChannel    = errors.New("no healthy upstream channel")
	ErrRateLimited         = errors.New("rate limited")
)

const (
	channelFailureCooldown   = 30 * time.Second
	channelRateLimitCooldown = 2 * time.Minute
	channelOverloadCooldown  = 5 * time.Minute
)

type UpstreamError struct {
	StatusCode int
	Body       []byte
}

func (e *UpstreamError) Error() string {
	if len(e.Body) == 0 {
		return http.StatusText(e.StatusCode)
	}
	return strings.TrimSpace(string(e.Body))
}

func NewGatewayService(repo repository.GatewayRepository, frozen redisstore.FrozenStore, defaultTokens int) GatewayService {
	if defaultTokens <= 0 {
		defaultTokens = 4096
	}
	return GatewayService{repo: repo, frozen: frozen, defaultTokens: defaultTokens}
}

func (s GatewayService) Models(ctx context.Context) ([]repository.GatewayModel, error) {
	return s.repo.ListEnabledModels(ctx)
}

func (s GatewayService) Chat(ctx context.Context, principal GatewayPrincipal, rawBody []byte, clientIP, sessionID string) (int, []byte, error) {
	start := time.Now()
	var req ChatRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return http.StatusBadRequest, nil, err
	}
	if req.Stream {
		return http.StatusBadRequest, nil, errors.New("use streaming path")
	}
	model, err := s.repo.FindEnabledModel(ctx, req.Model)
	if err != nil {
		return http.StatusNotFound, nil, errors.New("model not found")
	}
	releaseKey, err := s.acquireKeyLimits(ctx, principal)
	if err != nil {
		return http.StatusTooManyRequests, nil, err
	}
	defer releaseKey()

	multiplierUnits, _ := billing.DecimalStringToUnits(model.RateMultiplier)
	estimatedInput := billing.EstimateTokens(string(rawBody))
	if req.MaxTokens <= 0 {
		req.MaxTokens = int64(s.defaultTokens)
		rawBody = bodyWithMaxTokens(rawBody, s.defaultTokens)
	}
	estimate := estimateChargeForModel(model, req, estimatedInput, multiplierUnits)
	balanceUnits, _ := billing.DecimalStringToUnits(principal.Balance)
	if err := billing.Reserve(ctx, s.frozen, principal.UserID, balanceUnits, estimate.Charge); err != nil {
		return http.StatusPaymentRequired, nil, ErrInsufficientBalance
	}
	defer billing.Release(ctx, s.frozen, principal.UserID, estimate.Charge)

	channel, upstreamResp, err := s.forwardWithRetry(ctx, model.ID, rawBody, stickyKey(principal, req, sessionID))
	if err != nil {
		statusCode := http.StatusBadGateway
		var upstreamErr *UpstreamError
		if errors.As(err, &upstreamErr) {
			statusCode = upstreamErr.StatusCode
		}
		_ = s.repo.RecordAndCharge(ctx, repository.GatewayRequestRecord{
			RequestID:         uuid.NewString(),
			UserID:            principal.UserID,
			APIKeyID:          principal.APIKeyID,
			ModelID:           model.ID,
			ChannelID:         channel.ID,
			UpstreamModelName: channel.UpstreamModelName,
			Endpoint:          "/v1/chat/completions",
			Status:            "failed",
			HTTPStatus:        statusCode,
			BaseCost:          "0.000000",
			RateMultiplier:    model.RateMultiplier,
			Charge:            "0.000000",
			ErrorCode:         "upstream_error",
			ErrorMessage:      err.Error(),
			ClientIP:          clientIP,
			LatencyMS:         repository.NowMS(start),
		})
		if errors.Is(err, ErrRateLimited) {
			return http.StatusTooManyRequests, nil, err
		}
		if upstreamErr != nil {
			return upstreamErr.StatusCode, NormalizeUpstreamErrorBody(upstreamErr.StatusCode, upstreamErr.Body), nil
		}
		return http.StatusBadGateway, nil, err
	}

	usage := upstreamResp.Usage
	isEstimated := false
	if usage.TotalTokens == 0 {
		isEstimated = true
		usage.PromptTokens = int(estimatedInput)
		usage.CompletionTokens = int(req.MaxTokens)
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	actual := actualChargeForModel(model, req, usage, multiplierUnits)
	status := "success"
	if upstreamResp.StatusCode >= 400 {
		status = "failed"
		actual = billing.Charge{BaseCost: 0, RateMultiplier: multiplierUnits, Charge: 0}
	}
	if err := s.repo.RecordAndCharge(ctx, repository.GatewayRequestRecord{
		RequestID:           uuid.NewString(),
		UserID:              principal.UserID,
		APIKeyID:            principal.APIKeyID,
		ModelID:             model.ID,
		ChannelID:           channel.ID,
		Endpoint:            "/v1/chat/completions",
		Status:              status,
		HTTPStatus:          upstreamResp.StatusCode,
		PromptTokens:        usage.PromptTokens,
		CompletionTokens:    usage.CompletionTokens,
		TotalTokens:         usage.TotalTokens,
		CacheCreationTokens: usage.CacheCreationTokens,
		CacheReadTokens:     usage.CacheReadTokens,
		ImageOutputTokens:   usage.ImageOutputTokens,
		BaseCost:            billing.UnitsToDecimalString(actual.BaseCost),
		RateMultiplier:      model.RateMultiplier,
		Charge:              billing.UnitsToDecimalString(actual.Charge),
		IsEstimated:         isEstimated,
		LatencyMS:           repository.NowMS(start),
		UpstreamModelName:   channel.UpstreamModelName,
		ClientIP:            clientIP,
	}); err != nil {
		if errors.Is(err, repository.ErrSettlementInsufficientBalance) {
			return http.StatusPaymentRequired, nil, ErrInsufficientBalance
		}
		return http.StatusInternalServerError, nil, err
	}
	return upstreamResp.StatusCode, upstreamResp.Body, nil
}

func (s GatewayService) Embeddings(ctx context.Context, principal GatewayPrincipal, rawBody []byte, clientIP, sessionID string) (int, []byte, error) {
	start := time.Now()
	var req ChatRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return http.StatusBadRequest, nil, err
	}
	model, err := s.repo.FindEnabledModel(ctx, req.Model)
	if err != nil {
		return http.StatusNotFound, nil, errors.New("model not found")
	}
	releaseKey, err := s.acquireKeyLimits(ctx, principal)
	if err != nil {
		return http.StatusTooManyRequests, nil, err
	}
	defer releaseKey()

	multiplierUnits, _ := billing.DecimalStringToUnits(model.RateMultiplier)
	estimatedInput := billing.EstimateTokens(string(rawBody))
	estimate := estimateChargeForModel(model, req, estimatedInput, multiplierUnits)
	balanceUnits, _ := billing.DecimalStringToUnits(principal.Balance)
	if err := billing.Reserve(ctx, s.frozen, principal.UserID, balanceUnits, estimate.Charge); err != nil {
		return http.StatusPaymentRequired, nil, ErrInsufficientBalance
	}
	defer billing.Release(ctx, s.frozen, principal.UserID, estimate.Charge)

	channel, upstreamResp, err := s.forwardEmbeddingsWithRetry(ctx, model.ID, rawBody, stickyKey(principal, req, sessionID))
	if err != nil {
		statusCode := http.StatusBadGateway
		var upstreamErr *UpstreamError
		if errors.As(err, &upstreamErr) {
			statusCode = upstreamErr.StatusCode
		}
		_ = s.repo.RecordAndCharge(ctx, repository.GatewayRequestRecord{
			RequestID:         uuid.NewString(),
			UserID:            principal.UserID,
			APIKeyID:          principal.APIKeyID,
			ModelID:           model.ID,
			ChannelID:         channel.ID,
			UpstreamModelName: channel.UpstreamModelName,
			Endpoint:          "/v1/embeddings",
			Status:            "failed",
			HTTPStatus:        statusCode,
			BaseCost:          "0.000000",
			RateMultiplier:    model.RateMultiplier,
			Charge:            "0.000000",
			ErrorCode:         "upstream_error",
			ErrorMessage:      err.Error(),
			ClientIP:          clientIP,
			LatencyMS:         repository.NowMS(start),
		})
		if errors.Is(err, ErrRateLimited) {
			return http.StatusTooManyRequests, nil, err
		}
		if upstreamErr != nil {
			return upstreamErr.StatusCode, NormalizeUpstreamErrorBody(upstreamErr.StatusCode, upstreamErr.Body), nil
		}
		return http.StatusBadGateway, nil, err
	}

	usage := upstreamResp.Usage
	isEstimated := false
	if usage.TotalTokens == 0 {
		isEstimated = true
		usage.PromptTokens = int(estimatedInput)
		usage.CompletionTokens = 0
		usage.TotalTokens = usage.PromptTokens
	}
	actual := actualChargeForModel(model, req, usage, multiplierUnits)
	status := "success"
	if upstreamResp.StatusCode >= 400 {
		status = "failed"
		actual = billing.Charge{BaseCost: 0, RateMultiplier: multiplierUnits, Charge: 0}
	}
	if err := s.repo.RecordAndCharge(ctx, repository.GatewayRequestRecord{
		RequestID:           uuid.NewString(),
		UserID:              principal.UserID,
		APIKeyID:            principal.APIKeyID,
		ModelID:             model.ID,
		ChannelID:           channel.ID,
		Endpoint:            "/v1/embeddings",
		Status:              status,
		HTTPStatus:          upstreamResp.StatusCode,
		PromptTokens:        usage.PromptTokens,
		CompletionTokens:    usage.CompletionTokens,
		TotalTokens:         usage.TotalTokens,
		CacheCreationTokens: usage.CacheCreationTokens,
		CacheReadTokens:     usage.CacheReadTokens,
		ImageOutputTokens:   usage.ImageOutputTokens,
		BaseCost:            billing.UnitsToDecimalString(actual.BaseCost),
		RateMultiplier:      model.RateMultiplier,
		Charge:              billing.UnitsToDecimalString(actual.Charge),
		IsEstimated:         isEstimated,
		LatencyMS:           repository.NowMS(start),
		UpstreamModelName:   channel.UpstreamModelName,
		ClientIP:            clientIP,
	}); err != nil {
		if errors.Is(err, repository.ErrSettlementInsufficientBalance) {
			return http.StatusPaymentRequired, nil, ErrInsufficientBalance
		}
		return http.StatusInternalServerError, nil, err
	}
	return upstreamResp.StatusCode, upstreamResp.Body, nil
}

func (s GatewayService) OpenChatStream(ctx context.Context, principal GatewayPrincipal, rawBody []byte, sessionID string) (repository.GatewayModel, repository.GatewayChannel, int64, *http.Response, error) {
	var req ChatRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return repository.GatewayModel{}, repository.GatewayChannel{}, 0, nil, err
	}
	model, err := s.repo.FindEnabledModel(ctx, req.Model)
	if err != nil {
		return repository.GatewayModel{}, repository.GatewayChannel{}, 0, nil, errors.New("model not found")
	}
	releaseKey, err := s.acquireKeyLimits(ctx, principal)
	if err != nil {
		return repository.GatewayModel{}, repository.GatewayChannel{}, 0, nil, err
	}
	multiplierUnits, _ := billing.DecimalStringToUnits(model.RateMultiplier)
	estimatedInput := billing.EstimateTokens(string(rawBody))
	if req.MaxTokens <= 0 {
		req.MaxTokens = int64(s.defaultTokens)
		rawBody = bodyWithMaxTokens(rawBody, s.defaultTokens)
	}
	estimate := estimateChargeForModel(model, req, estimatedInput, multiplierUnits)
	balanceUnits, _ := billing.DecimalStringToUnits(principal.Balance)
	if err := billing.Reserve(ctx, s.frozen, principal.UserID, balanceUnits, estimate.Charge); err != nil {
		releaseKey()
		return repository.GatewayModel{}, repository.GatewayChannel{}, 0, nil, err
	}
	channel, resp, err := s.openStreamWithRetry(ctx, model.ID, rawBody, stickyKey(principal, req, sessionID))
	if err != nil {
		billing.Release(ctx, s.frozen, principal.UserID, estimate.Charge)
		releaseKey()
		return repository.GatewayModel{}, repository.GatewayChannel{}, 0, nil, err
	}
	return model, channel, estimate.Charge, resp, nil
}

func (s GatewayService) FinalizeStream(ctx context.Context, principal GatewayPrincipal, model repository.GatewayModel, channel repository.GatewayChannel, rawBody []byte, responseBody []byte, estimate int64, statusCode int, clientIP string, start time.Time, firstTokenMS int) {
	s.frozen.ReleaseConcurrency(ctx, "key:"+principal.APIKeyID)
	s.frozen.ReleaseConcurrency(ctx, "channel:"+channel.ID)
	defer billing.Release(ctx, s.frozen, principal.UserID, estimate)
	multiplierUnits, _ := billing.DecimalStringToUnits(model.RateMultiplier)
	// 优先使用上游在 SSE 末帧回灌的真实 usage；取不到才退回本地 tiktoken 估算。
	isEstimated := false
	usage := upstream.ExtractStreamUsage(string(responseBody))
	if usage.TotalTokens == 0 {
		isEstimated = true
		usage = upstream.Usage{
			PromptTokens:     int(billing.EstimateTokens(string(rawBody))),
			CompletionTokens: int(billing.EstimateStreamTokens(string(responseBody))),
		}
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	var req ChatRequest
	_ = json.Unmarshal(rawBody, &req)
	actual := actualChargeForModel(model, req, usage, multiplierUnits)
	status := "success"
	if statusCode >= 400 {
		status = "failed"
		actual = billing.Charge{BaseCost: 0, RateMultiplier: multiplierUnits, Charge: 0}
	}
	_ = s.repo.RecordAndCharge(ctx, repository.GatewayRequestRecord{
		RequestID:           uuid.NewString(),
		UserID:              principal.UserID,
		APIKeyID:            principal.APIKeyID,
		ModelID:             model.ID,
		ChannelID:           channel.ID,
		Endpoint:            "/v1/chat/completions",
		Status:              status,
		HTTPStatus:          statusCode,
		PromptTokens:        usage.PromptTokens,
		CompletionTokens:    usage.CompletionTokens,
		TotalTokens:         usage.TotalTokens,
		CacheCreationTokens: usage.CacheCreationTokens,
		CacheReadTokens:     usage.CacheReadTokens,
		ImageOutputTokens:   usage.ImageOutputTokens,
		BaseCost:            billing.UnitsToDecimalString(actual.BaseCost),
		RateMultiplier:      model.RateMultiplier,
		Charge:              billing.UnitsToDecimalString(actual.Charge),
		IsStream:            true,
		IsEstimated:         isEstimated,
		LatencyMS:           repository.NowMS(start),
		FirstTokenMS:        firstTokenMS,
		UpstreamModelName:   channel.UpstreamModelName,
		ClientIP:            clientIP,
	})
}

func (s GatewayService) acquireKeyLimits(ctx context.Context, principal GatewayPrincipal) (func(), error) {
	ok, err := s.frozen.AllowRPM(ctx, "key:"+principal.APIKeyID, principal.RPMLimit)
	if err != nil {
		return func() {}, err
	}
	if !ok {
		return func() {}, ErrRateLimited
	}
	ok, err = s.frozen.TryAcquire(ctx, "key:"+principal.APIKeyID, principal.ConcurrencyLimit, 5*time.Minute)
	if err != nil {
		return func() {}, err
	}
	if !ok {
		return func() {}, ErrRateLimited
	}
	return func() { s.frozen.ReleaseConcurrency(ctx, "key:"+principal.APIKeyID) }, nil
}

func estimateChargeForModel(model repository.GatewayModel, req ChatRequest, estimatedInput int64, multiplierUnits int64) billing.Charge {
	if model.BillingMode == "per_call" {
		priceUnits, _ := billing.DecimalStringToUnits(model.PricePerCall)
		return billing.CalculatePerCallCharge(priceUnits, callCount(req), multiplierUnits)
	}
	inputUnits, _ := billing.DecimalStringToUnits(model.InputPricePer1K)
	outputUnits, _ := billing.DecimalStringToUnits(model.OutputPricePer1K)
	cacheCreationUnits, _ := billing.DecimalStringToUnits(model.CacheCreationPricePer1K)
	cacheReadUnits, _ := billing.DecimalStringToUnits(model.CacheReadPricePer1K)
	return billing.CalculateTokenCharge(
		billing.TokenPricing{InputPricePer1K: inputUnits, OutputPricePer1K: outputUnits, CacheCreationPricePer1K: cacheCreationUnits, CacheReadPricePer1K: cacheReadUnits, RateMultiplier: multiplierUnits},
		billing.TokenUsage{InputTokens: estimatedInput, OutputTokens: req.MaxTokens},
	)
}

func actualChargeForModel(model repository.GatewayModel, req ChatRequest, usage upstream.Usage, multiplierUnits int64) billing.Charge {
	if model.BillingMode == "per_call" {
		priceUnits, _ := billing.DecimalStringToUnits(model.PricePerCall)
		return billing.CalculatePerCallCharge(priceUnits, callCount(req), multiplierUnits)
	}
	inputUnits, _ := billing.DecimalStringToUnits(model.InputPricePer1K)
	outputUnits, _ := billing.DecimalStringToUnits(model.OutputPricePer1K)
	cacheCreationUnits, _ := billing.DecimalStringToUnits(model.CacheCreationPricePer1K)
	cacheReadUnits, _ := billing.DecimalStringToUnits(model.CacheReadPricePer1K)
	return billing.CalculateTokenCharge(
		billing.TokenPricing{InputPricePer1K: inputUnits, OutputPricePer1K: outputUnits, CacheCreationPricePer1K: cacheCreationUnits, CacheReadPricePer1K: cacheReadUnits, RateMultiplier: multiplierUnits},
		billing.TokenUsage{InputTokens: int64(usage.PromptTokens), OutputTokens: int64(usage.CompletionTokens), CacheCreationTokens: int64(usage.CacheCreationTokens), CacheReadTokens: int64(usage.CacheReadTokens)},
	)
}

func callCount(req ChatRequest) int64 {
	if req.N <= 0 {
		return 1
	}
	return req.N
}

func (s GatewayService) forwardWithRetry(ctx context.Context, modelID string, rawBody []byte, sticky string) (repository.GatewayChannel, upstream.ChatResponse, error) {
	channels, err := s.repo.ListCandidateChannels(ctx, modelID)
	if err != nil {
		return repository.GatewayChannel{}, upstream.ChatResponse{}, err
	}
	if len(channels) == 0 {
		return repository.GatewayChannel{}, upstream.ChatResponse{}, ErrNoHealthyChannel
	}
	var lastChannel repository.GatewayChannel
	var lastErr error = ErrNoHealthyChannel
	limited := true
	excluded := map[string]struct{}{}
	ordered := s.orderChannels(ctx, sticky, channels, excluded)
	if len(ordered) == 0 {
		return lastChannel, upstream.ChatResponse{}, ErrNoHealthyChannel
	}
	for _, channel := range ordered {
		lastChannel = channel
		ok, err := s.acquireChannel(ctx, channel)
		if err != nil {
			lastErr = err
			continue
		}
		if !ok {
			continue
		}
		limited = false
		resp, err := s.forwardOnce(ctx, channel, rawBody)
		s.frozen.ReleaseConcurrency(ctx, "channel:"+channel.ID)
		if err != nil {
			lastErr = err
			s.markChannelFailure(ctx, channel, 0, err.Error(), excluded)
			continue
		}
		if shouldRetryStatus(resp.StatusCode) {
			lastErr = &UpstreamError{StatusCode: resp.StatusCode, Body: resp.Body}
			s.markChannelFailure(ctx, channel, resp.StatusCode, lastErr.Error(), excluded)
			continue
		}
		_ = s.repo.MarkChannelSuccess(ctx, channel.ID)
		s.rememberSticky(ctx, sticky, channel.ID)
		return channel, resp, nil
	}
	if limited {
		return lastChannel, upstream.ChatResponse{}, ErrRateLimited
	}
	return lastChannel, upstream.ChatResponse{}, lastErr
}

func (s GatewayService) forwardEmbeddingsWithRetry(ctx context.Context, modelID string, rawBody []byte, sticky string) (repository.GatewayChannel, upstream.ChatResponse, error) {
	channels, err := s.repo.ListCandidateChannels(ctx, modelID)
	if err != nil {
		return repository.GatewayChannel{}, upstream.ChatResponse{}, err
	}
	if len(channels) == 0 {
		return repository.GatewayChannel{}, upstream.ChatResponse{}, ErrNoHealthyChannel
	}
	var lastChannel repository.GatewayChannel
	var lastErr error = ErrNoHealthyChannel
	limited := true
	excluded := map[string]struct{}{}
	ordered := s.orderChannels(ctx, sticky, channels, excluded)
	if len(ordered) == 0 {
		return lastChannel, upstream.ChatResponse{}, ErrNoHealthyChannel
	}
	for _, channel := range ordered {
		lastChannel = channel
		ok, err := s.acquireChannel(ctx, channel)
		if err != nil {
			lastErr = err
			continue
		}
		if !ok {
			continue
		}
		limited = false
		resp, err := s.forwardEmbeddingsOnce(ctx, channel, rawBody)
		s.frozen.ReleaseConcurrency(ctx, "channel:"+channel.ID)
		if err != nil {
			lastErr = err
			s.markChannelFailure(ctx, channel, 0, err.Error(), excluded)
			continue
		}
		if shouldRetryStatus(resp.StatusCode) {
			lastErr = &UpstreamError{StatusCode: resp.StatusCode, Body: resp.Body}
			s.markChannelFailure(ctx, channel, resp.StatusCode, lastErr.Error(), excluded)
			continue
		}
		_ = s.repo.MarkChannelSuccess(ctx, channel.ID)
		s.rememberSticky(ctx, sticky, channel.ID)
		return channel, resp, nil
	}
	if limited {
		return lastChannel, upstream.ChatResponse{}, ErrRateLimited
	}
	return lastChannel, upstream.ChatResponse{}, lastErr
}

func (s GatewayService) openStreamWithRetry(ctx context.Context, modelID string, rawBody []byte, sticky string) (repository.GatewayChannel, *http.Response, error) {
	channels, err := s.repo.ListCandidateChannels(ctx, modelID)
	if err != nil {
		return repository.GatewayChannel{}, nil, err
	}
	if len(channels) == 0 {
		return repository.GatewayChannel{}, nil, ErrNoHealthyChannel
	}
	var lastChannel repository.GatewayChannel
	var lastErr error = ErrNoHealthyChannel
	limited := true
	excluded := map[string]struct{}{}
	ordered := s.orderChannels(ctx, sticky, channels, excluded)
	if len(ordered) == 0 {
		return lastChannel, nil, ErrNoHealthyChannel
	}
	for _, channel := range ordered {
		lastChannel = channel
		ok, err := s.acquireChannel(ctx, channel)
		if err != nil {
			lastErr = err
			continue
		}
		if !ok {
			continue
		}
		limited = false
		resp, err := s.openStreamOnce(ctx, channel, rawBody)
		if err != nil {
			s.frozen.ReleaseConcurrency(ctx, "channel:"+channel.ID)
			lastErr = err
			s.markChannelFailure(ctx, channel, 0, err.Error(), excluded)
			continue
		}
		if shouldRetryStatus(resp.StatusCode) {
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				body = []byte(readErr.Error())
			}
			lastErr = &UpstreamError{StatusCode: resp.StatusCode, Body: body}
			s.markChannelFailure(ctx, channel, resp.StatusCode, lastErr.Error(), excluded)
			s.frozen.ReleaseConcurrency(ctx, "channel:"+channel.ID)
			continue
		}
		_ = s.repo.MarkChannelSuccess(ctx, channel.ID)
		s.rememberSticky(ctx, sticky, channel.ID)
		return channel, resp, nil
	}
	if limited {
		return lastChannel, nil, ErrRateLimited
	}
	return lastChannel, nil, lastErr
}

func (s GatewayService) acquireChannel(ctx context.Context, channel repository.GatewayChannel) (bool, error) {
	ok, err := s.frozen.AllowRPM(ctx, "channel:"+channel.ID, channel.RPMLimit)
	if err != nil || !ok {
		return ok, err
	}
	return s.frozen.TryAcquire(ctx, "channel:"+channel.ID, channel.ConcurrencyLimit, 5*time.Minute)
}

func (s GatewayService) orderChannels(ctx context.Context, sticky string, channels []repository.GatewayChannel, excluded map[string]struct{}) []repository.GatewayChannel {
	ordered := make([]repository.GatewayChannel, 0, len(channels))
	remaining := make([]repository.GatewayChannel, 0, len(channels))
	for _, channel := range channels {
		if _, ok := excluded[channel.ID]; ok {
			continue
		}
		if cooling, err := s.frozen.IsChannelCooling(ctx, channel.ID); err == nil && cooling {
			continue
		}
		if limited, err := s.frozen.IsChannelRateLimited(ctx, channel.ID); err == nil && limited {
			continue
		}
		if overloaded, err := s.frozen.IsChannelOverloaded(ctx, channel.ID); err == nil && overloaded {
			continue
		}
		remaining = append(remaining, channel)
	}
	if sticky != "" {
		if stickyChannelID, err := s.frozen.GetStickyChannel(ctx, sticky); err == nil && stickyChannelID != "" {
			for i, channel := range remaining {
				if channel.ID == stickyChannelID {
					ordered = append(ordered, channel)
					remaining = append(remaining[:i], remaining[i+1:]...)
					break
				}
			}
		}
	}
	ordered = append(ordered, weightedRandomOrder(remaining)...)
	return ordered
}

func (s GatewayService) markChannelFailure(ctx context.Context, channel repository.GatewayChannel, status int, message string, excluded map[string]struct{}) {
	if channel.ID == "" {
		return
	}
	excluded[channel.ID] = struct{}{}
	_ = s.repo.MarkChannelFailure(ctx, channel.ID, message)
	cooldown := channelCooldownForStatus(status)
	_ = s.frozen.SetChannelCooldown(ctx, channel.ID, message, cooldown)
	if status == http.StatusTooManyRequests {
		_ = s.frozen.SetChannelRateLimited(ctx, channel.ID, message, cooldown)
	}
	if status == 529 {
		_ = s.frozen.SetChannelOverloaded(ctx, channel.ID, message, cooldown)
	}
}

func (s GatewayService) rememberSticky(ctx context.Context, sticky, channelID string) {
	if sticky == "" || channelID == "" {
		return
	}
	_ = s.frozen.SetStickyChannel(ctx, sticky, channelID, 30*time.Minute)
}

func weightedRandomOrder(channels []repository.GatewayChannel) []repository.GatewayChannel {
	pending := append([]repository.GatewayChannel(nil), channels...)
	ordered := make([]repository.GatewayChannel, 0, len(pending))
	for len(pending) > 0 {
		total := 0
		for _, channel := range pending {
			if channel.Weight > 0 {
				total += channel.Weight
			}
		}
		if total <= 0 {
			sort.SliceStable(pending, func(i, j int) bool { return pending[i].ID < pending[j].ID })
			return append(ordered, pending...)
		}
		pick := secureIntn(total)
		acc := 0
		selected := 0
		for i, channel := range pending {
			weight := channel.Weight
			if weight <= 0 {
				weight = 1
			}
			acc += weight
			if pick < acc {
				selected = i
				break
			}
		}
		ordered = append(ordered, pending[selected])
		pending = append(pending[:selected], pending[selected+1:]...)
	}
	return ordered
}

func secureIntn(max int) int {
	if max <= 1 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return int(time.Now().UnixNano() % int64(max))
	}
	return int(n.Int64())
}

func stickyKey(principal GatewayPrincipal, req ChatRequest, headerSessionID string) string {
	session := strings.TrimSpace(headerSessionID)
	if session == "" {
		session = strings.TrimSpace(req.SessionID)
	}
	if session == "" {
		session = strings.TrimSpace(req.ConversationID)
	}
	if session == "" {
		session = strings.TrimSpace(req.User)
	}
	if session == "" {
		return ""
	}
	return principal.APIKeyID + ":" + req.Model + ":" + session
}

func (s GatewayService) forwardOnce(ctx context.Context, channel repository.GatewayChannel, rawBody []byte) (upstream.ChatResponse, error) {
	upstreamKeyBytes, _ := base64.StdEncoding.DecodeString(channel.APIKeyEncrypted)
	provider := upstream.ProviderForType(channel.ProviderType)
	return provider.ForwardChat(ctx, channel.BaseURL, string(upstreamKeyBytes), channel.TimeoutSeconds, rawBody, channel.UpstreamModelName)
}

func (s GatewayService) forwardEmbeddingsOnce(ctx context.Context, channel repository.GatewayChannel, rawBody []byte) (upstream.ChatResponse, error) {
	upstreamKeyBytes, _ := base64.StdEncoding.DecodeString(channel.APIKeyEncrypted)
	provider := upstream.ProviderForType(channel.ProviderType)
	return provider.ForwardEmbeddings(ctx, channel.BaseURL, string(upstreamKeyBytes), channel.TimeoutSeconds, rawBody, channel.UpstreamModelName)
}

func (s GatewayService) openStreamOnce(ctx context.Context, channel repository.GatewayChannel, rawBody []byte) (*http.Response, error) {
	upstreamKeyBytes, _ := base64.StdEncoding.DecodeString(channel.APIKeyEncrypted)
	provider := upstream.ProviderForType(channel.ProviderType)
	return provider.OpenChatStream(ctx, channel.BaseURL, string(upstreamKeyBytes), channel.TimeoutSeconds, rawBody, channel.UpstreamModelName)
}

func shouldRetryStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == 529 || status >= 500
}

func channelCooldownForStatus(status int) time.Duration {
	switch status {
	case http.StatusTooManyRequests:
		return channelRateLimitCooldown
	case 529:
		return channelOverloadCooldown
	default:
		return channelFailureCooldown
	}
}

func NormalizeUpstreamErrorBody(status int, body []byte) []byte {
	trimmed := strings.TrimSpace(string(body))
	if trimmed != "" && json.Valid([]byte(trimmed)) {
		return []byte(trimmed)
	}
	message := trimmed
	if message == "" {
		message = http.StatusText(status)
	}
	out, err := json.Marshal(map[string]any{
		"error": map[string]any{
			"message":         message,
			"type":            "upstream_error",
			"upstream_status": status,
		},
	})
	if err != nil {
		return []byte(`{"error":{"message":"upstream error","type":"upstream_error"}}`)
	}
	return out
}

func bodyWithMaxTokens(rawBody []byte, maxTokens int) []byte {
	if maxTokens <= 0 {
		return rawBody
	}
	var payload map[string]any
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return rawBody
	}
	if existing, ok := payload["max_tokens"]; ok {
		switch value := existing.(type) {
		case float64:
			if value > 0 {
				return rawBody
			}
		case int:
			if value > 0 {
				return rawBody
			}
		case json.Number:
			if parsed, err := value.Int64(); err == nil && parsed > 0 {
				return rawBody
			}
		}
	}
	payload["max_tokens"] = maxTokens
	out, err := json.Marshal(payload)
	if err != nil {
		return rawBody
	}
	return out
}

func CopyAndCapture(dst http.ResponseWriter, src io.Reader) ([]byte, error) {
	captured, _, err := CopyAndCaptureWithFirstByte(dst, src, time.Now())
	return captured, err
}

func CopyAndCaptureWithFirstByte(dst http.ResponseWriter, src io.Reader, start time.Time) ([]byte, int, error) {
	var captured []byte
	firstByteMS := 0
	buf := make([]byte, 32*1024)
	flusher, _ := dst.(http.Flusher)
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if firstByteMS == 0 {
				firstByteMS = repository.NowMS(start)
			}
			chunk := buf[:n]
			captured = append(captured, chunk...)
			if _, err := dst.Write(chunk); err != nil {
				return captured, firstByteMS, err
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				return captured, firstByteMS, nil
			}
			return captured, firstByteMS, readErr
		}
	}
}
