package v1

import (
	"strings"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
)

// secretConfigKeys are never echoed in channel API responses.
var secretConfigKeys = map[string]struct{}{
	"access_token":       {},
	"secret":             {},
	"app_secret":         {},
	"token":              {},
	"verification_token": {},
	"encoding_aes_key":   {},
	"aes_key":            {},
	"password":           {},
	"api_key":            {},
	"bot_token":          {},
	"client_secret":      {},
	"corp_secret":        {},
	"private_key":        {},
	"refresh_token":      {},
}

// ChannelResponse is the safe HTTP DTO for channel resources.
// AccessToken and config secrets are never included.
type ChannelResponse struct {
	ID                   string                 `json:"id"`
	BotID                string                 `json:"bot_id"`
	PlatformType         string                 `json:"platform_type"`
	ChannelName          string                 `json:"channel_name"`
	WebhookPath          string                 `json:"webhook_path,omitempty"`
	Config               map[string]interface{} `json:"config,omitempty"`
	AccessTokenExpiresAt *time.Time             `json:"access_token_expires_at,omitempty"`
	ConnectionStatus     string                 `json:"connection_status"`
	LastConnectedAt      *time.Time             `json:"last_connected_at,omitempty"`
	Status               string                 `json:"status"`
	CreatedAt            time.Time              `json:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at"`
}

// ChannelListResponse is the safe paginated channel list DTO.
type ChannelListResponse struct {
	Items      []ChannelResponse `json:"items"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// ToChannelResponse maps an entity channel to a secret-safe response DTO.
func ToChannelResponse(ch *entity.Channel) ChannelResponse {
	if ch == nil {
		return ChannelResponse{}
	}
	return ChannelResponse{
		ID:                   ch.ID,
		BotID:                ch.BotID,
		PlatformType:         ch.PlatformType,
		ChannelName:          ch.ChannelName,
		WebhookPath:          ch.WebhookPath,
		Config:               RedactConfigSecrets(ch.Config),
		AccessTokenExpiresAt: ch.AccessTokenExpiresAt,
		ConnectionStatus:     ch.ConnectionStatus.String(),
		LastConnectedAt:      ch.LastConnectedAt,
		Status:               ch.Status.String(),
		CreatedAt:            ch.CreatedAt,
		UpdatedAt:            ch.UpdatedAt,
	}
}

// ToChannelListResponse maps a usecase list result to a secret-safe response.
func ToChannelListResponse(result *usecase.ChannelListResult) ChannelListResponse {
	if result == nil {
		return ChannelListResponse{Items: []ChannelResponse{}}
	}
	items := make([]ChannelResponse, 0, len(result.Items))
	for i := range result.Items {
		ch := result.Items[i]
		items = append(items, ToChannelResponse(&ch))
	}
	return ChannelListResponse{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}
}

// RedactConfigSecrets returns a shallow copy of config without secret keys.
func RedactConfigSecrets(cfg entity.JSONField) map[string]interface{} {
	if cfg == nil {
		return nil
	}
	out := make(map[string]interface{}, len(cfg))
	for k, v := range cfg {
		if isSecretConfigKey(k) {
			continue
		}
		out[k] = v
	}
	return out
}

func isSecretConfigKey(key string) bool {
	_, ok := secretConfigKeys[strings.ToLower(key)]
	return ok
}
