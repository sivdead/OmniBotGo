package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/config"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

const (
	defaultAPIBase       = "https://api.telegram.org"
	defaultPollTimeout   = 30
	defaultHTTPTimeout   = 60 * time.Second
	pollRetryBackoff     = 3 * time.Second
)

var (
	errAlreadyConnected = errors.New("telegram adapter is already connected")
	errMissingBotToken  = errors.New("bot_token is required for telegram platform")
)

// Adapter Telegram long-poll 适配器，实现 MessageSender / StreamAdapter / PlatformIdentifier / ConfigValidator。
type Adapter struct {
	logger     zerolog.Logger
	httpClient *http.Client
	apiBase    string

	config      map[string]interface{}
	cancel      context.CancelFunc
	isConnected bool
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

// NewAdapter 创建 Telegram 适配器。
func NewAdapter(logger zerolog.Logger) *Adapter {
	return &Adapter{
		logger: logger,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		apiBase:     defaultAPIBase,
		isConnected: false,
	}
}

// SetHTTPClient 注入 HTTP 客户端（测试用）。
func (a *Adapter) SetHTTPClient(client *http.Client) {
	a.httpClient = client
}

// SetAPIBase 注入 API Base URL（测试用，指向 httptest）。
func (a *Adapter) SetAPIBase(base string) {
	a.apiBase = base
}

// GetPlatformType 实现 PlatformIdentifier。
func (a *Adapter) GetPlatformType() entity.PlatformType {
	return entity.PlatformTypeTelegram
}

// ValidateConfig 实现 ConfigValidator。
func (a *Adapter) ValidateConfig(cfg map[string]interface{}) error {
	return config.ValidatePlatformConfig(entity.PlatformTypeTelegram, cfg)
}

// ParseInbound 将单个 Telegram Update JSON 解析为 UnifiedMessage。
// 跳过 bot 消息与空文本时返回 (nil, nil)。
func (a *Adapter) ParseInbound(raw []byte, channelID string) (*dto.UnifiedMessage, error) {
	var update Update
	if err := json.Unmarshal(raw, &update); err != nil {
		return nil, fmt.Errorf("failed to unmarshal telegram update: %w", err)
	}
	return updateToUnified(&update, channelID)
}

// Start 实现 StreamAdapter：后台 long-poll getUpdates。
func (a *Adapter) Start(ctx context.Context, messageHandler port.MessageHandler, cfg map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isConnected {
		return errAlreadyConnected
	}

	if err := a.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	tgCfg, err := config.ParseTelegramConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to parse telegram config: %w", err)
	}

	channelID, _ := cfg["channel_id"].(string)

	a.config = cfg
	loopCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.isConnected = true

	a.wg.Add(1)
	go a.pollLoop(loopCtx, messageHandler, tgCfg.BotToken, channelID)

	a.logger.Info().
		Str("channel_id", channelID).
		Msg("telegram adapter started (long-poll)")

	return nil
}

// Stop 实现 StreamAdapter：取消 long-poll 并等待退出。
func (a *Adapter) Stop(ctx context.Context) error {
	a.mu.Lock()
	if !a.isConnected {
		a.mu.Unlock()
		return nil
	}
	cancel := a.cancel
	a.cancel = nil
	a.isConnected = false
	a.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		a.logger.Warn().Msg("telegram adapter stop timed out waiting for poll loop")
	}

	a.logger.Info().Msg("telegram adapter stopped")
	return nil
}

// IsConnected 实现 StreamAdapter。
func (a *Adapter) IsConnected() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isConnected
}

// SendMessage 实现 MessageSender：POST sendMessage。
// accessToken 被忽略；bot_token 始终来自 config。
func (a *Adapter) SendMessage(ctx context.Context, message *dto.UnifiedMessage, cfg map[string]interface{}, _ string) error {
	if err := a.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	botToken, err := botTokenFromConfig(cfg)
	if err != nil {
		return err
	}

	chatID, err := resolveChatID(message)
	if err != nil {
		return err
	}

	text := message.Content
	if text == "" && message.MarkdownContent != nil {
		text = message.MarkdownContent.Content
	}

	body := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", a.apiBase, botToken)
	if err := a.doJSON(ctx, http.MethodPost, url, body, nil); err != nil {
		return fmt.Errorf("telegram sendMessage failed: %w", err)
	}

	a.logger.Info().
		Str("message_id", message.MessageID).
		Interface("chat_id", chatID).
		Msg("telegram message sent successfully")

	return nil
}

func (a *Adapter) pollLoop(ctx context.Context, handler port.MessageHandler, botToken, channelID string) {
	defer a.wg.Done()
	defer func() {
		a.mu.Lock()
		a.isConnected = false
		a.mu.Unlock()
	}()

	var offset int64

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		updates, nextOffset, err := a.getUpdates(ctx, botToken, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			a.logger.Error().Err(err).Msg("telegram getUpdates error, retrying")
			select {
			case <-ctx.Done():
				return
			case <-time.After(pollRetryBackoff):
			}
			continue
		}

		if nextOffset > offset {
			offset = nextOffset
		}

		for i := range updates {
			msg, err := updateToUnified(&updates[i], channelID)
			if err != nil {
				a.logger.Warn().Err(err).Int64("update_id", updates[i].UpdateID).Msg("skip telegram update")
				continue
			}
			if msg == nil {
				continue
			}
			if err := handler(ctx, msg); err != nil {
				a.logger.Error().Err(err).Str("message_id", msg.MessageID).Msg("telegram message handler error")
			}
		}
	}
}

func (a *Adapter) getUpdates(ctx context.Context, botToken string, offset int64) ([]Update, int64, error) {
	url := fmt.Sprintf("%s/bot%s/getUpdates?timeout=%d&offset=%d", a.apiBase, botToken, defaultPollTimeout, offset)

	var result []Update
	if err := a.doJSON(ctx, http.MethodGet, url, nil, &result); err != nil {
		return nil, offset, err
	}

	next := offset
	for _, u := range result {
		if u.UpdateID+1 > next {
			next = u.UpdateID + 1
		}
	}
	return result, next, nil
}

func (a *Adapter) doJSON(ctx context.Context, method, url string, body interface{}, resultDest interface{}) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(raw, &apiResp); err != nil {
		return fmt.Errorf("decode api response: %w", err)
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram api error %d: %s", apiResp.ErrorCode, apiResp.Description)
	}

	if resultDest != nil && len(apiResp.Result) > 0 {
		if err := json.Unmarshal(apiResp.Result, resultDest); err != nil {
			return fmt.Errorf("decode result: %w", err)
		}
	}
	return nil
}

func updateToUnified(update *Update, channelID string) (*dto.UnifiedMessage, error) {
	if update == nil {
		return nil, nil
	}

	msg := update.Message
	if msg == nil {
		msg = update.EditedMessage
	}
	if msg == nil {
		return nil, nil
	}

	if msg.From != nil && msg.From.IsBot {
		return nil, nil
	}
	if msg.Text == "" {
		return nil, nil
	}

	senderID := ""
	senderName := ""
	if msg.From != nil {
		senderID = strconv.FormatInt(msg.From.ID, 10)
		senderName = msg.From.FirstName
		if msg.From.Username != "" {
			senderName = msg.From.Username
		}
	}

	chatID := strconv.FormatInt(msg.Chat.ID, 10)
	receiverType := "user"
	if msg.Chat.Type == "group" || msg.Chat.Type == "supergroup" {
		receiverType = "group"
	}

	raw := map[string]interface{}{
		"update_id":  update.UpdateID,
		"chat_id":    msg.Chat.ID,
		"chat_type":  msg.Chat.Type,
		"message_id": msg.MessageID,
	}
	if channelID != "" {
		raw["channel_id"] = channelID
	}

	return &dto.UnifiedMessage{
		MessageID:         fmt.Sprintf("tg-%d-%d", update.UpdateID, msg.MessageID),
		MessageType:       entity.MessageTypeText,
		SenderID:          senderID,
		SenderName:        senderName,
		SenderType:        entity.SenderTypeUser,
		ReceiverID:        chatID,
		ReceiverType:      receiverType,
		Content:           msg.Text,
		RawContent:        raw,
		ConversationID:    chatID,
		PlatformMessageID: strconv.FormatInt(msg.MessageID, 10),
		PlatformTimestamp: time.Unix(msg.Date, 0).UTC(),
	}, nil
}

func resolveChatID(message *dto.UnifiedMessage) (interface{}, error) {
	if message.ReceiverID != "" {
		if id, err := strconv.ParseInt(message.ReceiverID, 10, 64); err == nil {
			return id, nil
		}
		return message.ReceiverID, nil
	}
	if message.RawContent != nil {
		if v, ok := message.RawContent["chat_id"]; ok {
			return v, nil
		}
	}
	return nil, errors.New("chat_id not found: set ReceiverID or RawContent.chat_id")
}

func botTokenFromConfig(cfg map[string]interface{}) (string, error) {
	token, ok := cfg["bot_token"].(string)
	if !ok || token == "" {
		return "", errMissingBotToken
	}
	return token, nil
}
