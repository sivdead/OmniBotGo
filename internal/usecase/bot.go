package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/repo"
)

// botUseCase 机器人管理业务逻辑实现
type botUseCase struct {
	botRepo     repo.BotRepo
	channelRepo repo.ChannelRepo
	logger      *zerolog.Logger
}

// NewBotUseCase 创建机器人管理业务逻辑实例
func NewBotUseCase(
	botRepo repo.BotRepo,
	channelRepo repo.ChannelRepo,
	logger *zerolog.Logger,
) BotUseCase {
	return &botUseCase{
		botRepo:     botRepo,
		channelRepo: channelRepo,
		logger:      logger,
	}
}

// CreateBot 创建机器人
func (uc *botUseCase) CreateBot(ctx context.Context, req CreateBotRequest) (*entity.Bot, error) {
	log := uc.logger.With().
		Str("method", "CreateBot").
		Str("bot_name", req.BotName).
		Str("bot_type", req.BotType).
		Str("created_by", req.CreatedBy).
		Logger()

	log.Info().Msg("开始创建机器人")

	// 检查机器人名称是否重复
	existingBot, err := uc.botRepo.GetByName(ctx, req.BotName)
	if err == nil && existingBot != nil {
		log.Warn().Msg("机器人名称已存在")
		return nil, fmt.Errorf("机器人名称已存在")
	}

	// 创建机器人实体
	bot := &entity.Bot{
		BotName:     req.BotName,
		BotType:     req.BotType,
		Description: req.Description,
		AvatarURL:   req.AvatarURL,
		Status:      entity.StatusActive,
		CreatedBy:   req.CreatedBy,
	}

	// 设置基础信息
	now := time.Now()
	bot.CreatedAt = now
	bot.UpdatedAt = now

	// 设置配置信息
	if req.Config != nil {
		bot.Config = entity.JSONField(req.Config)
	}

	// 验证机器人数据
	if err := bot.Validate(); err != nil {
		log.Error().Err(err).Msg("机器人数据验证失败")
		return nil, fmt.Errorf("机器人数据验证失败: %w", err)
	}

	// 保存机器人到数据库
	if err := uc.botRepo.Create(ctx, bot); err != nil {
		log.Error().Err(err).Msg("保存机器人失败")
		return nil, fmt.Errorf("保存机器人失败: %w", err)
	}

	log.Info().
		Int64("bot_id", bot.ID).
		Msg("机器人创建成功")

	return bot, nil
}

// UpdateBot 更新机器人
func (uc *botUseCase) UpdateBot(ctx context.Context, req UpdateBotRequest) (*entity.Bot, error) {
	log := uc.logger.With().
		Str("method", "UpdateBot").
		Int64("bot_id", req.ID).
		Logger()

	log.Info().Msg("开始更新机器人")

	// 获取现有机器人
	bot, err := uc.botRepo.GetByID(ctx, req.ID)
	if err != nil {
		log.Error().Err(err).Msg("获取机器人信息失败")
		return nil, fmt.Errorf("获取机器人信息失败: %w", err)
	}

	// 更新机器人名称
	if req.BotName != nil {
		// 检查新名称是否重复
		if *req.BotName != bot.BotName {
			existingBot, err := uc.botRepo.GetByName(ctx, *req.BotName)
			if err == nil && existingBot != nil {
				log.Warn().Msg("机器人名称已存在")
				return nil, fmt.Errorf("机器人名称已存在")
			}
		}
		bot.BotName = *req.BotName
	}

	// 更新描述
	if req.Description != nil {
		bot.Description = *req.Description
	}

	// 更新头像URL
	if req.AvatarURL != nil {
		bot.AvatarURL = *req.AvatarURL
	}

	// 更新配置信息
	if req.Config != nil {
		if bot.Config == nil {
			bot.Config = make(entity.JSONField)
		}
		for key, value := range req.Config {
			bot.Config.Set(key, value)
		}
	}

	// 更新状态
	if req.Status != nil {
		bot.Status = *req.Status
	}

	// 更新时间戳
	bot.UpdatedAt = time.Now()

	// 验证机器人数据
	if err := bot.Validate(); err != nil {
		log.Error().Err(err).Msg("机器人数据验证失败")
		return nil, fmt.Errorf("机器人数据验证失败: %w", err)
	}

	// 保存更新
	if err := uc.botRepo.Update(ctx, bot); err != nil {
		log.Error().Err(err).Msg("更新机器人失败")
		return nil, fmt.Errorf("更新机器人失败: %w", err)
	}

	log.Info().Msg("机器人更新成功")
	return bot, nil
}

// DeleteBot 删除机器人
func (uc *botUseCase) DeleteBot(ctx context.Context, id int64) error {
	log := uc.logger.With().
		Str("method", "DeleteBot").
		Int64("bot_id", id).
		Logger()

	log.Info().Msg("开始删除机器人")

	// 获取机器人信息
	bot, err := uc.botRepo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("获取机器人信息失败")
		return fmt.Errorf("获取机器人信息失败: %w", err)
	}

	// 检查机器人是否有活跃的通道
	activeChannelCount, err := uc.botRepo.GetActiveChannelCount(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("检查活跃通道失败")
		return fmt.Errorf("检查活跃通道失败: %w", err)
	}

	if activeChannelCount > 0 {
		log.Warn().Int64("active_channel_count", activeChannelCount).Msg("机器人存在活跃通道，无法删除")
		return fmt.Errorf("机器人存在 %d 个活跃通道，无法删除", activeChannelCount)
	}

	// 软删除机器人（标记为已删除状态）
	bot.Status = entity.StatusDeleted
	bot.UpdatedAt = time.Now()

	if err := uc.botRepo.Update(ctx, bot); err != nil {
		log.Error().Err(err).Msg("删除机器人失败")
		return fmt.Errorf("删除机器人失败: %w", err)
	}

	log.Info().Msg("机器人删除成功")
	return nil
}

// GetBot 获取机器人信息
func (uc *botUseCase) GetBot(ctx context.Context, id int64) (*entity.Bot, error) {
	log := uc.logger.With().
		Str("method", "GetBot").
		Int64("bot_id", id).
		Logger()

	log.Info().Msg("获取机器人信息")

	bot, err := uc.botRepo.GetByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Msg("获取机器人信息失败")
		return nil, fmt.Errorf("获取机器人信息失败: %w", err)
	}

	log.Info().
		Str("bot_name", bot.BotName).
		Str("bot_type", bot.BotType).
		Str("status", bot.Status.String()).
		Str("created_by", bot.CreatedBy).
		Msg("机器人信息获取成功")

	return bot, nil
}

// ListBots 获取机器人列表
func (uc *botUseCase) ListBots(ctx context.Context, params ListBotsParams) (*ListResult[entity.Bot], error) {
	log := uc.logger.With().
		Str("method", "ListBots").
		Logger()

	log.Info().Msg("获取机器人列表")

	// 构建查询过滤器
	filters := make(map[string]interface{})

	if params.BotType != nil {
		filters["bot_type"] = *params.BotType
	}
	if params.Status != nil {
		filters["status"] = *params.Status
	}
	if params.CreatedBy != nil {
		filters["created_by"] = *params.CreatedBy
	}
	// 查询机器人列表
	paginatedResult, err := uc.botRepo.List(ctx, repo.ListParams{
		Filters:  filters,
		Page:     params.Page,
		PageSize: params.PageSize,
		OrderBy:  "created_by",
	})
	if err != nil {
		log.Error().Err(err).Msg("查询机器人列表失败")
		return nil, fmt.Errorf("查询机器人列表失败: %w", err)
	}

	// 计算总页数
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	result := &BotListResult{
		Items:      bots,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}

	log.Info().
		Int64("total", total).
		Int("page", params.Page).
		Int("page_size", params.PageSize).
		Msg("机器人列表查询完成")

	return result, nil
}
