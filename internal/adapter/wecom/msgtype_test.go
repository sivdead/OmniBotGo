package wecom

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/stretchr/testify/require"
)

const testAgentIDKey = "agent_id"

func TestWecom_SendMessage_RejectsUnknownType(t *testing.T) {
	t.Parallel()

	a := NewWecomAdapter(zerolog.Nop())
	err := a.SendMessage(context.Background(), &dto.UnifiedMessage{
		MessageType: "sticker",
		Content:     "x",
		ReceiverID:  "u1",
	}, map[string]interface{}{testAgentIDKey: "1"}, "token")
	require.True(t, errors.Is(err, port.ErrUnsupportedMessageType))
}

func TestWecom_SupportedMessageTypes_IncludesAudio(t *testing.T) {
	t.Parallel()

	a := NewWecomAdapter(zerolog.Nop())

	var _ port.MessageTypeSupport = a

	require.Contains(t, a.SupportedMessageTypes(), entity.MessageTypeAudio)
}
