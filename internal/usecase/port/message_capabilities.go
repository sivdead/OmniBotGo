package port

import (
	"context"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
)

// InboundMessageReceiver normalizes platform messages using explicit channel scope.
type InboundMessageReceiver interface {
	ReceiveInboundMessage(ctx context.Context, channelID string, message *dto.UnifiedMessage) error
}

// MessageReplier resolves the destination from a persisted inbound message.
type MessageReplier interface {
	ReplyToMessage(ctx context.Context, originalMessageID string, reply *dto.UnifiedMessage) (*entity.Message, error)
}

// RoutingMatch describes a matching rule without executing its processor.
type RoutingMatch struct {
	RuleID      string `json:"rule_id"`
	RuleName    string `json:"rule_name"`
	ProcessorID string `json:"processor_id"`
	Priority    int    `json:"priority"`
	IsFallback  bool   `json:"is_fallback"`
	Selected    bool   `json:"selected"`
}

// RoutingPreview explains matching and selection with no delivery side effects.
type RoutingPreview struct {
	Matches              []RoutingMatch `json:"matches"`
	SelectedProcessorIDs []string       `json:"selected_processor_ids"`
	Unrouted             bool           `json:"unrouted"`
}

// RoutingPreviewer exposes the same selection semantics as message processing.
type RoutingPreviewer interface {
	PreviewRouting(ctx context.Context, message *entity.Message) (*RoutingPreview, error)
}
