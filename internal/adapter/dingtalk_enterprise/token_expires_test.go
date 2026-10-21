package dingtalk_enterprise

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

const (
	testDingtalkAppKey    = "app_key"
	testDingtalkAppSecret = "app_secret"
)

func TestGetAccessToken_CacheHitSetsExpiresAt(t *testing.T) {
	t.Parallel()

	adapter := NewAdapter(zerolog.Nop())
	expiresAt := time.Now().Add(30 * time.Minute)
	cacheKey := "key:secret"
	adapter.tokenCache[cacheKey] = &tokenCacheItem{
		Token:      "cached-token",
		ExpireTime: expiresAt,
	}

	resp, err := adapter.GetAccessToken(context.Background(), map[string]interface{}{
		testDingtalkAppKey:    "key",
		testDingtalkAppSecret: "secret",
	})
	require.NoError(t, err)
	require.Equal(t, "cached-token", resp.AccessToken)
	require.NotNil(t, resp.ExpiresAt)
	require.WithinDuration(t, expiresAt, *resp.ExpiresAt, time.Second)
}
