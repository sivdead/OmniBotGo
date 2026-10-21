package dto

import (
	"testing"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/require"
)

func TestBuildUnifiedContent_UsesTypedFieldsNotRaw(t *testing.T) {
	t.Parallel()

	msg := &UnifiedMessage{
		RawContent: map[string]interface{}{
			"original_message": map[string]interface{}{"foo": "bar"},
		},
		LocationContent: &entity.LocationMessage{
			Latitude: 1, Longitude: 2, Address: "here",
		},
		EventContent: &entity.EventMessage{
			EventType: entity.EventTypeUserSubscribe,
			EventKey:  "k",
		},
	}

	out := msg.BuildUnifiedContent()
	require.Contains(t, out, "location")
	require.Contains(t, out, "event")
	_, hasRaw := out["original_message"]
	require.False(t, hasRaw)
}
