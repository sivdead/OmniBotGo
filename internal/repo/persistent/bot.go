package persistent

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
	"github.com/sivdead/OmniBotGo/pkg/database"
)

// BotRepo Bot相关的数据访问层实现
type BotRepo struct {
	*BaseRepo
}

func (r *BotRepo) GetActiveChannelCount(ctx context.Context, id int64) (int64, error) {

	var count int64
	r.db.GetGORM().WithContext(ctx).Model(&entity.Channel{}).Where("bot_id = ? AND status = ?", id, entity.StatusActive).Count(&count)
	return count, nil
}

// NewBotRepo 创建Bot Repository实例
func NewBotRepo(db database.CommonDB) repo.BotRepo {
	return &BotRepo{
		BaseRepo: NewBaseRepo(db),
	}
}

// Create 创建新的Bot
func (r *BotRepo) Create(ctx context.Context, bot *entity.Bot) error {
	if err := bot.Validate(); err != nil {
		return fmt.Errorf("bot validation failed: %w", err)
	}

	if err := r.db.GetGORM().WithContext(ctx).Create(bot).Error; err != nil {
		return r.handleError(err, "create bot")
	}
	return nil
}

// GetByID 根据ID获取Bot
func (r *BotRepo) GetByID(ctx context.Context, id int64) (*entity.Bot, error) {
	var bot entity.Bot
	err := r.db.GetGORM().WithContext(ctx).Preload("Channels").First(&bot, id).Error
	if err != nil {
		return nil, r.handleError(err, "get bot by id")
	}
	return &bot, nil
}

// GetByName 根据名称获取Bot
func (r *BotRepo) GetByName(ctx context.Context, name string) (*entity.Bot, error) {
	var bot entity.Bot
	err := r.db.GetGORM().WithContext(ctx).Preload("Channels").Where("bot_name = ?", name).First(&bot).Error
	if err != nil {
		return nil, r.handleError(err, "get bot by name")
	}
	return &bot, nil
}

// Update 更新Bot
func (r *BotRepo) Update(ctx context.Context, bot *entity.Bot) error {
	if err := bot.Validate(); err != nil {
		return fmt.Errorf("bot validation failed: %w", err)
	}

	result := r.db.GetGORM().WithContext(ctx).Save(bot)
	if result.Error != nil {
		return r.handleError(result.Error, "update bot")
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除Bot（软删除）
func (r *BotRepo) Delete(ctx context.Context, id int64) error {
	return r.softDelete(ctx, &entity.Bot{}, id)
}

// List 获取Bot列表（分页）
func (r *BotRepo) List(ctx context.Context, params repo.ListParams) (*repo.PaginatedResult[*entity.Bot], error) {
	params = r.validateParams(params)

	var bots []*entity.Bot
	query := r.buildQuery(r.db.GetGORM().WithContext(ctx).Model(&entity.Bot{}), params)

	result, err := PaginateTyped(r.db.GetGORM(), ctx, query, params, &bots)
	if err != nil {
		return nil, r.handleError(err, "list bots")
	}

	return result, nil
}

// ListActive 获取所有激活状态的Bot
func (r *BotRepo) ListActive(ctx context.Context) ([]*entity.Bot, error) {
	var bots []*entity.Bot
	err := r.db.GetGORM().WithContext(ctx).
		Where("status = ?", entity.StatusActive).
		Find(&bots).Error
	if err != nil {
		return nil, r.handleError(err, "list active bots")
	}
	return bots, nil
}

// Exists 检查Bot是否存在
func (r *BotRepo) Exists(ctx context.Context, id int64) (bool, error) {
	return r.exists(ctx, &entity.Bot{}, "id = ?", id)
}

// ExistsByName 检查指定名称的Bot是否存在
func (r *BotRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	return r.exists(ctx, &entity.Bot{}, "bot_name = ?", name)
}
