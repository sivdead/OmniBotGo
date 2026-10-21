package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/stretchr/testify/require"
)

type memoryChannelRepo struct {
	byID map[string]*entity.Channel
}

func (m *memoryChannelRepo) Create(_ context.Context, channel *entity.Channel) error {
	if m.byID == nil {
		m.byID = map[string]*entity.Channel{}
	}
	cp := *channel
	m.byID[channel.ID] = &cp
	return nil
}
func (m *memoryChannelRepo) GetByID(_ context.Context, id string) (*entity.Channel, error) {
	ch, ok := m.byID[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	cp := *ch
	return &cp, nil
}
func (m *memoryChannelRepo) GetByBotID(context.Context, string) ([]*entity.Channel, error) {
	return nil, nil
}
func (m *memoryChannelRepo) GetByPlatformType(context.Context, string) ([]*entity.Channel, error) {
	return nil, nil
}
func (m *memoryChannelRepo) GetByWebhookPath(context.Context, string) (*entity.Channel, error) {
	return nil, fmt.Errorf("not found")
}
func (m *memoryChannelRepo) Update(_ context.Context, channel *entity.Channel) error {
	cp := *channel
	m.byID[channel.ID] = &cp
	return nil
}
func (m *memoryChannelRepo) Delete(context.Context, string) error { return nil }
func (m *memoryChannelRepo) List(context.Context, port.ListParams) (*port.PaginatedResult[*entity.Channel], error) {
	return &port.PaginatedResult[*entity.Channel]{}, nil
}
func (m *memoryChannelRepo) ListActive(context.Context) ([]*entity.Channel, error) { return nil, nil }
func (m *memoryChannelRepo) UpdateConnectionStatus(context.Context, string, entity.ConnectionStatus) error {
	return nil
}
func (m *memoryChannelRepo) UpdateAccessToken(context.Context, string, string, *time.Time) error {
	return nil
}
func (m *memoryChannelRepo) Exists(context.Context, string) (bool, error) { return false, nil }

type memoryBotRepo struct{}

func (m *memoryBotRepo) Create(context.Context, *entity.Bot) error { return nil }
func (m *memoryBotRepo) GetByID(_ context.Context, id string) (*entity.Bot, error) {
	return &entity.Bot{
		BaseEntity: entity.BaseEntity{ID: id},
		BotName:    "bot",
		Status:     entity.StatusActive,
	}, nil
}
func (m *memoryBotRepo) GetByName(context.Context, string) (*entity.Bot, error) {
	return nil, fmt.Errorf("not found")
}
func (m *memoryBotRepo) Update(context.Context, *entity.Bot) error { return nil }
func (m *memoryBotRepo) Delete(context.Context, string) error      { return nil }
func (m *memoryBotRepo) List(context.Context, port.ListParams) (*port.PaginatedResult[*entity.Bot], error) {
	return &port.PaginatedResult[*entity.Bot]{}, nil
}
func (m *memoryBotRepo) ListActive(context.Context) ([]*entity.Bot, error) { return nil, nil }
func (m *memoryBotRepo) Exists(context.Context, string) (bool, error)       { return true, nil }
func (m *memoryBotRepo) ExistsByName(context.Context, string) (bool, error) {
	return false, nil
}
func (m *memoryBotRepo) GetActiveChannelCount(context.Context, string) (int64, error) {
	return 0, nil
}

type nopAdapterManager struct{}

func (nopAdapterManager) GetAdapter(entity.PlatformType) (interface{}, error) {
	return nil, fmt.Errorf("nop")
}
func (nopAdapterManager) GetMessageSender(entity.PlatformType) (port.MessageSender, error) {
	return nil, fmt.Errorf("nop")
}
func (nopAdapterManager) GetWebhookProcessor(entity.PlatformType) (port.WebhookProcessor, error) {
	return nil, fmt.Errorf("nop")
}
func (nopAdapterManager) GetTokenManager(entity.PlatformType) (port.TokenManager, error) {
	return nil, fmt.Errorf("nop")
}
func (nopAdapterManager) GetStreamAdapter(entity.PlatformType) (port.StreamAdapter, error) {
	return nil, fmt.Errorf("nop")
}
func (nopAdapterManager) GetConfigValidator(entity.PlatformType) (port.ConfigValidator, error) {
	return nil, fmt.Errorf("nop")
}
func (nopAdapterManager) GetAIModelManager() port.AIModelManager { return nil }
func (nopAdapterManager) GetSupportedPlatforms() []entity.PlatformType {
	return nil
}
func (nopAdapterManager) ValidateConfig(entity.PlatformType, map[string]interface{}) error {
	return nil
}
func (nopAdapterManager) SendMessage(context.Context, entity.PlatformType, *dto.UnifiedMessage, map[string]interface{}, string) error {
	return nil
}

func TestCreateChannel_WebhookPathMatchesRegisteredRoute(t *testing.T) {
	repo := &memoryChannelRepo{}
	uc := NewChannelUseCase(repo, &memoryBotRepo{}, nopAdapterManager{}, logger.New("error"))

	ch, err := uc.CreateChannel(context.Background(), CreateChannelRequest{
		BotID:        "bot-1",
		PlatformType: "feishu",
		ChannelName:  "main",
		Config:       map[string]interface{}{"app_id": "x"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, ch.ID)
	require.Equal(t, "/webhook/feishu/"+ch.ID, ch.WebhookPath)

	stored, err := repo.GetByID(context.Background(), ch.ID)
	require.NoError(t, err)
	require.Equal(t, ch.WebhookPath, stored.WebhookPath)
	require.Equal(t, "https://example.com"+ch.WebhookPath, stored.GetWebhookURL("https://example.com"))
}
