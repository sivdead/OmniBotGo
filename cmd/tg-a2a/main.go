// Command tg-a2a is the private MVP golden path: Telegram long-poll → A2A client → Telegram reply.
// It bypasses DB / ConnectionManager / processor routing so outbound send is not blocked by
// stale channel.connection_status (「通道未就绪」).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/a2a"
	"github.com/sivdead/OmniBotGo/internal/adapter/telegram"
	"github.com/sivdead/OmniBotGo/internal/dto"
	"github.com/spf13/viper"
)

const (
	defaultA2ABaseURL = "http://127.0.0.1:10000"
	defaultTokenPath  = "/home/box/.config/omnibotgo/telegram_bot_token"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	token, err := loadBotToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	a2aBase := loadA2ABaseURL()
	client := a2a.NewClient(a2aBase)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	card, err := client.FetchAgentCard(ctx)
	if err != nil {
		logger.Warn().Err(err).Str("base_url", a2aBase).Msg("failed to fetch agent card (will still try message/send)")
	} else {
		logger.Info().
			Str("agent", card.Name).
			Str("protocol", card.ProtocolVersion).
			Str("rpc_url", card.URL).
			Msg("A2A agent card loaded")
	}

	ad := telegram.NewAdapter(logger)
	cfg := map[string]interface{}{"bot_token": token, "channel_id": "tg-a2a"}

	handler := func(hctx context.Context, msg *dto.UnifiedMessage) error {
		logger.Info().
			Str("from", msg.SenderID).
			Str("text", msg.Content).
			Msg("inbound telegram")

		reply, err := client.SendText(hctx, msg.Content)
		if err != nil {
			logger.Error().Err(err).Msg("a2a send failed")
			reply = "Sorry, the agent is unavailable right now."
		}

		out := &dto.UnifiedMessage{
			MessageID:   "a2a-" + msg.MessageID,
			MessageType: msg.MessageType,
			ReceiverID:  msg.ReceiverID,
			Content:     reply,
			RawContent:  msg.RawContent,
		}
		if err := ad.SendMessage(hctx, out, cfg, ""); err != nil {
			return fmt.Errorf("telegram send: %w", err)
		}
		logger.Info().Str("reply_len", fmt.Sprintf("%d", len(reply))).Msg("outbound telegram sent")
		return nil
	}

	if err := ad.Start(ctx, handler, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	logger.Info().
		Str("a2a_base_url", a2aBase).
		Msg("listening for Telegram → A2A (Ctrl+C to stop)")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	stopCtx, c2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer c2()
	_ = ad.Stop(stopCtx)
}

func loadBotToken() (string, error) {
	if t := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")); t != "" {
		return t, nil
	}
	path := os.Getenv("TELEGRAM_BOT_TOKEN_FILE")
	if path == "" {
		path = defaultTokenPath
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("telegram token: set TELEGRAM_BOT_TOKEN or create %s: %w", path, err)
	}
	return string(trimSpace(b)), nil
}

func loadA2ABaseURL() string {
	if u := strings.TrimSpace(os.Getenv("A2A_BASE_URL")); u != "" {
		return u
	}
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.SetDefault("a2a.base_url", defaultA2ABaseURL)
	_ = v.ReadInConfig()
	return strings.TrimSpace(v.GetString("a2a.base_url"))
}

func trimSpace(b []byte) []byte {
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	for j > i && (b[j-1] == ' ' || b[j-1] == '\n' || b[j-1] == '\r' || b[j-1] == '\t') {
		j--
	}
	return b[i:j]
}
