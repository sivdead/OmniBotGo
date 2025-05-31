package persistent

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/pkg/database"
)

// MessageRepo Message相关的数据访问层实现
type MessageRepo struct {
	*BaseRepo
}

// NewMessageRepo 创建Message Repository实例
func NewMessageRepo(db database.CommonDB) repo.MessageRepo {
	return &MessageRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的Message
func (r *MessageRepo) Create(ctx context.Context, message *entity.Message) error {
	if err := message.Validate(); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(message).Error; err != nil {
		return r.handleError(err, "create message")
	}
	return nil
}

// GetByID 根据ID获取Message
func (r *MessageRepo) GetByID(ctx context.Context, id int64) (*entity.Message, error) {
	var message entity.Message
	err := r.db.GetGORM().WithContext(ctx).
		Preload("Channel").
		Preload("ParentMessage").
		Preload("Replies").
		First(&message, id).Error
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

// GetByChannelID 根据通道ID获取Message列表（分页）
func (r *MessageRepo) GetByChannelID(ctx context.Context, channelID int64, params repo.ListParams) (*repo.PaginatedResult, error) {
	params = r.validateParams(params)

	var messages []*entity.Message
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.Message{}).Where("channel_id = ?", channelID), params)

	result, err := r.paginate(ctx, query, params, &messages)
	if err != nil {
		return nil, r.handleError(err, "get messages by channel id")
	}

	result.Items = messages
	return result, nil
}

// GetByConversationID 根据会话ID获取Message列表（分页）
func (r *MessageRepo) GetByConversationID(ctx context.Context, conversationID string, params repo.ListParams) (*repo.PaginatedResult, error) {
	params = r.validateParams(params)

	var messages []*entity.Message
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.Message{}).Where("conversation_id = ?", conversationID), params)

	result, err := r.paginate(ctx, query, params, &messages)
	if err != nil {
		return nil, r.handleError(err, "get messages by conversation id")
	}

	result.Items = messages
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
	if err := message.Validate(); err != nil {
		return fmt.Errorf("message validation failed: %w", err)
	}

	result := r.db.GetGORM().WithContext(ctx).Save(message)
	if result.Error != nil {
		return r.handleError(result.Error, "update message")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除Message（软删除）
func (r *MessageRepo) Delete(ctx context.Context, id int64) error {
	return r.softDelete(ctx, &entity.Message{}, id)
}

// List 获取Message列表（分页）
func (r *MessageRepo) List(ctx context.Context, params repo.ListParams) (*repo.PaginatedResult, error) {
	params = r.validateParams(params)

	var messages []*entity.Message
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.Message{}), params)

	result, err := r.paginate(ctx, query, params, &messages)
	if err != nil {
		return nil, r.handleError(err, "list messages")
	}

	result.Items = messages
	return result, nil
}

// UpdateStatus 更新消息状态
func (r *MessageRepo) UpdateStatus(ctx context.Context, id int64, status entity.MessageStatus) error {
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
func (r *MessageRepo) IncrementRetryCount(ctx context.Context, id int64) error {
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
func (r *MessageRepo) MarkAsProcessed(ctx context.Context, id int64) error {
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
func (r *MessageRepo) MarkAsSent(ctx context.Context, id int64) error {
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
func (r *MessageRepo) MarkAsFailed(ctx context.Context, id int64, errorMsg string) error {
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
func (r *MessageRepo) Exists(ctx context.Context, id int64) (bool, error) {
	return r.exists(ctx, &entity.Message{}, "id = ?", id)
}

// ExistsByMessageID 检查指定消息ID的Message是否存在
func (r *MessageRepo) ExistsByMessageID(ctx context.Context, messageID string) (bool, error) {
	return r.exists(ctx, &entity.Message{}, "message_id = ?", messageID)
}
