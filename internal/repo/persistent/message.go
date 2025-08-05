package persistent

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/database"
)

// MessageRepo Message相关的数据访问层实现
type MessageRepo struct {
	*BaseRepo
}

// NewMessageRepository 创建Message Repository实例
func NewMessageRepository(db database.CommonDB) port.MessageRepository {
	return &MessageRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的Message
func (r *MessageRepo) Create(ctx context.Context, message *entity.Message) error {
	// 简单的数据验证
	if message.MessageID == "" {
		return fmt.Errorf("消息ID不能为空")
	}
	if message.ChannelID == "" {
		return fmt.Errorf("通道ID不能为空")
	}
	if message.MessageType == "" {
		return fmt.Errorf("消息类型不能为空")
	}

	return r.db.GetGORM().WithContext(ctx).Create(message).Error
}

// GetByID 根据ID获取Message
func (r *MessageRepo) GetByID(ctx context.Context, id string) (*entity.Message, error) {
	var message entity.Message
	err := r.db.GetGORM().WithContext(ctx).
		Preload("Channel").
		Preload("ParentMessage").
		Preload("Replies").
		First(&message, "id = ?", id).Error
	if err != nil {
		return nil, r.handleError(err, "get message by id")
	}
	return &message, nil
}

// GetByMessageID 根据消息ID获取Message
func (r *MessageRepo) GetByMessageID(ctx context.Context, messageID string) (*entity.Message, error) {
	var message entity.Message
	err := r.db.GetGORM().WithContext(ctx).
		Preload("Channel").
		Preload("ParentMessage").
		Preload("Replies").
		Where("message_id = ?", messageID).
		First(&message).Error
	if err != nil {
		return nil, r.handleError(err, "get message by message id")
	}
	return &message, nil
}

// GetByPlatformMessageID 根据平台消息ID获取Message
func (r *MessageRepo) GetByPlatformMessageID(ctx context.Context, channelID string, platformMessageID string) (*entity.Message, error) {
	var message entity.Message
	err := r.db.GetGORM().WithContext(ctx).
		Where("channel_id = ? AND platform_message_id = ?", channelID, platformMessageID).
		First(&message).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 返回nil表示不存在
		}
		return nil, r.handleError(err, "get message by platform message id")
	}
	return &message, nil
}

// GetByChannelID 根据通道ID获取消息列表
func (r *MessageRepo) GetByChannelID(ctx context.Context, channelID string, params port.ListParams) (*port.PaginatedResult[*entity.Message], error) {
	// 转换参数类型为内部使用的类型
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var messages []*entity.Message
	query := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Message{}).
		Where("channel_id = ?", channelID)

	query = r.buildQuery(query, internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &messages)
	if err != nil {
		return nil, r.handleError(err, "get messages by channel ID")
	}

	return result, nil
}

// GetByConversationID 根据会话ID获取消息列表
func (r *MessageRepo) GetByConversationID(ctx context.Context, conversationID string, params port.ListParams) (*port.PaginatedResult[*entity.Message], error) {
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var messages []*entity.Message
	query := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Message{}).
		Where("conversation_id = ?", conversationID)

	query = r.buildQuery(query, internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &messages)
	if err != nil {
		return nil, r.handleError(err, "get messages by conversation ID")
	}

	return result, nil
}

// GetPendingMessages 获取待处理的消息
func (r *MessageRepo) GetPendingMessages(ctx context.Context, limit int) ([]*entity.Message, error) {
	var messages []*entity.Message
	err := r.db.GetGORM().WithContext(ctx).
		Where("message_status = ?", entity.MessageStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, r.handleError(err, "get pending messages")
	}
	return messages, nil
}

// GetFailedMessages 获取失败的消息
func (r *MessageRepo) GetFailedMessages(ctx context.Context, limit int) ([]*entity.Message, error) {
	var messages []*entity.Message
	err := r.db.GetGORM().WithContext(ctx).
		Where("message_status = ?", entity.MessageStatusFailed).
		Order("updated_at DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, r.handleError(err, "get failed messages")
	}
	return messages, nil
}

// Update 更新Message
func (r *MessageRepo) Update(ctx context.Context, message *entity.Message) error {
	// 简单的数据验证
	if message.MessageID == "" {
		return fmt.Errorf("消息ID不能为空")
	}
	if message.ChannelID == "" {
		return fmt.Errorf("通道ID不能为空")
	}
	if message.MessageType == "" {
		return fmt.Errorf("消息类型不能为空")
	}

	return r.db.GetGORM().WithContext(ctx).Save(message).Error
}

// Delete 删除Message（软删除）
func (r *MessageRepo) Delete(ctx context.Context, id string) error {
	return r.softDelete(ctx, &entity.Message{}, id)
}

// List 获取消息列表（分页）
func (r *MessageRepo) List(ctx context.Context, params port.ListParams) (*port.PaginatedResult[*entity.Message], error) {
	internalParams := convertToInternalParams(params)
	internalParams = r.validateParams(internalParams)

	var messages []*entity.Message
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.Message{}), internalParams)

	result, err := PaginateTypedForPort(r.db.GetGORM(), ctx, query, internalParams, &messages)
	if err != nil {
		return nil, r.handleError(err, "list messages")
	}

	return result, nil
}

// UpdateStatus 更新消息状态
func (r *MessageRepo) UpdateStatus(ctx context.Context, id string, status entity.MessageStatus) error {
	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Message{}).
		Where("id = ?", id).
		Update("message_status", status)

	if result.Error != nil {
		return r.handleError(result.Error, "update message status")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// IncrementRetryCount 增加重试次数
func (r *MessageRepo) IncrementRetryCount(ctx context.Context, id string) error {
	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Message{}).
		Where("id = ?", id).
		Update("retry_count", gorm.Expr("retry_count + 1"))

	if result.Error != nil {
		return r.handleError(result.Error, "increment retry count")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAsProcessed 标记消息为已处理
func (r *MessageRepo) MarkAsProcessed(ctx context.Context, id string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"message_status": entity.MessageStatusProcessed,
		"processed_at":   &now,
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Message{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "mark message as processed")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAsSent 标记消息为已发送
func (r *MessageRepo) MarkAsSent(ctx context.Context, id string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"message_status": entity.MessageStatusSent,
		"sent_at":        &now,
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Message{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "mark message as sent")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAsFailed 标记消息为失败
func (r *MessageRepo) MarkAsFailed(ctx context.Context, id string, errorMsg string) error {
	updates := map[string]interface{}{
		"message_status": entity.MessageStatusFailed,
		"error_message":  errorMsg,
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Message{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "mark message as failed")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Exists 检查Message是否存在
func (r *MessageRepo) Exists(ctx context.Context, id string) (bool, error) {
	return r.exists(ctx, &entity.Message{}, "id = ?", id)
}

// ExistsByMessageID 检查指定消息ID的Message是否存在
func (r *MessageRepo) ExistsByMessageID(ctx context.Context, messageID string) (bool, error) {
	return r.exists(ctx, &entity.Message{}, "message_id = ?", messageID)
}

// 确保 MessageRepo 实现了 port.MessageRepository 接口
var _ port.MessageRepository = (*MessageRepo)(nil)
