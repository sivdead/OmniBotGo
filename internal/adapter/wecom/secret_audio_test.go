package wecom

import (
	"testing"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/require"
)

const (
	testWecomSecretKey    = "secret"
	testWecomAppSecretKey = "app_secret"
)

func TestWecomCorpSecret_PrefersSecret(t *testing.T) {
	t.Parallel()

	secret, err := wecomCorpSecret(map[string]interface{}{
		testWecomSecretKey:    "from-secret",
		testWecomAppSecretKey: "from-app-secret",
	})
	require.NoError(t, err)
	require.Equal(t, "from-secret", secret)
}

func TestWecomCorpSecret_AppSecretAlias(t *testing.T) {
	t.Parallel()

	secret, err := wecomCorpSecret(map[string]interface{}{
		testWecomAppSecretKey: "alias-only",
	})
	require.NoError(t, err)
	require.Equal(t, "alias-only", secret)
}

func TestWecomCorpSecret_Missing(t *testing.T) {
	t.Parallel()

	_, err := wecomCorpSecret(map[string]interface{}{})
	require.ErrorIs(t, err, errWecomMissingSecret)
}

func TestMapToWecomMessageType_AudioIsVoice(t *testing.T) {
	t.Parallel()

	require.Equal(t, "voice", mapToWecomMessageType(entity.MessageTypeAudio))
}

func TestSendMessage_AudioUsesVoiceMediaID(t *testing.T) {
	t.Parallel()

	msg := &dto.UnifiedMessage{
		MessageType: entity.MessageTypeAudio,
		ReceiverID:  "user1",
		MediaURL:    "media-audio-1",
	}

	sendMsg := WecomSendMessage{
		ToUser:  msg.ReceiverID,
		MsgType: mapToWecomMessageType(msg.MessageType),
		AgentID: "1000001",
	}
	require.Equal(t, "voice", sendMsg.MsgType)

	sendMsg.Voice = &WecomVoiceMessage{MediaID: msg.MediaURL}
	require.NotNil(t, sendMsg.Voice)
	require.Equal(t, "media-audio-1", sendMsg.Voice.MediaID)
	require.Nil(t, sendMsg.Text)
}
