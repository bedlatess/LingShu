package public

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	settings, err := h.settings(r.Context(),
		"site_name",
		"registration_enabled",
		"registration_mode",
		"contact_info",
		"contact_email",
		"site_logo_url",
		"site_icp",
		"site_police_beian",
		"tos_url",
		"privacy_url",
		"brand_primary_color",
		"captcha_enabled",
		"captcha_provider",
		"captcha_site_key",
	)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	registrationMode := firstSetting(settings, "registration_mode", "")
	registrationEnabled := registrationMode == "open"
	if registrationMode == "" {
		registrationEnabled = firstSetting(settings, "registration_enabled", "false") == "true"
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"site_name":            firstSetting(settings, "site_name", "LingShu"),
		"registration_enabled": registrationEnabled,
		"registration_mode":    firstNonEmpty(registrationMode, mapBool(registrationEnabled, "open", "closed")),
		"contact_info":         firstSetting(settings, "contact_info", ""),
		"contact_email":        firstSetting(settings, "contact_email", ""),
		"site_logo_url":        firstSetting(settings, "site_logo_url", ""),
		"site_icp":             firstSetting(settings, "site_icp", ""),
		"site_police_beian":    firstSetting(settings, "site_police_beian", ""),
		"tos_url":              firstSetting(settings, "tos_url", "/legal/tos"),
		"privacy_url":          firstSetting(settings, "privacy_url", "/legal/privacy"),
		"brand_primary_color":  firstSetting(settings, "brand_primary_color", ""),
		"captcha_enabled":      firstSetting(settings, "captcha_enabled", "false") == "true",
		"captcha_provider":     firstSetting(settings, "captcha_provider", ""),
		"captcha_site_key":     firstSetting(settings, "captcha_site_key", ""),
		"login_url":            "/login",
	})
}

func (h Handler) Legal(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	key := ""
	switch slug {
	case "tos":
		key = "legal_tos_markdown"
	case "privacy":
		key = "legal_privacy_markdown"
	default:
		httpx.Error(w, http.StatusNotFound, "legal page not found")
		return
	}
	settings, err := h.settings(r.Context(), key)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{
		"slug":     slug,
		"markdown": firstSetting(settings, key, ""),
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func mapBool(value bool, yes, no string) string {
	if value {
		return yes
	}
	return no
}

func encodePublicModelsForTest(items []PublicModelDTO) ([]byte, error) {
	return json.Marshal(map[string]any{"items": items})
}
