package entity

import (
	"time"
)

// ConnectionLog 连接日志实体，记录通道连接状态变化
type ConnectionLog struct {
	BaseEntity
	ChannelID    int64     `json:"channel_id" gorm:"column:channel_id;not null;index;comment:通道ID"`
	EventType    string    `json:"event_type" gorm:"column:event_type;type:varchar(50);not null;comment:事件类型"`
	Status       Status    `json:"status" gorm:"column:status;type:tinyint;not null;comment:状态:0-失败,1-成功"`
	ErrorCode    string    `json:"error_code" gorm:"column:error_code;type:varchar(50);comment:错误代码"`
	ErrorMessage string    `json:"error_message" gorm:"column:error_message;type:text;comment:错误消息"`
	Details      JSONField `json:"details" gorm:"column:details;type:json;comment:详细信息"`

	// 关联关系
	Channel *Channel `json:"channel,omitempty" gorm:"foreignKey:ChannelID;references:ID"`
}

// TableName 指定表名
func (ConnectionLog) TableName() string {
	return "connection_logs"
}

// IsSuccess 检查是否为成功状态
func (cl *ConnectionLog) IsSuccess() bool {
	return cl.Status == StatusActive
}

// SetDetail 设置详细信息中的特定值
func (cl *ConnectionLog) SetDetail(key string, value interface{}) {
	if cl.Details == nil {
		cl.Details = make(JSONField)
	}
	cl.Details.Set(key, value)
}

// GetDetail 获取详细信息中的特定值
func (cl *ConnectionLog) GetDetail(key string) interface{} {
	if cl.Details == nil {
		return nil
	}
	return cl.Details.Get(key)
}

// Validate 验证ConnectionLog实体数据
func (cl *ConnectionLog) Validate() error {
	if cl.ChannelID <= 0 {
		return NewValidationError("channel_id", "通道ID必须大于0")
	}
	if cl.EventType == "" {
		return NewValidationError("event_type", "事件类型不能为空")
	}
	return nil
}

// APICallLog API调用日志实体，记录对外API调用详情
type APICallLog struct {
	BaseEntity
	RequestID       string    `json:"request_id" gorm:"column:request_id;type:varchar(100);not null;index;comment:请求ID"`
	ChannelID       string    `json:"channel_id" gorm:"column:channel_id;index;comment:通道ID"`
	ProcessorID     string    `json:"processor_id" gorm:"column:processor_id;index;comment:处理器ID"`
	Method          string    `json:"method" gorm:"column:method;type:varchar(10);not null;comment:HTTP方法"`
	URL             string    `json:"url" gorm:"column:url;type:varchar(500);not null;comment:请求URL"`
	RequestHeaders  JSONField `json:"request_headers" gorm:"column:request_headers;type:json;comment:请求头"`
	RequestBody     string    `json:"request_body" gorm:"column:request_body;type:text;comment:请求体"`
	ResponseStatus  int       `json:"response_status" gorm:"column:response_status;comment:响应状态码"`
	ResponseHeaders JSONField `json:"response_headers" gorm:"column:response_headers;type:json;comment:响应头"`
	ResponseBody    string    `json:"response_body" gorm:"column:response_body;type:text;comment:响应体"`
	StartTime       time.Time `json:"start_time" gorm:"column:start_time;not null;comment:开始时间"`
	EndTime         time.Time `json:"end_time" gorm:"column:end_time;not null;comment:结束时间"`
	DurationMs      int       `json:"duration_ms" gorm:"column:duration_ms;comment:耗时(毫秒)"`
	Success         bool      `json:"success" gorm:"column:success;comment:是否成功"`
	ErrorMessage    string    `json:"error_message" gorm:"column:error_message;type:text;comment:错误消息"`

	// 关联关系
	Channel   *Channel          `json:"channel,omitempty" gorm:"foreignKey:ChannelID;references:ID"`
	Processor *MessageProcessor `json:"processor,omitempty" gorm:"foreignKey:ProcessorID;references:ID"`
}

// TableName 指定表名
func (APICallLog) TableName() string {
	return "api_call_logs"
}

// IsSuccess 检查API调用是否成功
func (acl *APICallLog) IsSuccess() bool {
	return acl.Success
}

// GetDuration 获取调用耗时
func (acl *APICallLog) GetDuration() time.Duration {
	return acl.EndTime.Sub(acl.StartTime)
}

// SetRequestHeader 设置请求头中的特定值
func (acl *APICallLog) SetRequestHeader(key, value string) {
	if acl.RequestHeaders == nil {
		acl.RequestHeaders = make(JSONField)
	}
	acl.RequestHeaders.Set(key, value)
}

// GetRequestHeader 获取请求头中的特定值
func (acl *APICallLog) GetRequestHeader(key string) string {
	if acl.RequestHeaders == nil {
		return ""
	}
	return acl.RequestHeaders.GetString(key)
}

// SetResponseHeader 设置响应头中的特定值
func (acl *APICallLog) SetResponseHeader(key, value string) {
	if acl.ResponseHeaders == nil {
		acl.ResponseHeaders = make(JSONField)
	}
	acl.ResponseHeaders.Set(key, value)
}

// GetResponseHeader 获取响应头中的特定值
func (acl *APICallLog) GetResponseHeader(key string) string {
	if acl.ResponseHeaders == nil {
		return ""
	}
	return acl.ResponseHeaders.GetString(key)
}

// MarkAsCompleted 标记调用完成并计算耗时
func (acl *APICallLog) MarkAsCompleted(responseStatus int, success bool) {
	acl.EndTime = time.Now()
	acl.ResponseStatus = responseStatus
	acl.Success = success
	acl.DurationMs = int(acl.GetDuration().Milliseconds())
}

// MarkAsFailed 标记调用失败并记录错误
func (acl *APICallLog) MarkAsFailed(errorMsg string) {
	acl.EndTime = time.Now()
	acl.Success = false
	acl.ErrorMessage = errorMsg
	acl.DurationMs = int(acl.GetDuration().Milliseconds())
}

// IsSlowCall 检查是否为慢调用（超过指定阈值）
func (acl *APICallLog) IsSlowCall(thresholdMs int) bool {
	return acl.DurationMs > thresholdMs
}

// Validate 验证APICallLog实体数据
func (acl *APICallLog) Validate() error {
	if acl.RequestID == "" {
		return NewValidationError("request_id", "请求ID不能为空")
	}
	if acl.Method == "" {
		return NewValidationError("method", "HTTP方法不能为空")
	}
	if acl.URL == "" {
		return NewValidationError("url", "请求URL不能为空")
	}
	return nil
}
