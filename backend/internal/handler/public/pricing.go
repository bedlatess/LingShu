package public

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"lingshu/backend/internal/billing"
	"lingshu/backend/internal/pkg/httpx"
)

type Handler struct {
	db *pgxpool.Pool
}

type PublicModelDTO struct {
	ID               string `json:"id"`
	PublicName       string `json:"public_name"`
	Type             string `json:"type"`
	Group            string `json:"group,omitempty"`
	BillingMode      string `json:"billing_mode"`
	InputPricePer1M  string `json:"input_price_per_1m"`
	OutputPricePer1M string `json:"output_price_per_1m"`
	PricePerCall     string `json:"price_per_call,omitempty"`
	Currency         string `json:"currency"`
}

func New(db *pgxpool.Pool) Handler {
	return Handler{db: db}
}

func (h Handler) ListModels(w http.ResponseWriter, r *http.Request) {
	items, err := h.listModels(r.Context())
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h Handler) SiteInfo(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settings(r.Context(), "site_name", "registration_enabled", "contact_info")
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"site_name":            firstSetting(settings, "site_name", "LingShu"),
		"registration_enabled": firstSetting(settings, "registration_enabled", "false") == "true",
		"contact_info":         firstSetting(settings, "contact_info", ""),
		"login_url":            "/login",
	})
}

func (h Handler) listModels(ctx context.Context) ([]PublicModelDTO, error) {
	rows, err := h.db.Query(ctx, `
		SELECT id::text, public_name, type, model_group, billing_mode,
		       input_price_per_1k::text, output_price_per_1k::text, price_per_call::text,
		       rate_multiplier::text
		FROM models
		WHERE status='enabled'
		  AND deleted_at IS NULL
		  AND EXISTS (
		  	SELECT 1
		  	FROM channel_models cm
		  	JOIN upstream_channels c ON c.id = cm.channel_id AND c.deleted_at IS NULL
		  	WHERE cm.model_id=models.id AND cm.status='enabled' AND c.status='enabled'
		  )
		ORDER BY sort_order ASC, created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []PublicModelDTO{}
	for rows.Next() {
		var id, publicName, modelType, group, billingMode, inputPer1K, outputPer1K, pricePerCall, multiplier string
		if err := rows.Scan(&id, &publicName, &modelType, &group, &billingMode, &inputPer1K, &outputPer1K, &pricePerCall, &multiplier); err != nil {
			return nil, err
		}
		items = append(items, PublicModelDTO{
			ID:               id,
			PublicName:       publicName,
			Type:             modelType,
			Group:            group,
			BillingMode:      billingMode,
			InputPricePer1M:  publicPricePer1M(inputPer1K, multiplier),
			OutputPricePer1M: publicPricePer1M(outputPer1K, multiplier),
			PricePerCall:     publicPrice(pricePerCall, multiplier),
			Currency:         "USD",
		})
	}
	return items, rows.Err()
}

func (h Handler) settings(ctx context.Context, keys ...string) (map[string]string, error) {
	rows, err := h.db.Query(ctx, `SELECT key, value FROM system_settings WHERE key = ANY($1)`, keys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		out[key] = value
	}
	return out, rows.Err()
}

func publicPricePer1M(per1K string, multiplier string) string {
	units, _ := billing.DecimalStringToUnits(per1K)
	multiplierUnits, _ := billing.DecimalStringToUnits(multiplier)
	return billing.UnitsToDecimalString(units * multiplierUnits * 1000 / billing.Scale)
}

func publicPrice(price string, multiplier string) string {
	units, _ := billing.DecimalStringToUnits(price)
	multiplierUnits, _ := billing.DecimalStringToUnits(multiplier)
	return billing.UnitsToDecimalString(units * multiplierUnits / billing.Scale)
}

func firstSetting(settings map[string]string, key, fallback string) string {
	if value := settings[key]; value != "" {
		return value
	}
	return fallback
}

func encodePublicModelsForTest(items []PublicModelDTO) ([]byte, error) {
	return json.Marshal(map[string]any{"items": items})
}
