package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/stretchr/testify/require"
)

func TestDeriveMessagePipelineStage(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cases := []struct {
		name string
		msg  *entity.Message
		want MessagePipelineStage
	}{
		{"pending receive", &entity.Message{MessageStatus: entity.MessageStatusPending}, MessageStageReceive},
		{"processing", &entity.Message{MessageStatus: entity.MessageStatusProcessing}, MessageStageProcess},
		{"processed", &entity.Message{MessageStatus: entity.MessageStatusProcessed}, MessageStageProcess},
		{"sent deliver", &entity.Message{MessageStatus: entity.MessageStatusSent}, MessageStageDeliver},
		{"failed outbound", &entity.Message{MessageStatus: entity.MessageStatusFailed, Direction: entity.MessageDirectionOutbound}, MessageStageDeliver},
		{"failed inbound", &entity.Message{MessageStatus: entity.MessageStatusFailed, Direction: entity.MessageDirectionInbound, ProcessedAt: &now}, MessageStageProcess},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, deriveMessagePipelineStage(tc.msg))
		})
	}
}

func TestTroubleshootMessage_RecoverableFailedInbound(t *testing.T) {
	t.Parallel()

	repo := &msgRepoStub{
		byMessageID: map[string]*entity.Message{
			"biz-1": {
				BaseEntity:    entity.BaseEntity{ID: "pk-1"},
				MessageID:     "biz-1",
				ChannelID:     "ch-1",
				Direction:     entity.MessageDirectionInbound,
				MessageStatus: entity.MessageStatusFailed,
				ErrorMessage:  "processor boom",
				RetryCount:    1,
				ReceivedAt:    time.Now().UTC(),
			},
		},
	}

	uc := &messageUseCase{
		messageRepo: repo,
		logger:      logger.New("error"),
	}

	result, err := uc.TroubleshootMessage(context.Background(), "biz-1")
	require.NoError(t, err)
	require.Equal(t, "biz-1", result.MessageID)
	require.Equal(t, MessageStageProcess, result.Stage)
	require.True(t, result.CanRetry)
	require.True(t, result.Recoverable)
	require.Equal(t, "retry_inbound_process", result.SuggestedAction)
}
