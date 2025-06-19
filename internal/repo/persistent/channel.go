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

// ChannelRepo Channel相关的数据访问层实现
type ChannelRepo struct {
	*BaseRepo
}

// NewChannelRepo 创建Channel Repository实例
func NewChannelRepo(db database.CommonDB) repo.ChannelRepo {
	return &ChannelRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的Channel
func (r *ChannelRepo) Create(ctx context.Context, channel *entity.Channel) error {
	if err := channel.Validate(); err != nil {
		return fmt.Errorf("channel validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(channel).Error; err != nil {
		return r.handleError(err, "create channel")
	}
	return nil
}

// GetByID 根据ID获取Channel
func (r *ChannelRepo) GetByID(ctx context.Context, id int64) (*entity.Channel, error) {
	var channel entity.Channel
	err := r.db.GetGORM().WithContext(ctx).
		Preload("Bot").
		Preload("Messages").
		Preload("ConnectionLogs").
		Preload("APICallLogs").
		First(&channel, id).Error
	if err != nil {
		return nil, r.handleError(err, "get channel by id")
	}
	return &channel, nil
}

// GetByBotID 根据Bot ID获取所有Channel
func (r *ChannelRepo) GetByBotID(ctx context.Context, botID int64) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	err := r.db.GetGORM().WithContext(ctx).
		Where("bot_id = ?", botID).
		Find(&channels).Error
	if err != nil {
		return nil, r.handleError(err, "get channels by bot id")
	}
	return channels, nil
}

// GetByPlatformType 根据平台类型获取Channel
func (r *ChannelRepo) GetByPlatformType(ctx context.Context, platformType string) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	err := r.db.GetGORM().WithContext(ctx).
		Where("platform_type = ?", platformType).
		Find(&channels).Error
	if err != nil {
		return nil, r.handleError(err, "get channels by platform type")
	}
	return channels, nil
}

// GetByWebhookPath 根据Webhook路径获取Channel
func (r *ChannelRepo) GetByWebhookPath(ctx context.Context, path string) (*entity.Channel, error) {
	var channel entity.Channel
	err := r.db.GetGORM().WithContext(ctx).
		Where("webhook_path = ?", path).
		First(&channel).Error
	if err != nil {
		return nil, r.handleError(err, "get channel by webhook path")
	}
	return &channel, nil
}

// GetByName 根据机器人ID和通道名称获取Channel
func (r *ChannelRepo) GetByName(ctx context.Context, botID int64, channelName string) (*entity.Channel, error) {
	var channel entity.Channel
	err := r.db.GetGORM().WithContext(ctx).
		Where("bot_id = ? AND channel_name = ?", botID, channelName).
		First(&channel).Error
	if err != nil {
		return nil, r.handleError(err, "get channel by name")
	}
	return &channel, nil
}

// Update 更新Channel
func (r *ChannelRepo) Update(ctx context.Context, channel *entity.Channel) error {
	if err := channel.Validate(); err != nil {
		return fmt.Errorf("channel validation failed: %w", err)
	}

	result := r.db.GetGORM().WithContext(ctx).Save(channel)
	if result.Error != nil {
		return r.handleError(result.Error, "update channel")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除Channel（软删除）
func (r *ChannelRepo) Delete(ctx context.Context, id int64) error {
	return r.softDelete(ctx, &entity.Channel{}, id)
}

// List 获取Channel列表（分页）
func (r *ChannelRepo) List(ctx context.Context, params repo.ListParams) (*repo.PaginatedResult[*entity.Channel], error) {
	params = r.validateParams(params)

	var channels []*entity.Channel
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.Channel{}), params)

	result, err := PaginateTyped(r.db.GetGORM(), ctx, query, params, &channels)
	if err != nil {
		return nil, r.handleError(err, "list channels")
	}

	return result, nil
}

// ListActive 获取所有激活状态的Channel
func (r *ChannelRepo) ListActive(ctx context.Context) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	err := r.db.GetGORM().WithContext(ctx).
		Where("status = ?", entity.StatusActive).
		Find(&channels).Error
	if err != nil {
		return nil, r.handleError(err, "list active channels")
	}
	return channels, nil
}

// UpdateConnectionStatus 更新连接状态
func (r *ChannelRepo) UpdateConnectionStatus(ctx context.Context, id int64, status entity.ConnectionStatus) error {
	updates := map[string]interface{}{
		"connection_status": status,
	}

	if status == entity.ConnectionStatusConnected {
		updates["last_connected_at"] = time.Now()
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Channel{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "update connection status")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdateAccessToken 更新访问令牌
func (r *ChannelRepo) UpdateAccessToken(ctx context.Context, id int64, token string, expiresAt *time.Time) error {
	updates := map[string]interface{}{
		"access_token": token,
	}

	if expiresAt != nil {
		updates["access_token_expires_at"] = expiresAt
	}

	result := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Channel{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return r.handleError(result.Error, "update access token")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Exists 检查Channel是否存在
func (r *ChannelRepo) Exists(ctx context.Context, id int64) (bool, error) {
	return r.exists(ctx, &entity.Channel{}, "id = ?", id)
}

// GetPendingMessageCount 获取通道的待处理消息数量
func (r *ChannelRepo) GetPendingMessageCount(ctx context.Context, channelID int64) (int64, error) {
	var count int64
	err := r.db.GetGORM().WithContext(ctx).
		Model(&entity.Message{}).
		Where("channel_id = ? AND message_status = ?", channelID, entity.MessageStatusPending).
		Count(&count).Error
	if err != nil {
		return 0, r.handleError(err, "get pending message count")
	}
	return count, nil
}
