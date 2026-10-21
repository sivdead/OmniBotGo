package usecase

import (
	"context"
	"testing"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/stretchr/testify/require"
)

const testAIMsgID = "m1"

type convRepoStub struct {
	msgRepoStub
	lastChannelID string
	lastConvID    string
	items         []*entity.Message
}

func (c *convRepoStub) GetByConversationID(_ context.Context, channelID, conversationID string, _ port.ListParams) (*port.PaginatedResult[*entity.Message], error) {
	c.lastChannelID = channelID
	c.lastConvID = conversationID

	return &port.PaginatedResult[*entity.Message]{Items: c.items}, nil
}

func TestGetConversationHistory_ScopesByChannel(t *testing.T) {
	t.Parallel()

	repo := &convRepoStub{
		items: []*entity.Message{{MessageID: testAIMsgID, ChannelID: "ch-a", ConversationID: "same-conv"}},
	}
	uc := &messageUseCase{messageRepo: repo, logger: logger.New("error")}

	got, err := uc.getConversationHistory(context.Background(), "ch-a", "same-conv", 10)
	require.NoError(t, err)
	require.Len(t, got, 1)
	require.Equal(t, "ch-a", repo.lastChannelID)
	require.Equal(t, "same-conv", repo.lastConvID)

	empty, err := uc.getConversationHistory(context.Background(), "", "same-conv", 10)
	require.NoError(t, err)
	require.Empty(t, empty)
}
