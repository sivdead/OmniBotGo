package a2a

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	defaultTimeout    = 120 * time.Second
	agentCardPath     = "/.well-known/agent.json"
	protocolJSONRPC   = "2.0"
	methodMessageSend = "message/send"
)

// Client is a minimal A2A JSON-RPC client (protocol 0.3.0).
// It discovers the agent via Agent Card and sends user text, returning the final reply text.
type Client struct {
	baseURL    string
	httpClient *http.Client
	rpcURL     string // resolved from agent card when available
}

// NewClient creates an A2A client targeting baseURL (e.g. http://127.0.0.1:10000).
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// SetHTTPClient injects an HTTP client (tests).
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// FetchAgentCard GET /.well-known/agent.json and caches the preferred RPC URL.
func (c *Client) FetchAgentCard(ctx context.Context) (*AgentCard, error) {
	url := c.baseURL + agentCardPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("a2a: create agent card request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("a2a: fetch agent card: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("a2a: read agent card: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("a2a: agent card HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var card AgentCard
	if err := json.Unmarshal(body, &card); err != nil {
		return nil, fmt.Errorf("a2a: decode agent card: %w", err)
	}
	if card.URL != "" {
		c.rpcURL = strings.TrimRight(card.URL, "/")
	}
	return &card, nil
}

// SendText sends user text via message/send (blocking) and returns the final reply text.
func (c *Client) SendText(ctx context.Context, text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("a2a: empty text")
	}

	rpcURL := c.rpcURL
	if rpcURL == "" {
		rpcURL = c.baseURL
	}

	msgID := uuid.NewString()
	reqBody := jsonRPCRequest{
		JSONRPC: protocolJSONRPC,
		ID:      msgID,
		Method:  methodMessageSend,
		Params: messageSendParams{
			Message: a2aMessage{
				Kind:      "message",
				MessageID: msgID,
				Role:      "user",
				Parts:     []a2aPart{{Kind: "text", Text: text}},
			},
			Configuration: &messageSendConfiguration{Blocking: true},
		},
	}

	raw, err := c.doJSONRPC(ctx, rpcURL, reqBody)
	if err != nil {
		return "", err
	}

	return extractFinalText(raw)
}

func (c *Client) doJSONRPC(ctx context.Context, url string, payload jsonRPCRequest) (json.RawMessage, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("a2a: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ensureURL(url), bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("a2a: create rpc request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("a2a: rpc call: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("a2a: read rpc response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("a2a: rpc HTTP %d: %s", resp.StatusCode, truncate(string(body), 200))
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(body, &rpcResp); err != nil {
		return nil, fmt.Errorf("a2a: decode rpc response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("a2a: rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	if len(rpcResp.Result) == 0 {
		return nil, fmt.Errorf("a2a: empty rpc result")
	}
	return rpcResp.Result, nil
}

func ensureURL(u string) string {
	u = strings.TrimRight(u, "/")
	return u + "/"
}

func extractFinalText(result json.RawMessage) (string, error) {
	var tm taskOrMessage
	if err := json.Unmarshal(result, &tm); err != nil {
		return "", fmt.Errorf("a2a: decode result: %w", err)
	}

	// Prefer completed task artifacts.
	if tm.Kind == "task" || tm.Kind == "" {
		for i := len(tm.Artifacts) - 1; i >= 0; i-- {
			if t := partsText(tm.Artifacts[i].Parts); t != "" {
				return t, nil
			}
		}
		if tm.Status != nil && tm.Status.Message != nil {
			if t := partsText(tm.Status.Message.Parts); t != "" {
				return t, nil
			}
		}
		for i := len(tm.History) - 1; i >= 0; i-- {
			if tm.History[i].Role == "agent" {
				if t := partsText(tm.History[i].Parts); t != "" {
					return t, nil
				}
			}
		}
		if tm.Status != nil && tm.Status.State != "" && tm.Status.State != "completed" {
			return "", fmt.Errorf("a2a: task state %q with no reply text", tm.Status.State)
		}
	}

	if tm.Kind == "message" || len(tm.Parts) > 0 {
		if t := partsText(tm.Parts); t != "" {
			return t, nil
		}
	}

	return "", fmt.Errorf("a2a: no text in result")
}

func partsText(parts []a2aPart) string {
	var b strings.Builder
	for _, p := range parts {
		if p.Kind == "text" && p.Text != "" {
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
