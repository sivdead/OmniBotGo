package v1

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/sivdead/OmniBotGo/internal/adapter"
	"github.com/sivdead/OmniBotGo/internal/adapter/feishu"
	"github.com/sivdead/OmniBotGo/internal/controller/http/middleware"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/stretchr/testify/require"
)

const (
	testChannelNameFeishuMain = "feishu-main"
	testAPIKey                = "test-api-key"
	testVerificationTokenVal  = "vt"
)

var errNotFound = errors.New("通道不存在")

type stubChannelUC struct {
	channel *entity.Channel
}

func (s *stubChannelUC) CreateChannel(context.Context, usecase.CreateChannelRequest) (*entity.Channel, error) {
	return nil, nil
}

func (s *stubChannelUC) UpdateChannel(context.Context, usecase.UpdateChannelRequest) (*entity.Channel, error) {
	return nil, nil
}

func (s *stubChannelUC) DeleteChannel(context.Context, string) error { return nil }

func (s *stubChannelUC) GetChannel(_ context.Context, id string) (*entity.Channel, error) {
	if s.channel != nil && s.channel.ID == id {
		return s.channel, nil
	}

	return nil, errNotFound
}

func (s *stubChannelUC) ListChannels(context.Context, usecase.ListChannelsParams) (*usecase.ChannelListResult, error) {
	return &usecase.ChannelListResult{}, nil
}

func (s *stubChannelUC) UpdateChannelStatus(context.Context, string, entity.ConnectionStatus) error {
	return nil
}

func (s *stubChannelUC) RefreshChannelToken(context.Context, string) error { return nil }

func (s *stubChannelUC) GetActiveChannels(context.Context) ([]*entity.Channel, error) {
	return nil, nil
}

func (s *stubChannelUC) IsChannelConnected(context.Context, string) bool { return false }

type stubMessageUC struct{}

func (s *stubMessageUC) ProcessInboundMessage(context.Context, *entity.Message) error { return nil }
func (s *stubMessageUC) SendMessage(context.Context, *entity.Message) error           { return nil }
func (s *stubMessageUC) ReplyToOriginal(context.Context, string, usecase.ReplyToOriginalRequest) (*entity.Message, error) {
	return nil, nil
}

func (s *stubMessageUC) GetMessageHistory(context.Context, usecase.GetMessageHistoryParams) (*usecase.MessageHistoryResult, error) {
	return nil, nil
}

func (s *stubMessageUC) GetMessage(context.Context, string) (*entity.Message, error) { return nil, nil }
func (s *stubMessageUC) RetryFailedMessage(context.Context, string) error            { return nil }
func (s *stubMessageUC) CreateStreamMessageHandler() port.MessageHandler             { return nil }

func newWebhookTestApp(t *testing.T, ch *entity.Channel) *fiber.App {
	t.Helper()

	mgr := adapter.NewManagerWithRegistry(map[entity.PlatformType]interface{}{
		entity.PlatformTypeFeishu: feishu.NewFeishuAdapter(),
	})
	v1ctrl := &V1{
		channelUC: &stubChannelUC{channel: ch},
		messageUC: &stubMessageUC{},
		l:         logger.New("error"),
	}
	wc := NewWebhookController(v1ctrl, mgr)

	app := fiber.New()
	app.Get("/webhook/:platform/:channel_id", wc.HandleWebhookGET)
	app.Post("/webhook/:platform/:channel_id", wc.HandleWebhook)

	return app
}

func TestBatch1_AdminAuthPattern(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	api := app.Group("/api/v1", middleware.RequireAPIKey(testAPIKey))
	api.Get("/channels", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	app.Post("/webhook/:platform/:channel_id", func(c *fiber.Ctx) error {
		return c.SendStatus(http.StatusOK)
	})
	app.Get("/live", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	t.Run("unauth admin rejected", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/channels", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		defer resp.Body.Close()
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid key allowed", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/channels", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+testAPIKey)
		resp, err := app.Test(req)
		require.NoError(t, err)

		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("webhook not requiring admin key", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(
			context.Background(),
			http.MethodPost,
			"/webhook/feishu/019a0000-0000-7000-8000-000000000001",
			http.NoBody,
		)
		resp, err := app.Test(req)
		require.NoError(t, err)

		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("live public", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/live", http.NoBody)
		resp, err := app.Test(req)
		require.NoError(t, err)

		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestBatch1_EmptyManagerGone_RegisteredPlatformReachesAdapter(t *testing.T) {
	t.Parallel()

	channelID := "019a0000-0000-7000-8000-0000000000aa"
	ch := &entity.Channel{
		BaseEntity:   entity.BaseEntity{ID: channelID},
		PlatformType: string(entity.PlatformTypeFeishu),
		ChannelName:  testChannelNameFeishuMain,
		WebhookPath:  "/webhook/feishu/" + channelID,
		Config:       entity.JSONField{configKeyVerificationToken: testVerificationTokenVal},
		Status:       entity.StatusActive,
	}
	app := newWebhookTestApp(t, ch)

	// Empty Manager would return "不支持的平台类型". With DI registry, Feishu verify runs.
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/webhook/feishu/"+channelID,
		strings.NewReader(`{"type":"event"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotContains(t, string(raw), "不支持的平台类型")
	require.Contains(t, string(raw), "签名验证失败")
}

func TestBatch1_GetWebhookInfoUUIDNoParseInt(t *testing.T) {
	t.Parallel()

	channelID := "019a0000-0000-7000-8000-0000000000bb"
	path := "/webhook/feishu/" + channelID
	ch := &entity.Channel{
		BaseEntity:       entity.BaseEntity{ID: channelID},
		PlatformType:     string(entity.PlatformTypeFeishu),
		ChannelName:      testChannelNameFeishuMain,
		WebhookPath:      path,
		Status:           entity.StatusActive,
		ConnectionStatus: entity.ConnectionStatusDisconnected,
	}
	app := newWebhookTestApp(t, ch)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, path, http.NoBody)
	resp, err := app.Test(req)
	require.NoError(t, err)

	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var envelope struct {
		Data WebhookInfo `json:"data"`
	}

	require.NoError(t, json.Unmarshal(raw, &envelope))
	require.Equal(t, channelID, envelope.Data.ChannelID)
	require.Equal(t, path, envelope.Data.WebhookPath)
}

func TestBatch1_WebhookPathHitsFiberRoute(t *testing.T) {
	t.Parallel()

	channelID := "019a0000-0000-7000-8000-0000000000cc"
	path := "/webhook/feishu/" + channelID
	ch := &entity.Channel{
		BaseEntity:   entity.BaseEntity{ID: channelID},
		PlatformType: string(entity.PlatformTypeFeishu),
		ChannelName:  testChannelNameFeishuMain,
		WebhookPath:  path,
		Config:       entity.JSONField{configKeyVerificationToken: testVerificationTokenVal},
		Status:       entity.StatusActive,
	}
	app := newWebhookTestApp(t, ch)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		path,
		strings.NewReader(`{"token":"wrong"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)

	defer resp.Body.Close()
	require.NotEqual(t, http.StatusNotFound, resp.StatusCode)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
