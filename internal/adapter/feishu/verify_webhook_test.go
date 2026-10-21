package feishu

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyWebhook_RejectsMissingWrongNonStringToken(t *testing.T) {
	adapter := NewFeishuAdapter()
	cfg := map[string]interface{}{"verification_token": "expected-token"}
	ctx := context.Background()

	t.Run("missing token field", func(t *testing.T) {
		err := adapter.VerifyWebhook(ctx, "", "", "", []byte(`{"challenge":"c"}`), cfg)
		require.ErrorIs(t, err, errFeishuWebhookTokenMismatch)
	})

	t.Run("wrong token", func(t *testing.T) {
		err := adapter.VerifyWebhook(ctx, "", "", "", []byte(`{"token":"nope"}`), cfg)
		require.ErrorIs(t, err, errFeishuWebhookTokenMismatch)
	})

	t.Run("non-string token", func(t *testing.T) {
		err := adapter.VerifyWebhook(ctx, "", "", "", []byte(`{"token":123}`), cfg)
		require.ErrorIs(t, err, errFeishuWebhookTokenMismatch)
	})

	t.Run("matching token", func(t *testing.T) {
		err := adapter.VerifyWebhook(ctx, "", "", "", []byte(`{"token":"expected-token"}`), cfg)
		require.NoError(t, err)
	})

	t.Run("unconfigured verification_token skips check", func(t *testing.T) {
		err := adapter.VerifyWebhook(ctx, "", "", "", []byte(`{}`), map[string]interface{}{})
		require.NoError(t, err)
	})
}
