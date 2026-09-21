package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfigReadsDBDSNFromEnv(t *testing.T) {
	t.Setenv("DB_DSN", "user:pass@tcp(db:3306)/omnibotgo?charset=utf8mb4&parseTime=True&loc=Local")
	t.Setenv("APP_NAME", "OmniBotGo")
	t.Setenv("APP_VERSION", "1.0.0")
	t.Setenv("HTTP_PORT", "8080")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("GRPC_PORT", "8081")
	t.Setenv("RMQ_RPC_SERVER", "rpc_server")
	t.Setenv("RMQ_RPC_CLIENT", "rpc_client")

	// Ensure missing config file path does not hide env (cwd may still have config.yaml)
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("chdir cleanup: %v", err)
		}
	})
	tmp := t.TempDir()
	require.NoError(t, os.Chdir(tmp))

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.Equal(t, "user:pass@tcp(db:3306)/omnibotgo?charset=utf8mb4&parseTime=True&loc=Local", cfg.DB.DSN)
	require.Equal(t, "rpc_server", cfg.RMQ.ServerExchange)
	require.Equal(t, "rpc_client", cfg.RMQ.ClientExchange)
}
