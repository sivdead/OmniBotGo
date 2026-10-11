package port

import (
	"context"

	"github.com/sivdead/OmniBotGo/internal/entity"
)

// ChannelRuntime groups capabilities owned by one configured channel.
type ChannelRuntime struct {
	Sender  MessageSender
	Webhook WebhookProcessor
	Token   TokenManager
	Stream  StreamAdapter
}

// ChannelRuntimeProvider resolves isolated capabilities without changing legacy ports.
type ChannelRuntimeProvider interface {
	ResolveChannelRuntime(ctx context.Context, channel *entity.Channel) (*ChannelRuntime, error)
	ReleaseChannelRuntime(ctx context.Context, channelID string) error
}
