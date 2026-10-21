package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/stretchr/testify/require"
)

const (
	previewChannelID   = "ch-preview-1"
	previewBotID       = "bot-preview-1"
	previewFallbackCh  = "ch-preview-fb"
	previewBroadcastCh = "ch-preview-b"
	previewProcNormal  = "p-normal"
	previewProcFB      = "p-fb"
	previewProc1       = "p1"
	previewProc2       = "p2"
	previewContent     = "hello"
	previewSender      = "u-preview"
	jsonIDsKey         = "ids"
)

var errPreviewNotFound = errors.New("not found")

type nopRoutingLogger struct{}

func (nopRoutingLogger) Debug(interface{}, ...interface{}) {}
func (nopRoutingLogger) Info(string, ...interface{})       {}
func (nopRoutingLogger) Warn(string, ...interface{})       {}
func (nopRoutingLogger) Error(interface{}, ...interface{}) {}
func (nopRoutingLogger) Fatal(interface{}, ...interface{}) {}

var _ logger.Interface = nopRoutingLogger{}

type memChannelRepo struct {
	ch *entity.Channel
}

func (m *memChannelRepo) Create(context.Context, *entity.Channel) error { return nil }
func (m *memChannelRepo) GetByID(_ context.Context, id string) (*entity.Channel, error) {
	if m.ch != nil && m.ch.ID == id {
		return m.ch, nil
	}

	return nil, errPreviewNotFound
}

func (m *memChannelRepo) GetByBotID(context.Context, string) ([]*entity.Channel, error) {
	return nil, nil
}

func (m *memChannelRepo) GetByPlatformType(context.Context, string) ([]*entity.Channel, error) {
	return nil, nil
}

func (m *memChannelRepo) GetByWebhookPath(context.Context, string) (*entity.Channel, error) {
	return nil, errPreviewNotFound
}
func (m *memChannelRepo) Update(context.Context, *entity.Channel) error { return nil }
func (m *memChannelRepo) Delete(context.Context, string) error          { return nil }
func (m *memChannelRepo) List(context.Context, port.ListParams) (*port.PaginatedResult[*entity.Channel], error) {
	return nil, nil
}
func (m *memChannelRepo) ListActive(context.Context) ([]*entity.Channel, error) { return nil, nil }
func (m *memChannelRepo) UpdateConnectionStatus(context.Context, string, entity.ConnectionStatus) error {
	return nil
}

func (m *memChannelRepo) UpdateAccessToken(context.Context, string, string, *time.Time) error {
	return nil
}
func (m *memChannelRepo) Exists(context.Context, string) (bool, error) { return false, nil }

type memRuleRepo struct {
	rules []*entity.MessageRoutingRule
}

func (m *memRuleRepo) Create(context.Context, *entity.MessageRoutingRule) error { return nil }
func (m *memRuleRepo) GetByID(context.Context, string) (*entity.MessageRoutingRule, error) {
	return nil, nil
}

func (m *memRuleRepo) GetByProcessorID(context.Context, string) ([]*entity.MessageRoutingRule, error) {
	return nil, nil
}

func (m *memRuleRepo) GetActiveRules(context.Context) ([]*entity.MessageRoutingRule, error) {
	return m.rules, nil
}

func (m *memRuleRepo) GetMatchingRules(context.Context, string, string, string, string) ([]*entity.MessageRoutingRule, error) {
	return m.rules, nil
}

func (m *memRuleRepo) GetFallbackRules(context.Context) ([]*entity.MessageRoutingRule, error) {
	return nil, nil
}
func (m *memRuleRepo) Update(context.Context, *entity.MessageRoutingRule) error { return nil }
func (m *memRuleRepo) Delete(context.Context, string) error                     { return nil }
func (m *memRuleRepo) List(context.Context, port.ListParams) (*port.PaginatedResult[*entity.MessageRoutingRule], error) {
	return nil, nil
}

func (m *memRuleRepo) ListByPriority(context.Context) ([]*entity.MessageRoutingRule, error) {
	return m.rules, nil
}
func (m *memRuleRepo) Exists(context.Context, string) (bool, error) { return false, nil }

type memProcessorRepo struct {
	byID map[string]*entity.MessageProcessor
}

func (m *memProcessorRepo) Create(context.Context, *entity.MessageProcessor) error { return nil }
func (m *memProcessorRepo) GetByID(_ context.Context, id string) (*entity.MessageProcessor, error) {
	p, ok := m.byID[id]
	if !ok {
		return nil, errPreviewNotFound
	}

	return p, nil
}

func (m *memProcessorRepo) GetByName(context.Context, string) (*entity.MessageProcessor, error) {
	return nil, nil
}

func (m *memProcessorRepo) GetByType(context.Context, string) ([]*entity.MessageProcessor, error) {
	return nil, nil
}
func (m *memProcessorRepo) Update(context.Context, *entity.MessageProcessor) error { return nil }
func (m *memProcessorRepo) Delete(context.Context, string) error                   { return nil }
func (m *memProcessorRepo) List(context.Context, port.ListParams) (*port.PaginatedResult[*entity.MessageProcessor], error) {
	return nil, nil
}

func (m *memProcessorRepo) ListActive(context.Context) ([]*entity.MessageProcessor, error) {
	return nil, nil
}

func (m *memProcessorRepo) ListByPriority(context.Context) ([]*entity.MessageProcessor, error) {
	return nil, nil
}
func (m *memProcessorRepo) Exists(context.Context, string) (bool, error) { return false, nil }
func (m *memProcessorRepo) ExistsByName(context.Context, string) (bool, error) {
	return false, nil
}

func TestPreviewRoute_Over20RulesMatchesLive(t *testing.T) {
	t.Parallel()

	channel := &entity.Channel{
		BaseEntity:   entity.BaseEntity{ID: previewChannelID},
		BotID:        previewBotID,
		PlatformType: string(entity.PlatformTypeWecom),
		Status:       entity.StatusActive,
	}

	processors := map[string]*entity.MessageProcessor{}
	rules := make([]*entity.MessageRoutingRule, 0, 25)

	for i := 0; i < 25; i++ {
		pid := fmt.Sprintf("proc-%d", i)
		processors[pid] = &entity.MessageProcessor{
			BaseEntity:    entity.BaseEntity{ID: pid},
			ProcessorName: pid,
			ProcessorType: "webhook_forward",
			Status:        entity.StatusActive,
		}

		rule := &entity.MessageRoutingRule{
			BaseEntity:  entity.BaseEntity{ID: fmt.Sprintf("rule-%d", i)},
			RuleName:    fmt.Sprintf("rule-%d", i),
			ProcessorID: pid,
			RouteType:   entity.RouteTypeDirect,
			Priority:    100 - i,
			Status:      entity.StatusActive,
			IsFallback:  false,
		}
		if i == 24 {
			rule.Priority = 1
			rule.ChannelIDs = entity.JSONField{jsonIDsKey: []interface{}{previewChannelID}}
		}

		rules = append(rules, rule)
	}

	uc := NewRoutingUseCase(
		&memRuleRepo{rules: rules},
		&memProcessorRepo{byID: processors},
		&memChannelRepo{ch: channel},
		nopRoutingLogger{},
	)

	msg := &entity.Message{
		ChannelID:   previewChannelID,
		MessageType: entity.MessageTypeText,
		Content:     previewContent,
		SenderID:    previewSender,
	}

	preview, err := uc.PreviewRoute(context.Background(), msg)
	require.NoError(t, err)
	require.False(t, preview.UsedFallback)
	require.Len(t, preview.SelectedProcessors, 1)
	require.Equal(t, "proc-24", preview.SelectedProcessors[0].ID)

	live, err := uc.RouteMessage(context.Background(), msg)
	require.NoError(t, err)
	require.Equal(t, preview.SelectedProcessors[0].ID, live[0].ID)
}

func TestPreviewRoute_FallbackWhenNoNormalMatch(t *testing.T) {
	t.Parallel()

	channel := &entity.Channel{
		BaseEntity:   entity.BaseEntity{ID: previewFallbackCh},
		BotID:        previewBotID,
		PlatformType: string(entity.PlatformTypeWecom),
		Status:       entity.StatusActive,
	}

	processors := map[string]*entity.MessageProcessor{
		previewProcNormal: {BaseEntity: entity.BaseEntity{ID: previewProcNormal}, ProcessorName: "n", Status: entity.StatusActive},
		previewProcFB:     {BaseEntity: entity.BaseEntity{ID: previewProcFB}, ProcessorName: "fb", Status: entity.StatusActive},
	}

	rules := []*entity.MessageRoutingRule{
		{
			BaseEntity:  entity.BaseEntity{ID: "r-normal"},
			RuleName:    "normal",
			ProcessorID: previewProcNormal,
			RouteType:   entity.RouteTypeDirect,
			Priority:    10,
			Status:      entity.StatusActive,
			ChannelIDs:  entity.JSONField{jsonIDsKey: []interface{}{"other-channel"}},
		},
		{
			BaseEntity:  entity.BaseEntity{ID: "r-fb"},
			RuleName:    "fallback",
			ProcessorID: previewProcFB,
			RouteType:   entity.RouteTypeDirect,
			Priority:    99,
			Status:      entity.StatusActive,
			IsFallback:  true,
		},
	}

	uc := NewRoutingUseCase(
		&memRuleRepo{rules: rules},
		&memProcessorRepo{byID: processors},
		&memChannelRepo{ch: channel},
		nopRoutingLogger{},
	)

	preview, err := uc.PreviewRoute(context.Background(), &entity.Message{
		ChannelID:   previewFallbackCh,
		MessageType: entity.MessageTypeText,
		Content:     "x",
	})
	require.NoError(t, err)
	require.True(t, preview.UsedFallback)
	require.Equal(t, previewProcFB, preview.SelectedProcessors[0].ID)
}

func TestPreviewRoute_BroadcastSelectsMultiple(t *testing.T) {
	t.Parallel()

	channel := &entity.Channel{
		BaseEntity:   entity.BaseEntity{ID: previewBroadcastCh},
		BotID:        previewBotID,
		PlatformType: string(entity.PlatformTypeWecom),
		Status:       entity.StatusActive,
	}

	processors := map[string]*entity.MessageProcessor{
		previewProc1: {BaseEntity: entity.BaseEntity{ID: previewProc1}, Status: entity.StatusActive, ProcessorName: previewProc1},
		previewProc2: {BaseEntity: entity.BaseEntity{ID: previewProc2}, Status: entity.StatusActive, ProcessorName: previewProc2},
	}

	rules := []*entity.MessageRoutingRule{
		{
			BaseEntity: entity.BaseEntity{ID: "r1"}, RuleName: "b1", ProcessorID: previewProc1,
			RouteType: entity.RouteTypeBroadcast, Priority: 1, Status: entity.StatusActive,
		},
		{
			BaseEntity: entity.BaseEntity{ID: "r2"}, RuleName: "b2", ProcessorID: previewProc2,
			RouteType: entity.RouteTypeBroadcast, Priority: 2, Status: entity.StatusActive,
		},
	}

	uc := NewRoutingUseCase(
		&memRuleRepo{rules: rules},
		&memProcessorRepo{byID: processors},
		&memChannelRepo{ch: channel},
		nopRoutingLogger{},
	)

	preview, err := uc.PreviewRoute(context.Background(), &entity.Message{
		ChannelID: previewBroadcastCh, MessageType: entity.MessageTypeText, Content: "x",
	})
	require.NoError(t, err)
	require.Len(t, preview.SelectedProcessors, 2)
}
