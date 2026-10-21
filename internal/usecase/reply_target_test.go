package usecase

import (
	"testing"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/require"
)

const (
	testReplyChannelID = "ch-reply-1"
	testReplyBotID     = "bot-reply-1"
	testReplyHello     = "hello"
	testReplyGroupChat = "chat-group-7"
	testReplyUserID    = "user-reply-1"
)

func TestBuildOutboundReplyFromInbound_DM(t *testing.T) {
	t.Parallel()

	inbound := &entity.Message{
		BaseEntity:     entity.BaseEntity{ID: "in-1"},
		MessageID:      "msg-dm-1",
		ChannelID:      testReplyChannelID,
		Direction:      entity.MessageDirectionInbound,
		SenderID:       testReplyUserID,
		SenderName:     "Alice",
		SenderType:     entity.SenderTypeUser,
		ReceiverID:     testReplyBotID,
		ReceiverName:   "Bot",
		ReceiverType:   entity.ReceiverTypeBot,
		ConversationID: "conv-dm",
		Content:        testReplyHello,
	}

	reply := BuildOutboundReplyFromInbound(inbound, "hi back", entity.MessageTypeText)

	require.Equal(t, testReplyChannelID, reply.ChannelID)
	require.Equal(t, entity.MessageDirectionOutbound, reply.Direction)
	require.Equal(t, testReplyUserID, reply.ReceiverID)
	require.Equal(t, entity.ReceiverTypeUser, reply.ReceiverType)
	require.Equal(t, "conv-dm", reply.ConversationID)
	require.NotNil(t, reply.ParentMessageID)
	require.Equal(t, "in-1", *reply.ParentMessageID)
	require.Equal(t, "hi back", reply.Content)
	require.NotEqual(t, testReplyUserID, reply.SenderID)
}

func TestBuildOutboundReplyFromInbound_Group(t *testing.T) {
	t.Parallel()

	inbound := &entity.Message{
		BaseEntity:     entity.BaseEntity{ID: "in-2"},
		MessageID:      "msg-group-1",
		ChannelID:      "ch-reply-2",
		Direction:      entity.MessageDirectionInbound,
		SenderID:       "user-99",
		SenderName:     "Bob",
		SenderType:     entity.SenderTypeUser,
		ReceiverID:     testReplyGroupChat,
		ReceiverName:   "Dev Group",
		ReceiverType:   entity.ReceiverTypeGroup,
		ConversationID: testReplyGroupChat,
		Content:        "ping",
		RawContent: entity.JSONField{
			rawKeySessionWebhook:   "https://example.com/session",
			rawKeyConversationType: dingTalkGroupConvType,
		},
	}

	reply := BuildOutboundReplyFromInbound(inbound, "pong", "")

	require.Equal(t, testReplyGroupChat, reply.ReceiverID)
	require.Equal(t, entity.ReceiverTypeGroup, reply.ReceiverType)
	require.Equal(t, testReplyGroupChat, reply.ConversationID)
	require.Equal(t, entity.MessageTypeText, reply.MessageType)
	require.Equal(t, "https://example.com/session", reply.RawContent[rawKeySessionWebhook])
	require.NotEqual(t, "user-99", reply.ReceiverID)
}

func TestBuildOutboundReplyFromInbound_GroupViaConversationType(t *testing.T) {
	t.Parallel()

	inbound := &entity.Message{
		BaseEntity:     entity.BaseEntity{ID: "in-3"},
		ChannelID:      "ch-reply-3",
		SenderID:       testReplyUserID,
		ReceiverID:     replySenderBotID,
		ReceiverType:   entity.ReceiverTypeBot,
		ConversationID: "cid-9",
		RawContent: entity.JSONField{
			rawKeyConversationType: float64(2),
			rawKeySessionWebhook:   "https://hook",
		},
	}

	reply := BuildOutboundReplyFromInbound(inbound, "ok", entity.MessageTypeText)
	require.Equal(t, entity.ReceiverTypeGroup, reply.ReceiverType)
	require.Equal(t, "cid-9", reply.ReceiverID)
	require.Equal(t, "https://hook", reply.RawContent[rawKeySessionWebhook])
}
