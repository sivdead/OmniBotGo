package dto

import (
	"testing"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/require"
)

const testReceiverTypeBot = "bot"

func TestUnifiedMessage_ToInboundEntity(t *testing.T) {
	t.Parallel()

	u := &UnifiedMessage{
		MessageID:         "m1",
		MessageType:       entity.MessageTypeText,
		SenderID:          "u1",
		SenderName:        "Alice",
		SenderType:        "user",
		ReceiverID:        testReceiverTypeBot,
		ReceiverType:      testReceiverTypeBot,
		Content:           "hi",
		MediaType:         "text/plain",
		ConversationID:    "conv-1",
		PlatformMessageID: "p1",
		PlatformTimestamp: time.Unix(1710000000, 0).UTC(),
		RawContent:        map[string]interface{}{"k": "v"},
	}

	msg := u.ToInboundEntity("ch-42")
	require.NotNil(t, msg)
	require.Equal(t, "ch-42", msg.ChannelID)
	require.Equal(t, "m1", msg.MessageID)
	require.Equal(t, entity.MessageDirectionInbound, msg.Direction)
	require.Equal(t, entity.MessageStatusPending, msg.MessageStatus)
	require.Equal(t, "text/plain", msg.ContentType)
	require.Equal(t, "conv-1", msg.ConversationID)
	require.Equal(t, "ch-42", msg.RawContent["channel_id"])
	require.Equal(t, "v", msg.RawContent["k"])
	require.False(t, msg.ReceivedAt.IsZero())
}

func TestUnifiedMessage_ToInboundEntity_Nil(t *testing.T) {
	t.Parallel()

	var u *UnifiedMessage

	require.Nil(t, u.ToInboundEntity("ch"))
}
