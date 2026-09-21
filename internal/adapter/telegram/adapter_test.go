package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPlatformType(t *testing.T) {
	a := NewAdapter(zerolog.Nop())
	assert.Equal(t, entity.PlatformTypeTelegram, a.GetPlatformType())
}

func TestValidateConfig(t *testing.T) {
	a := NewAdapter(zerolog.Nop())

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "valid bot_token",
			config:  map[string]interface{}{"bot_token": "123:ABC"},
			wantErr: false,
		},
		{
			name:    "missing bot_token",
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name:    "empty bot_token",
			config:  map[string]interface{}{"bot_token": ""},
			wantErr: true,
		},
		{
			name:    "non-string bot_token",
			config:  map[string]interface{}{"bot_token": 123},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := a.ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseInbound_SampleGetUpdates(t *testing.T) {
	a := NewAdapter(zerolog.Nop())

	sample := `{
		"update_id": 900001,
		"message": {
			"message_id": 42,
			"from": {
				"id": 10001,
				"is_bot": false,
				"first_name": "Alice",
				"username": "alice"
			},
			"chat": {
				"id": 10001,
				"type": "private",
				"first_name": "Alice",
				"username": "alice"
			},
			"date": 1710000000,
			"text": "hello bot"
		}
	}`

	msg, err := a.ParseInbound([]byte(sample), "ch-1")
	require.NoError(t, err)
	require.NotNil(t, msg)

	assert.Equal(t, entity.MessageTypeText, msg.MessageType)
	assert.Equal(t, "10001", msg.SenderID)
	assert.Equal(t, "alice", msg.SenderName)
	assert.Equal(t, entity.SenderTypeUser, msg.SenderType)
	assert.Equal(t, "10001", msg.ReceiverID)
	assert.Equal(t, "hello bot", msg.Content)
	assert.Equal(t, "42", msg.PlatformMessageID)
	assert.Equal(t, "ch-1", msg.RawContent["channel_id"])
	assert.Equal(t, "tg-900001-42", msg.MessageID)
}

func TestParseInbound_SkipBotAndEmpty(t *testing.T) {
	a := NewAdapter(zerolog.Nop())

	botMsg := `{
		"update_id": 1,
		"message": {
			"message_id": 1,
			"from": {"id": 2, "is_bot": true, "first_name": "Bot"},
			"chat": {"id": 2, "type": "private"},
			"date": 1,
			"text": "should skip"
		}
	}`
	msg, err := a.ParseInbound([]byte(botMsg), "")
	require.NoError(t, err)
	assert.Nil(t, msg)

	empty := `{
		"update_id": 2,
		"message": {
			"message_id": 2,
			"from": {"id": 3, "is_bot": false, "first_name": "Bob"},
			"chat": {"id": 3, "type": "private"},
			"date": 1
		}
	}`
	msg, err = a.ParseInbound([]byte(empty), "")
	require.NoError(t, err)
	assert.Nil(t, msg)
}

func TestSendMessage_RequestShape(t *testing.T) {
	var gotPath string
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))

		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(raw, &gotBody))

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer srv.Close()

	a := NewAdapter(zerolog.Nop())
	a.SetAPIBase(srv.URL)
	a.SetHTTPClient(srv.Client())

	cfg := map[string]interface{}{"bot_token": "999:TOKEN"}
	msg := &dto.UnifiedMessage{
		MessageID:   "out-1",
		MessageType: entity.MessageTypeText,
		ReceiverID:  "10001",
		Content:     "pong",
	}

	err := a.SendMessage(context.Background(), msg, cfg, "ignored-access-token")
	require.NoError(t, err)

	assert.Equal(t, "/bot999:TOKEN/sendMessage", gotPath)
	assert.Equal(t, "pong", gotBody["text"])
	// chat_id should be numeric when ReceiverID is digits
	assert.EqualValues(t, 10001, gotBody["chat_id"])
}

func TestSendMessage_ChatIDFromRawContent(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer srv.Close()

	a := NewAdapter(zerolog.Nop())
	a.SetAPIBase(srv.URL)
	a.SetHTTPClient(srv.Client())

	msg := &dto.UnifiedMessage{
		Content: "hi",
		RawContent: map[string]interface{}{
			"chat_id": int64(-100123),
		},
	}
	err := a.SendMessage(context.Background(), msg, map[string]interface{}{"bot_token": "t"}, "")
	require.NoError(t, err)
	assert.EqualValues(t, -100123, gotBody["chat_id"])
}

func TestStartStop_LongPoll(t *testing.T) {
	var polls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		polls.Add(1)
		// Return one update then empty
		if polls.Load() == 1 {
			_, _ = w.Write([]byte(`{
				"ok": true,
				"result": [{
					"update_id": 10,
					"message": {
						"message_id": 1,
						"from": {"id": 7, "is_bot": false, "first_name": "U"},
						"chat": {"id": 7, "type": "private"},
						"date": 1710000000,
						"text": "ping"
					}
				}]
			}`))
			return
		}
		// Simulate long-poll idle
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`{"ok":true,"result":[]}`))
	}))
	defer srv.Close()

	a := NewAdapter(zerolog.Nop())
	a.SetAPIBase(srv.URL)
	a.SetHTTPClient(srv.Client())

	received := make(chan *dto.UnifiedMessage, 1)
	handler := func(_ context.Context, m *dto.UnifiedMessage) error {
		received <- m
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := map[string]interface{}{
		"bot_token":  "tok",
		"channel_id": "c1",
	}
	require.NoError(t, a.Start(ctx, handler, cfg))
	assert.True(t, a.IsConnected())

	select {
	case m := <-received:
		assert.Equal(t, "ping", m.Content)
		assert.Equal(t, "c1", m.RawContent["channel_id"])
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for inbound message")
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer stopCancel()
	require.NoError(t, a.Stop(stopCtx))
	assert.False(t, a.IsConnected())
}
