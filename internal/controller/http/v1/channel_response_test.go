package v1

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/stretchr/testify/require"
)

func TestToChannelResponse_OmitsAccessTokenAndSecrets(t *testing.T) {
	t.Parallel()

	expires := time.Now().UTC().Add(time.Hour)
	ch := &entity.Channel{
		BaseEntity:   entity.BaseEntity{ID: "019a0000-0000-7000-8000-000000000001"},
		BotID:        "bot-1",
		PlatformType: "feishu",
		ChannelName:  "main",
		WebhookPath:  "/webhook/feishu/019a0000-0000-7000-8000-000000000001",
		Config: entity.JSONField{
			"app_id":                   "cli_xxx",
			configKeyAppSecret:         "super-secret",
			configKeyVerificationToken: "vt-secret",
			configKeyToken:             "tok",
		},
		AccessToken:          "raw-access-token",
		AccessTokenExpiresAt: &expires,
		ConnectionStatus:     entity.ConnectionStatusDisconnected,
		Status:               entity.StatusActive,
	}

	resp := ToChannelResponse(ch)
	raw, err := json.Marshal(resp)
	require.NoError(t, err)

	var asMap map[string]interface{}

	require.NoError(t, json.Unmarshal(raw, &asMap))

	_, hasToken := asMap[configKeyAccessToken]
	require.False(t, hasToken, "access_token must not appear in JSON")

	cfg, ok := asMap["config"].(map[string]interface{})
	require.True(t, ok)
	require.Equal(t, "cli_xxx", cfg["app_id"])

	_, hasSecret := cfg[configKeyAppSecret]
	require.False(t, hasSecret)

	_, hasVT := cfg[configKeyVerificationToken]
	require.False(t, hasVT)

	_, hasTok := cfg[configKeyToken]
	require.False(t, hasTok)

	require.Equal(t, ch.WebhookPath, resp.WebhookPath)
	require.Equal(t, "raw-access-token", ch.AccessToken, "entity itself must remain unchanged")
}

func TestToChannelListResponse_RedactsAllItems(t *testing.T) {
	t.Parallel()

	result := &usecase.ChannelListResult{
		Items: []entity.Channel{{
			BaseEntity:   entity.BaseEntity{ID: "id-1"},
			PlatformType: "wecom",
			ChannelName:  "c1",
			AccessToken:  "secret-token",
			Config:       entity.JSONField{configKeySecret: "s", "corp_id": "ww"},
			Status:       entity.StatusActive,
		}},
		Total:      1,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}

	out := ToChannelListResponse(result)
	require.Len(t, out.Items, 1)

	raw, err := json.Marshal(out.Items[0])
	require.NoError(t, err)
	require.NotContains(t, string(raw), "secret-token")
	require.NotContains(t, string(raw), `"`+configKeySecret+`"`)
	require.Contains(t, string(raw), "corp_id")
}
