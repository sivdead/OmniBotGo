package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/stretchr/testify/require"
)

const (
	testValidateBotID  = "bot-1"
	testValidatePlat   = "feishu"
	testValidateChName = "main"
)

var errBadPlatformConfig = errors.New("bad platform config")

type validatingAdapterManager struct {
	nopAdapterManager
	called bool
	err    error
}

func (v *validatingAdapterManager) ValidateConfig(entity.PlatformType, map[string]interface{}) error {
	v.called = true

	return v.err
}

func TestCreateChannel_ValidatesConfig(t *testing.T) {
	t.Parallel()

	mgr := &validatingAdapterManager{err: errBadPlatformConfig}
	uc := NewChannelUseCase(&memoryChannelRepo{}, &memoryBotRepo{}, mgr, logger.New("error"))

	_, err := uc.CreateChannel(context.Background(), CreateChannelRequest{
		BotID:        testValidateBotID,
		PlatformType: testValidatePlat,
		ChannelName:  testValidateChName,
		Config:       map[string]interface{}{testAppID: "x"},
	})
	require.ErrorIs(t, err, errBadPlatformConfig)
	require.True(t, mgr.called)
}

func TestCreateChannel_CallsValidateEvenWhenConfigNil(t *testing.T) {
	t.Parallel()

	mgr := &validatingAdapterManager{}
	uc := NewChannelUseCase(&memoryChannelRepo{}, &memoryBotRepo{}, mgr, logger.New("error"))

	ch, err := uc.CreateChannel(context.Background(), CreateChannelRequest{
		BotID:        testValidateBotID,
		PlatformType: testValidatePlat,
		ChannelName:  testValidateChName,
	})
	require.NoError(t, err)
	require.NotNil(t, ch)
	require.True(t, mgr.called)
}
