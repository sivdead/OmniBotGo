package entity

// Status 通用状态类型
type Status int8

const (
	StatusInactive  Status = 0 // 未激活/禁用
	StatusActive    Status = 1 // 激活/启用
	StatusDeleted   Status = 2 // 已删除
	StatusSuspended Status = 3 // 暂停
)

// String 返回状态的字符串表示
func (s Status) String() string {
	switch s {
	case StatusInactive:
		return "inactive"
	case StatusActive:
		return "active"
	case StatusDeleted:
		return "deleted"
	case StatusSuspended:
		return "suspended"
	default:
		return "unknown"
	}
}

// IsActive 检查状态是否为激活状态
func (s Status) IsActive() bool {
	return s == StatusActive
}

// MessageDirection 消息方向
type MessageDirection int8

const (
	MessageDirectionInbound  MessageDirection = 1 // 入站消息
	MessageDirectionOutbound MessageDirection = 2 // 出站消息
)

// String 返回消息方向的字符串表示
func (d MessageDirection) String() string {
	switch d {
	case MessageDirectionInbound:
		return "inbound"
	case MessageDirectionOutbound:
		return "outbound"
	default:
		return "unknown"
	}
}

// MessageStatus 消息状态
type MessageStatus int8

const (
	MessageStatusPending    MessageStatus = 0 // 待处理
	MessageStatusProcessing MessageStatus = 1 // 处理中
	MessageStatusProcessed  MessageStatus = 2 // 已处理
	MessageStatusSent       MessageStatus = 3 // 已发送
	MessageStatusFailed     MessageStatus = 4 // 处理失败
	MessageStatusExpired    MessageStatus = 5 // 已过期
)

// String 返回消息状态的字符串表示
func (s MessageStatus) String() string {
	switch s {
	case MessageStatusPending:
		return "pending"
	case MessageStatusProcessing:
		return "processing"
	case MessageStatusProcessed:
		return "processed"
	case MessageStatusSent:
		return "sent"
	case MessageStatusFailed:
		return "failed"
	case MessageStatusExpired:
		return "expired"
	default:
		return "unknown"
	}
}

// IsProcessed 检查消息是否已处理完成
func (s MessageStatus) IsProcessed() bool {
	return s == MessageStatusProcessed || s == MessageStatusSent
}

// IsFinal 检查消息是否处于最终状态
func (s MessageStatus) IsFinal() bool {
	return s == MessageStatusSent || s == MessageStatusFailed || s == MessageStatusExpired
}

// CanRetry 检查消息是否可以重试
func (s MessageStatus) CanRetry() bool {
	return s == MessageStatusFailed
}

// ConnectionStatus 连接状态
type ConnectionStatus int8

const (
	ConnectionStatusDisconnected ConnectionStatus = 0 // 断开连接
	ConnectionStatusConnected    ConnectionStatus = 1 // 已连接
	ConnectionStatusConnecting   ConnectionStatus = 2 // 连接中
	ConnectionStatusError        ConnectionStatus = 3 // 连接错误
)

// String 返回连接状态的字符串表示
func (s ConnectionStatus) String() string {
	switch s {
	case ConnectionStatusDisconnected:
		return "disconnected"
	case ConnectionStatusConnected:
		return "connected"
	case ConnectionStatusConnecting:
		return "connecting"
	case ConnectionStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// QueueStatus 队列状态
type QueueStatus int8

const (
	QueueStatusPending   QueueStatus = 0 // 待处理
	QueueStatusRunning   QueueStatus = 1 // 运行中
	QueueStatusCompleted QueueStatus = 2 // 已完成
	QueueStatusFailed    QueueStatus = 3 // 处理失败
	QueueStatusRetrying  QueueStatus = 4 // 重试中
	QueueStatusCancelled QueueStatus = 5 // 已取消
)

// String 返回队列状态的字符串表示
func (s QueueStatus) String() string {
	switch s {
	case QueueStatusPending:
		return "pending"
	case QueueStatusRunning:
		return "running"
	case QueueStatusCompleted:
		return "completed"
	case QueueStatusFailed:
		return "failed"
	case QueueStatusRetrying:
		return "retrying"
	case QueueStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// RouteType 路由类型
type RouteType int8

const (
	RouteTypeDirect      RouteType = 1 // 直接路由
	RouteTypeForward     RouteType = 2 // 转发路由
	RouteTypeBroadcast   RouteType = 3 // 广播路由
	RouteTypeConditional RouteType = 4 // 条件路由
)

// String 返回路由类型的字符串表示
func (t RouteType) String() string {
	switch t {
	case RouteTypeDirect:
		return "direct"
	case RouteTypeForward:
		return "forward"
	case RouteTypeBroadcast:
		return "broadcast"
	case RouteTypeConditional:
		return "conditional"
	default:
		return "unknown"
	}
}
