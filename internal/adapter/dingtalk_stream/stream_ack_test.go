package dingtalk_stream

import (
	"context"
	"errors"
	"testing"

	"github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/stretchr/testify/require"
)

var errProcessBoom = errors.New("process boom")

func TestHandleChatBotMessage_ReturnsErrorOnProcessFailure(t *testing.T) {
	t.Parallel()

	ctrl := NewStreamController(
		func(context.Context, *dto.UnifiedMessage) error { return errProcessBoom },
		zerolog.Nop(),
		&config.DingtalkStreamConfig{},
		"ch-1",
	)

	data := &chatbot.BotCallbackDataModel{
		MsgId:            "mid-1",
		Msgtype:          string(DingtalkMessageTypeText),
		SenderId:         "u1",
		SenderNick:       "n",
		ConversationId:   "c1",
		ConversationType: "1",
		ChatbotUserId:    receiverTypeBot,
		CreateAt:         1710000000000,
		Text:             chatbot.BotCallbackDataTextModel{Content: "hi"},
	}

	body, err := ctrl.handleChatBotMessage(context.Background(), data)
	require.Error(t, err)
	require.Nil(t, body)
	require.ErrorIs(t, err, errProcessBoom)
}
