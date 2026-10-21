package adapter

import (
	"fmt"

	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// EnsureMessageTypeSupported rejects undeclared types when the adapter declares capabilities (T08).
func EnsureMessageTypeSupported(platformAdapter interface{}, msgType string) error {
	caps, ok := platformAdapter.(port.MessageTypeSupport)
	if !ok {
		return nil
	}

	for _, supported := range caps.SupportedMessageTypes() {
		if supported == msgType {
			return nil
		}
	}

	return fmt.Errorf("%w: %q", port.ErrUnsupportedMessageType, msgType)
}
