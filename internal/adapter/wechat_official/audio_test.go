package wechat_official

import (
	"testing"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestMapToWechatOfficialMessageType_AudioIsVoice(t *testing.T) {
	t.Parallel()

	require.Equal(t, "voice", mapToWechatOfficialMessageType(entity.MessageTypeAudio))
	require.Equal(t, "voice", mapToWechatOfficialMessageType("audio"))
	require.Equal(t, "voice", mapToWechatOfficialMessageType("voice"))
}
