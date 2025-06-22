package dto

import (
	"time"
)

// AccessTokenResponse 访问令牌响应
type AccessTokenResponse struct {
	AccessToken string     `json:"access_token"`
	ExpiresIn   int        `json:"expires_in"` // 过期时间，秒
	ExpiresAt   *time.Time `json:"expires_at"` // 过期时间点
}

// WebhookEvent Webhook事件基础结构
type WebhookEvent struct {
	EventType string                 `json:"event_type"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}
