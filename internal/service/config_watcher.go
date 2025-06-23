package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
	"github.com/sivdead/OmniBotGo/pkg/logger"
)

// ConfigWatcher 配置监视器
type ConfigWatcher struct {
	systemRepo     port.SystemConfigRepository
	channelRepo    port.ChannelRepository
	processorRepo  port.MessageProcessorRepository
	adapterManager port.AdapterManager
	logger         logger.Interface

	// 配置变更回调
	callbacks   map[string][]ConfigChangeCallback
	callbacksMu sync.RWMutex

	// 监控状态
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
	interval time.Duration

	// 配置缓存
	lastConfigs map[string]*entity.SystemConfig
	configsMu   sync.RWMutex
}

// ConfigChangeCallback 配置变更回调函数
type ConfigChangeCallback func(config *entity.SystemConfig) error

// NewConfigWatcher 创建配置监视器
func NewConfigWatcher(
	systemRepo port.SystemConfigRepository,
	channelRepo port.ChannelRepository,
	processorRepo port.MessageProcessorRepository,
	adapterManager port.AdapterManager,
	logger logger.Interface,
) *ConfigWatcher {
	return &ConfigWatcher{
		systemRepo:     systemRepo,
		channelRepo:    channelRepo,
		processorRepo:  processorRepo,
		adapterManager: adapterManager,
		logger:         logger,
		callbacks:      make(map[string][]ConfigChangeCallback),
		lastConfigs:    make(map[string]*entity.SystemConfig),
		interval:       30 * time.Second, // 默认30秒检查一次
	}
}

// RegisterCallback 注册配置变更回调
func (w *ConfigWatcher) RegisterCallback(configKey string, callback ConfigChangeCallback) {
	w.callbacksMu.Lock()
	defer w.callbacksMu.Unlock()

	w.callbacks[configKey] = append(w.callbacks[configKey], callback)
	w.logger.Info("注册配置变更回调", "config_key", configKey)
}

// Start 启动配置监视器
func (w *ConfigWatcher) Start(ctx context.Context) error {
	if w.running {
		return fmt.Errorf("配置监视器已经在运行")
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	w.running = true

	// 初始化配置
	if err := w.loadInitialConfigs(); err != nil {
		w.running = false
		return fmt.Errorf("加载初始配置失败: %w", err)
	}

	// 注册默认的配置处理器
	w.registerDefaultHandlers()

	// 启动监控协程
	go w.watch()

	w.logger.Info("配置监视器已启动", "interval", w.interval)
	return nil
}

// Stop 停止配置监视器
func (w *ConfigWatcher) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.running = false
	w.logger.Info("配置监视器已停止")
}

// loadInitialConfigs 加载初始配置
func (w *ConfigWatcher) loadInitialConfigs() error {
	configs, err := w.systemRepo.List(w.ctx, port.ListParams{
		Page:     1,
		PageSize: 1000, // 获取所有配置
	})
	if err != nil {
		return err
	}

	w.configsMu.Lock()
	defer w.configsMu.Unlock()

	for _, config := range configs.Items {
		w.lastConfigs[config.ConfigKey] = config
	}

	w.logger.Info("加载初始配置完成", "count", len(w.lastConfigs))
	return nil
}

// watch 监控配置变更
func (w *ConfigWatcher) watch() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.checkConfigChanges(); err != nil {
				w.logger.Error("检查配置变更失败", "error", err)
			}
		}
	}
}

// checkConfigChanges 检查配置变更
func (w *ConfigWatcher) checkConfigChanges() error {
	configs, err := w.systemRepo.List(w.ctx, port.ListParams{
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		return err
	}

	// 检查变更
	for _, config := range configs.Items {
		if w.isConfigChanged(config) {
			w.logger.Info("检测到配置变更",
				"config_key", config.ConfigKey,
				"config_group", config.ConfigGroup)

			// 更新缓存
			w.updateConfigCache(config)

			// 触发回调
			if err := w.triggerCallbacks(config); err != nil {
				w.logger.Error("触发配置变更回调失败",
					"config_key", config.ConfigKey,
					"error", err)
			}
		}
	}

	return nil
}

// isConfigChanged 检查配置是否变更
func (w *ConfigWatcher) isConfigChanged(config *entity.SystemConfig) bool {
	w.configsMu.RLock()
	defer w.configsMu.RUnlock()

	lastConfig, exists := w.lastConfigs[config.ConfigKey]
	if !exists {
		return true // 新配置
	}

	// 比较配置值
	if config.ConfigValue != lastConfig.ConfigValue {
		return true
	}

	// 比较更新时间
	if config.UpdatedAt.After(lastConfig.UpdatedAt) {
		return true
	}

	return false
}

// updateConfigCache 更新配置缓存
func (w *ConfigWatcher) updateConfigCache(config *entity.SystemConfig) {
	w.configsMu.Lock()
	defer w.configsMu.Unlock()

	w.lastConfigs[config.ConfigKey] = config
}

// triggerCallbacks 触发配置变更回调
func (w *ConfigWatcher) triggerCallbacks(config *entity.SystemConfig) error {
	w.callbacksMu.RLock()
	callbacks := w.callbacks[config.ConfigKey]
	w.callbacksMu.RUnlock()

	for _, callback := range callbacks {
		if err := callback(config); err != nil {
			return err
		}
	}

	return nil
}

// registerDefaultHandlers 注册默认的配置处理器
func (w *ConfigWatcher) registerDefaultHandlers() {
	// 处理平台配置变更
	w.RegisterCallback("platform.*", func(config *entity.SystemConfig) error {
		return w.handlePlatformConfigChange(config)
	})

	// 处理通道配置变更
	w.RegisterCallback("channel.*", func(config *entity.SystemConfig) error {
		return w.handleChannelConfigChange(config)
	})

	// 处理处理器配置变更
	w.RegisterCallback("processor.*", func(config *entity.SystemConfig) error {
		return w.handleProcessorConfigChange(config)
	})

	// 处理速率限制配置变更
	w.RegisterCallback("rate_limit.*", func(config *entity.SystemConfig) error {
		return w.handleRateLimitConfigChange(config)
	})
}

// handlePlatformConfigChange 处理平台配置变更
func (w *ConfigWatcher) handlePlatformConfigChange(config *entity.SystemConfig) error {
	w.logger.Info("处理平台配置变更", "config", config)

	// 解析配置值
	var platformConfig map[string]interface{}
	if err := json.Unmarshal([]byte(config.ConfigValue), &platformConfig); err != nil {
		return fmt.Errorf("解析平台配置失败: %w", err)
	}

	// TODO: 更新平台适配器配置
	// 例如：更新API密钥、回调URL等

	return nil
}

// handleChannelConfigChange 处理通道配置变更
func (w *ConfigWatcher) handleChannelConfigChange(config *entity.SystemConfig) error {
	w.logger.Info("处理通道配置变更", "config", config)

	// 获取通道ID
	var channelConfig struct {
		ChannelID int64                  `json:"channel_id"`
		Config    map[string]interface{} `json:"config"`
	}

	if err := json.Unmarshal([]byte(config.ConfigValue), &channelConfig); err != nil {
		return fmt.Errorf("解析通道配置失败: %w", err)
	}

	// 获取通道
	channel, err := w.channelRepo.GetByID(w.ctx, channelConfig.ChannelID)
	if err != nil {
		return fmt.Errorf("获取通道失败: %w", err)
	}

	// 更新通道配置
	channel.Config = entity.JSONField(channelConfig.Config)
	if err := w.channelRepo.Update(w.ctx, channel); err != nil {
		return fmt.Errorf("更新通道配置失败: %w", err)
	}

	w.logger.Info("通道配置已更新", "channel_id", channelConfig.ChannelID)
	return nil
}

// handleProcessorConfigChange 处理处理器配置变更
func (w *ConfigWatcher) handleProcessorConfigChange(config *entity.SystemConfig) error {
	w.logger.Info("处理处理器配置变更", "config", config)

	// 获取处理器ID
	var processorConfig struct {
		ProcessorID int64                  `json:"processor_id"`
		Config      map[string]interface{} `json:"config"`
	}

	if err := json.Unmarshal([]byte(config.ConfigValue), &processorConfig); err != nil {
		return fmt.Errorf("解析处理器配置失败: %w", err)
	}

	// 获取处理器
	processor, err := w.processorRepo.GetByID(w.ctx, processorConfig.ProcessorID)
	if err != nil {
		return fmt.Errorf("获取处理器失败: %w", err)
	}

	// 更新处理器配置
	processor.Config = entity.JSONField(processorConfig.Config)
	if err := w.processorRepo.Update(w.ctx, processor); err != nil {
		return fmt.Errorf("更新处理器配置失败: %w", err)
	}

	w.logger.Info("处理器配置已更新", "processor_id", processorConfig.ProcessorID)
	return nil
}

// handleRateLimitConfigChange 处理速率限制配置变更
func (w *ConfigWatcher) handleRateLimitConfigChange(config *entity.SystemConfig) error {
	w.logger.Info("处理速率限制配置变更", "config", config)

	// TODO: 更新速率限制器配置
	// 需要与速率限制中间件集成

	return nil
}

// GetConfig 获取配置值
func (w *ConfigWatcher) GetConfig(configKey string) (*entity.SystemConfig, error) {
	w.configsMu.RLock()
	defer w.configsMu.RUnlock()

	config, exists := w.lastConfigs[configKey]
	if !exists {
		return nil, fmt.Errorf("配置不存在: %s", configKey)
	}

	return config, nil
}

// UpdateConfig 更新配置（会触发热加载）
func (w *ConfigWatcher) UpdateConfig(ctx context.Context, configKey string, configValue string) error {
	// 获取现有配置
	config, err := w.systemRepo.GetByKey(ctx, configKey)
	if err != nil {
		// 如果不存在，创建新配置
		config = &entity.SystemConfig{
			ConfigKey:   configKey,
			ConfigValue: configValue,
			ConfigGroup: "dynamic",
			Description: "动态配置",
		}
		return w.systemRepo.Create(ctx, config)
	}

	// 更新配置值
	config.ConfigValue = configValue
	return w.systemRepo.Update(ctx, config)
}
