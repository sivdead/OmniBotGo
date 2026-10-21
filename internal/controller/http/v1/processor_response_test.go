package v1

import (
	"testing"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestToProcessorResponse_RedactsAPIKey(t *testing.T) {
	t.Parallel()

	p := &entity.MessageProcessor{
		BaseEntity:    entity.BaseEntity{ID: "p1"},
		ProcessorName: "ai",
		ProcessorType: "ai_chat",
		Config: entity.JSONField{
			"api_key":  "sk-secret",
			"model":    "gpt",
			"endpoint": "https://api.example.com",
		},
		Priority: 10,
		Status:   entity.StatusActive,
	}

	resp := ToProcessorResponse(p)
	require.NotNil(t, resp)
	_, hasKey := resp.Config["api_key"]
	require.False(t, hasKey)
	require.Equal(t, "gpt", resp.Config["model"])
	require.Equal(t, "https://api.example.com", resp.Config["endpoint"])
}
