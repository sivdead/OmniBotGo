package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/goccy/go-json"
)

const (
	// Base settings
	host     = "app"
	attempts = 20

	// Attempts connection
	httpURL        = "http://" + host + ":8080"
	healthPath     = httpURL + "/healthz"
	requestTimeout = 5 * time.Second

	// HTTP REST
	basePathV1 = httpURL + "/api/v1"
)

var errHealthCheck = fmt.Errorf("url %s is not available", healthPath)

func doWebRequestWithTimeout(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
}

func getHealthCheck(url string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func healthCheck(attempts int) error {
	for attempts > 0 {
		statusCode, err := getHealthCheck(healthPath)
		if err != nil {
			log.Printf("Integration tests: url %s is not available, attempts left: %d: %v", healthPath, attempts, err)
			time.Sleep(time.Second)

			attempts--

			continue
		}

		if statusCode == http.StatusOK {
			return nil
		}

		log.Printf("Integration tests: url %s is not available, attempts left: %d", healthPath, attempts)

		time.Sleep(time.Second)

		attempts--
	}

	return errHealthCheck
}

func TestMain(m *testing.M) {
	err := healthCheck(attempts)
	if err != nil {
		log.Fatalf("Integration tests: httpURL %s is not available: %s", httpURL, err)
	}

	log.Printf("Integration tests: httpURL %s is available", httpURL)

	code := m.Run()
	os.Exit(code)
}

// HTTP POST: /api/v1/bots - 创建机器人
func TestHTTPCreateBot(t *testing.T) {
	tests := []struct {
		description string
		body        string
		expected    int
	}{
		{
			description: "Create Bot Success",
			body: `{
				"bot_name": "测试机器人",
				"bot_type": "general",
				"description": "集成测试机器人",
				"created_by": "test"
			}`,
			expected: http.StatusCreated,
		},
		{
			description: "Create Bot Fail - Missing Name",
			body: `{
				"bot_type": "general",
				"description": "测试机器人"
			}`,
			expected: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			url := basePathV1 + "/bots"
			ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

			defer cancel()

			resp, err := doWebRequestWithTimeout(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			defer resp.Body.Close()

			if resp.StatusCode != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, resp.StatusCode)
			}
		})
	}
}

// HTTP GET: /api/v1/bots - 获取机器人列表
func TestHTTPListBots(t *testing.T) {
	url := basePathV1 + "/bots"
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)

	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Items []struct {
				ID          string `json:"id"`
				BotName     string `json:"bot_name"`
				BotType     string `json:"bot_type"`
				Description string `json:"description"`
				Status      int8   `json:"status"`
			} `json:"items"`
			Total      int64 `json:"total"`
			Page       int   `json:"page"`
			PageSize   int   `json:"page_size"`
			TotalPages int   `json:"total_pages"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if !body.Success {
		t.Error("Expected success=true")
	}

	// 检查响应结构是否正确（StandardResponse.data 内的分页字段）
	if body.Data.Page == 0 {
		t.Error("Expected page > 0")
	}

	if body.Data.PageSize == 0 {
		t.Error("Expected page_size > 0")
	}
}

// HTTP POST: /api/v1/channels - 创建通道
func TestHTTPCreateChannel(t *testing.T) {
	// 首先创建一个机器人
	botBody := `{
		"bot_name": "通道测试机器人",
		"bot_type": "general",
		"description": "用于通道测试的机器人",
		"created_by": "test"
	}`

	url := basePathV1 + "/bots"
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := doWebRequestWithTimeout(ctx, http.MethodPost, url, bytes.NewBuffer([]byte(botBody)))
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Skipf("Cannot create bot for channel test, status: %d", resp.StatusCode)
		return
	}

	var botResp struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&botResp); err != nil {
		t.Fatalf("Failed to decode bot response: %v", err)
	}

	if botResp.Data.ID == "" {
		t.Fatal("Expected non-empty bot id (UUID string)")
	}

	// 现在创建通道（bot_id 为 UUID 字符串）
	channelBody := fmt.Sprintf(`{
		"bot_id": %q,
		"platform_type": "wecom",
		"channel_name": "测试通道",
		"config": {
			"corp_id": "test_corp",
			"agent_id": "test_agent",
			"secret": "test_secret"
		}
	}`, botResp.Data.ID)

	channelURL := basePathV1 + "/channels"
	ctx2, cancel2 := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel2()

	channelResp, err := doWebRequestWithTimeout(ctx2, http.MethodPost, channelURL, bytes.NewBuffer([]byte(channelBody)))
	if err != nil {
		t.Fatalf("Failed to create channel: %v", err)
	}
	defer channelResp.Body.Close()

	if channelResp.StatusCode != http.StatusCreated {
		bodyBytes, readErr := io.ReadAll(channelResp.Body)
		if readErr != nil {
			t.Errorf("Expected status %d, got %d (also failed reading body: %v)", http.StatusCreated, channelResp.StatusCode, readErr)
		} else {
			t.Errorf("Expected status %d, got %d body=%s", http.StatusCreated, channelResp.StatusCode, string(bodyBytes))
		}
	}
}
