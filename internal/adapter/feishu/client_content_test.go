package feishu

import (
	"encoding/json"
	"testing"

	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/require"
)

const (
	testFeishuAppID     = "app-a"
	testFeishuAppIDKey  = "app_id"
	testFeishuAppSecret = "app_secret"
	testFeishuSecretA   = "sec-a"
	testFeishuSecretB   = "sec-b"
)

func TestBuildMessageContent_EscapesQuotes(t *testing.T) {
	t.Parallel()

	f := NewFeishuAdapter()
	raw, err := f.buildMessageContent(&dto.UnifiedMessage{
		MessageType: entity.MessageTypeText,
		Content:     `say "hello"\nworld`,
	})
	require.NoError(t, err)

	var parsed map[string]string

	require.NoError(t, json.Unmarshal([]byte(raw), &parsed))
	require.Equal(t, `say "hello"\nworld`, parsed[feishuContentKeyText])
}

func TestGetClient_KeysByAppCredentials(t *testing.T) {
	t.Parallel()

	f := NewFeishuAdapter()

	c1, err := f.getClient(map[string]interface{}{
		testFeishuAppIDKey:  testFeishuAppID,
		testFeishuAppSecret: testFeishuSecretA,
	})
	require.NoError(t, err)

	c1b, err := f.getClient(map[string]interface{}{
		testFeishuAppIDKey:  testFeishuAppID,
		testFeishuAppSecret: testFeishuSecretA,
	})
	require.NoError(t, err)
	require.Same(t, c1, c1b)

	c2, err := f.getClient(map[string]interface{}{
		testFeishuAppIDKey:  testFeishuAppID,
		testFeishuAppSecret: testFeishuSecretB,
	})
	require.NoError(t, err)
	require.NotSame(t, c1, c2)
	require.Len(t, f.clients, 2)
}
