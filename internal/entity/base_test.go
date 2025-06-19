package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBaseEntity(t *testing.T) {
	entity := BaseEntity{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, int64(1), entity.ID)
	assert.False(t, entity.CreatedAt.IsZero())
	assert.False(t, entity.UpdatedAt.IsZero())
}

func TestJSONField_SetAndGet(t *testing.T) {
	jsonField := make(JSONField)

	// 测试设置和获取字符串值
	jsonField.Set("key1", "value1")
	assert.Equal(t, "value1", jsonField.GetString("key1"))

	// 测试设置和获取整数值
	jsonField.Set("key2", 42)
	assert.Equal(t, 42, jsonField.GetInt("key2"))

	// 测试设置和获取布尔值
	jsonField.Set("key3", true)
	assert.True(t, jsonField.GetBool("key3"))

	// 测试获取不存在的键
	assert.Empty(t, jsonField.GetString("nonexistent"))
	assert.Equal(t, 0, jsonField.GetInt("nonexistent"))
	assert.False(t, jsonField.GetBool("nonexistent"))
}

func TestJSONField_Value(t *testing.T) {
	jsonField := JSONField{
		"string_field": "test",
		"int_field":    123,
		"bool_field":   true,
	}

	value, err := jsonField.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// 测试nil值
	var nilField JSONField
	value, err = nilField.Value()
	assert.NoError(t, err)
	assert.Nil(t, value)
}

func TestJSONField_Scan(t *testing.T) {
	jsonField := JSONField{}

	// 测试从字节数组扫描
	jsonBytes := []byte(`{"key1":"value1","key2":42}`)
	err := jsonField.Scan(jsonBytes)
	assert.NoError(t, err)
	assert.Equal(t, "value1", jsonField.GetString("key1"))
	assert.Equal(t, 42, jsonField.GetInt("key2"))

	// 测试从字符串扫描
	jsonField2 := JSONField{}
	jsonString := `{"key3":"value3","key4":true}`
	err = jsonField2.Scan(jsonString)
	assert.NoError(t, err)
	assert.Equal(t, "value3", jsonField2.GetString("key3"))
	assert.True(t, jsonField2.GetBool("key4"))

	// 测试扫描nil值
	jsonField3 := JSONField{}
	err = jsonField3.Scan(nil)
	assert.NoError(t, err)
	assert.NotNil(t, jsonField3)
	assert.Empty(t, jsonField3.GetString("any"))
}

func TestJSONField_String(t *testing.T) {
	jsonField := JSONField{
		"key1": "value1",
		"key2": 42,
	}

	str := jsonField.String()
	assert.Contains(t, str, "key1")
	assert.Contains(t, str, "value1")
	assert.Contains(t, str, "key2")
}

func TestStatus(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
		isActive bool
	}{
		{StatusInactive, "inactive", false},
		{StatusActive, "active", true},
		{StatusDeleted, "deleted", false},
		{StatusSuspended, "suspended", false},
		{Status(99), "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
			assert.Equal(t, tt.isActive, tt.status.IsActive())
		})
	}
}

func TestMessageDirection(t *testing.T) {
	tests := []struct {
		direction MessageDirection
		expected  string
	}{
		{MessageDirectionInbound, "inbound"},
		{MessageDirectionOutbound, "outbound"},
		{MessageDirection(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.direction.String())
		})
	}
}

func TestMessageStatus(t *testing.T) {
	tests := []struct {
		status   MessageStatus
		expected string
	}{
		{MessageStatusPending, "pending"},
		{MessageStatusProcessing, "processing"},
		{MessageStatusProcessed, "processed"},
		{MessageStatusSent, "sent"},
		{MessageStatusFailed, "failed"},
		{MessageStatusExpired, "expired"},
		{MessageStatus(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestConnectionStatus(t *testing.T) {
	tests := []struct {
		status   ConnectionStatus
		expected string
	}{
		{ConnectionStatusDisconnected, "disconnected"},
		{ConnectionStatusConnected, "connected"},
		{ConnectionStatusConnecting, "connecting"},
		{ConnectionStatusError, "error"},
		{ConnectionStatus(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestQueueStatus(t *testing.T) {
	tests := []struct {
		status   QueueStatus
		expected string
	}{
		{QueueStatusPending, "pending"},
		{QueueStatusRunning, "running"},
		{QueueStatusCompleted, "completed"},
		{QueueStatusFailed, "failed"},
		{QueueStatusRetrying, "retrying"},
		{QueueStatusCancelled, "cancelled"},
		{QueueStatus(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestRouteType(t *testing.T) {
	tests := []struct {
		routeType RouteType
		expected  string
	}{
		{RouteTypeDirect, "direct"},
		{RouteTypeForward, "forward"},
		{RouteTypeBroadcast, "broadcast"},
		{RouteTypeConditional, "conditional"},
		{RouteType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.routeType.String())
		})
	}
}
