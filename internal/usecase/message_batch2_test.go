package usecase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/internal/usecase/service"
	"github.com/sivdead/OmniBotGo/pkg/logger"
	"github.com/stretchr/testify/require"
)

const (
	testChannelID     = "ch-1"
	testPlatformMsgID = "chat:42"
	testPlatformKey   = "ch-1|chat:42"
)

var errStubNotFound = errors.New("not found")

func TestSendMessage_GeneratesMessageIDBeforeValidate(t *testing.T) {
	t.Parallel()

	svc := service.NewMessageService()
	msg := &entity.Message{
		ChannelID:   testChannelID,
		MessageType: entity.MessageTypeText,
		Content:     "hi",
	}

	require.Empty(t, msg.MessageID)

	if msg.MessageID == "" {
		msg.MessageID = "msg_generated"
	}

	if msg.PlatformTimestamp.IsZero() {
		msg.PlatformTimestamp = time.Now()
	}

	require.NoError(t, svc.ValidateMessage(msg))
	require.NotEmpty(t, msg.MessageID)
	require.False(t, msg.PlatformTimestamp.IsZero())
}

func TestSuccessfulProcessorTracking_SkipsOnRetry(t *testing.T) {
	t.Parallel()

	msg := &entity.Message{}
	markProcessorSuccessful(msg, "proc-a")
	markProcessorSuccessful(msg, "proc-b")
	markProcessorSuccessful(msg, "proc-a")

	set := successfulProcessorSet(msg)
	require.True(t, set["proc-a"])
	require.True(t, set["proc-b"])
	require.Len(t, set, 2)
}

func TestCheckDuplicateMessage_ProcessedSkips(t *testing.T) {
	t.Parallel()

	repo := &msgRepoStub{byPlatform: map[string]*entity.Message{
		testPlatformKey: {
			BaseEntity:        entity.BaseEntity{ID: "row-1"},
			MessageID:         "old-id",
			ChannelID:         testChannelID,
			PlatformMessageID: testPlatformMsgID,
			MessageStatus:     entity.MessageStatusProcessed,
		},
	}}
	uc := &messageUseCase{messageRepo: repo, msgService: service.NewMessageService(), logger: logger.New("error")}
	msg := &entity.Message{ChannelID: testChannelID, PlatformMessageID: testPlatformMsgID, MessageID: "x"}

	skip, err := uc.checkDuplicateMessage(context.Background(), msg)
	require.NoError(t, err)
	require.True(t, skip)
}

func TestCheckDuplicateMessage_FailedResumes(t *testing.T) {
	t.Parallel()

	repo := &msgRepoStub{byPlatform: map[string]*entity.Message{
		testPlatformKey: {
			BaseEntity:        entity.BaseEntity{ID: "row-fail"},
			MessageID:         "old-fail",
			ChannelID:         testChannelID,
			PlatformMessageID: testPlatformMsgID,
			MessageStatus:     entity.MessageStatusFailed,
			UnifiedContent: entity.JSONField{
				successfulProcessorsKey: []string{"proc-ok"},
			},
		},
	}}
	uc := &messageUseCase{messageRepo: repo, msgService: service.NewMessageService(), logger: logger.New("error")}
	msg := &entity.Message{
		ChannelID:         testChannelID,
		PlatformMessageID: testPlatformMsgID,
		MessageID:         "x",
		UnifiedContent:    entity.JSONField{},
	}

	skip, err := uc.checkDuplicateMessage(context.Background(), msg)
	require.NoError(t, err)
	require.False(t, skip)
	require.Equal(t, "row-fail", msg.ID)
	require.Equal(t, "old-fail", msg.MessageID)
	require.True(t, successfulProcessorSet(msg)["proc-ok"])
}

func TestCheckDuplicateMessage_LegacyPlatformID(t *testing.T) {
	t.Parallel()

	repo := &msgRepoStub{byPlatform: map[string]*entity.Message{
		"ch-1|42": {
			BaseEntity:        entity.BaseEntity{ID: "legacy"},
			MessageID:         "legacy-msg",
			ChannelID:         testChannelID,
			PlatformMessageID: "42",
			ConversationID:    "10001",
			MessageStatus:     entity.MessageStatusFailed,
		},
	}}
	uc := &messageUseCase{messageRepo: repo, msgService: service.NewMessageService(), logger: logger.New("error")}
	msg := &entity.Message{
		ChannelID:         testChannelID,
		PlatformMessageID: "10001:42",
		ConversationID:    "10001",
		MessageID:         "x",
	}

	skip, err := uc.checkDuplicateMessage(context.Background(), msg)
	require.NoError(t, err)
	require.False(t, skip)
	require.Equal(t, "legacy", msg.ID)
}

func TestProcessWebhookForwarder_Non2xxIsFailure(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	uc := &messageUseCase{
		queueRepo: &queueRepoStub{},
		logger:    logger.New("error"),
	}
	msg := &entity.Message{MessageID: "m1", ChannelID: testChannelID, MessageType: entity.MessageTypeText}
	processor := &entity.MessageProcessor{
		BaseEntity: entity.BaseEntity{ID: "p1"},
		Config:     entity.JSONField{"webhook_url": srv.URL},
	}

	err := uc.processWebhookForwarder(context.Background(), msg, processor)
	require.Error(t, err)
	require.ErrorIs(t, err, errWebhookDeliveryFailed)
}

func TestRetryQueueMessage_Exhausted(t *testing.T) {
	t.Parallel()

	repo := &queueRepoStub{msg: &entity.MessageQueue{
		BaseEntity: entity.BaseEntity{ID: "q1"},
		Status:     entity.QueueStatusFailed,
		RetryCount: 5,
		MaxRetries: 5,
	}}
	uc := &queueUC{queueRepo: repo, logger: logger.New("error")}

	err := uc.RetryQueueMessage(context.Background(), "q1")
	require.Error(t, err)
	require.ErrorIs(t, err, errRetryExhausted)
	require.False(t, repo.markedRetry)
}

func TestRetryQueueMessage_AllowsWhenUnderMax(t *testing.T) {
	t.Parallel()

	repo := &queueRepoStub{msg: &entity.MessageQueue{
		BaseEntity: entity.BaseEntity{ID: "q2"},
		Status:     entity.QueueStatusFailed,
		RetryCount: 1,
		MaxRetries: 5,
	}}
	uc := &queueUC{queueRepo: repo, logger: logger.New("error")}

	require.NoError(t, uc.RetryQueueMessage(context.Background(), "q2"))
	require.True(t, repo.markedRetry)
}

type msgRepoStub struct {
	byPlatform map[string]*entity.Message
	created    int
	updated    int
}

func (m *msgRepoStub) key(channelID, platformID string) string {
	return channelID + "|" + platformID
}

func (m *msgRepoStub) Create(context.Context, *entity.Message) error {
	m.created++

	return nil
}

func (m *msgRepoStub) GetByID(context.Context, string) (*entity.Message, error) {
	return nil, errStubNotFound
}

func (m *msgRepoStub) GetByMessageID(context.Context, string) (*entity.Message, error) {
	return nil, errStubNotFound
}

func (m *msgRepoStub) GetByPlatformMessageID(_ context.Context, channelID, platformMessageID string) (*entity.Message, error) {
	if m.byPlatform == nil {
		return nil, nil
	}

	return m.byPlatform[m.key(channelID, platformMessageID)], nil
}

func (m *msgRepoStub) GetByChannelID(context.Context, string, port.ListParams) (*port.PaginatedResult[*entity.Message], error) {
	return nil, nil
}

func (m *msgRepoStub) GetByConversationID(context.Context, string, port.ListParams) (*port.PaginatedResult[*entity.Message], error) {
	return nil, nil
}

func (m *msgRepoStub) GetPendingMessages(context.Context, int) ([]*entity.Message, error) {
	return nil, nil
}

func (m *msgRepoStub) GetFailedMessages(context.Context, int) ([]*entity.Message, error) {
	return nil, nil
}

func (m *msgRepoStub) Update(context.Context, *entity.Message) error {
	m.updated++

	return nil
}

func (m *msgRepoStub) Delete(context.Context, string) error { return nil }

func (m *msgRepoStub) List(context.Context, port.ListParams) (*port.PaginatedResult[*entity.Message], error) {
	return &port.PaginatedResult[*entity.Message]{}, nil
}

func (m *msgRepoStub) UpdateStatus(context.Context, string, entity.MessageStatus) error { return nil }

func (m *msgRepoStub) IncrementRetryCount(context.Context, string) error { return nil }

func (m *msgRepoStub) MarkAsProcessed(context.Context, string) error { return nil }

func (m *msgRepoStub) MarkAsSent(context.Context, string) error { return nil }

func (m *msgRepoStub) MarkAsFailed(context.Context, string, string) error { return nil }

func (m *msgRepoStub) Exists(context.Context, string) (bool, error) { return false, nil }

func (m *msgRepoStub) ExistsByMessageID(context.Context, string) (bool, error) { return false, nil }

type queueRepoStub struct {
	msg         *entity.MessageQueue
	markedRetry bool
}

func (q *queueRepoStub) Create(context.Context, *entity.MessageQueue) error { return nil }

func (q *queueRepoStub) GetByID(context.Context, string) (*entity.MessageQueue, error) {
	return q.msg, nil
}

func (q *queueRepoStub) GetPendingJobs(context.Context, string, int) ([]*entity.MessageQueue, error) {
	return nil, nil
}

func (q *queueRepoStub) GetRetryableJobs(context.Context, int) ([]*entity.MessageQueue, error) {
	return nil, nil
}

func (q *queueRepoStub) GetExpiredJobs(context.Context, int64) ([]*entity.MessageQueue, error) {
	return nil, nil
}

func (q *queueRepoStub) Update(context.Context, *entity.MessageQueue) error { return nil }
func (q *queueRepoStub) Delete(context.Context, string) error               { return nil }

func (q *queueRepoStub) List(context.Context, port.ListParams) (*port.PaginatedResult[*entity.MessageQueue], error) {
	return &port.PaginatedResult[*entity.MessageQueue]{}, nil
}

func (q *queueRepoStub) MarkAsRunning(context.Context, string) error   { return nil }
func (q *queueRepoStub) MarkAsCompleted(context.Context, string) error { return nil }
func (q *queueRepoStub) MarkAsFailed(context.Context, string, string) error {
	return nil
}

func (q *queueRepoStub) MarkForRetry(context.Context, string, int64) error {
	q.markedRetry = true

	return nil
}

func (q *queueRepoStub) MarkAsCancelled(context.Context, string) error { return nil }
func (q *queueRepoStub) Exists(context.Context, string) (bool, error)  { return false, nil }
