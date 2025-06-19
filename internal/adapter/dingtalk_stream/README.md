# 钉钉Stream适配器

钉钉适配器实现了与钉钉平台的集成，支持Stream模式的消息收发。

## 架构设计

钉钉适配器包含以下主要组件：

1. **Adapter** - 实现`PlatformAdapter`接口，提供平台标准功能
2. **StreamController** - 管理钉钉Stream连接，处理实时消息
3. **Model** - 钉钉相关的数据结构定义

## 主要功能

### Adapter (adapter.go)
- ✅ 实现PlatformAdapter接口
- ✅ 获取和刷新访问令牌
- ✅ 解析入站消息
- ✅ 发送出站消息
- ✅ 支持多种消息类型（文本、Markdown、ActionCard等）

### StreamController (stream_controller.go)
- ✅ 基于钉钉Stream SDK的实时消息接收
- ✅ 自动消息格式转换
- ✅ 集成到现有UseCase层
- ✅ 支持优雅启动和停止

## 支持的消息类型

### 接收消息
- 文本消息 (`text`)
- 图片消息 (`picture`) 
- 音频消息 (`audio`)
- 视频消息 (`video`)
- 文件消息 (`file`)
- 交互式消息 (`interactive`)

### 发送消息
- 文本消息 (`text`)
- Markdown消息 (`markdown`)
- ActionCard消息 (`actionCard`)

## 使用方法

### 1. 通过AdapterManager使用

```go
// 获取钉钉适配器
adapterManager := adapter.NewManager()
dingtalkAdapter, err := adapterManager.GetAdapter(entity.PlatformTypeDingtalk)
if err != nil {
    log.Fatal("Failed to get dingtalk adapter:", err)
}

// 配置参数
config := map[string]interface{}{
    "app_key":       "your_app_key",
    "app_secret":    "your_app_secret",
    "client_id":     "your_client_id", 
    "client_secret": "your_client_secret",
}

// 验证配置
err = dingtalkAdapter.ValidateConfig(config)
if err != nil {
    log.Fatal("Invalid config:", err)
}
```

### 2. Stream模式集成

```go
// 创建Stream控制器（通过接口注入消息处理器）
streamController := NewStreamController(
    messageUsecase, // 实现MessageHandler接口的任何实例
    logger,
    &DingtalkStreamConfig{
        AppKey:       "your_app_key",
        AppSecret:    "your_app_secret",
        ClientID:     "your_client_id",
        ClientSecret: "your_client_secret",
    },
    channelID,
)

// 启动Stream客户端
ctx := context.Background()
err := streamController.Start(ctx)
if err != nil {
    log.Fatal("Failed to start stream controller:", err)
}

// 优雅关闭
defer streamController.Stop()
```

## 架构集成

钉钉适配器完全集成到现有架构中：

1. **标准接口**：实现`PlatformAdapter`接口，与其他平台适配器保持一致
2. **依赖倒置**：StreamController通过`MessageHandler`接口处理业务逻辑，避免循环依赖
3. **统一管理**：通过`AdapterManager`统一管理所有平台适配器
4. **配置统一**：使用标准的配置格式和验证机制

### MessageHandler接口

为了避免循环依赖，StreamController使用接口依赖倒置：

```go
type MessageHandler interface {
    ProcessInboundMessage(ctx context.Context, message *entity.Message) error
}
```

任何实现了这个接口的类型都可以作为消息处理器注入到StreamController中。

## 配置说明

### 基本配置
```go
type DingtalkStreamConfig struct {
    AppKey         string `json:"app_key" validate:"required"`
    AppSecret      string `json:"app_secret" validate:"required"`
    ClientID       string `json:"client_id" validate:"required"`
    ClientSecret   string `json:"client_secret" validate:"required"`
    SubscriptionID string `json:"subscription_id,omitempty"`
    Topic          string `json:"topic,omitempty"`
}
```

### 获取配置参数

1. 登录钉钉开放平台
2. 创建企业应用
3. 获取`AppKey`和`AppSecret`
4. 配置Stream推送地址
5. 获取`ClientID`和`ClientSecret`

## 扩展说明

如果需要添加其他钉钉功能：

1. 在`model.go`中添加相关数据结构
2. 在`Adapter`中实现相关方法
3. 在`StreamController`中添加相应的消息处理逻辑
4. 更新配置结构和验证逻辑 

## 实际集成示例

### 在应用层集成StreamController

```go
package main

import (
    "context"
    "log" 
    "os"
    "os/signal"
    "syscall"

    "github.com/rs/zerolog"
    "github.com/sivdead/OmniBotGo/internal/adapter/dingtalk_stream"
    "github.com/sivdead/OmniBotGo/internal/usecase"
    "github.com/sivdead/OmniBotGo/internal/entity"
)

func main() {
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    
    // 创建MessageUseCase实例
    messageUsecase := usecase.NewMessageUseCase(
        messageRepo,
        adapterManager,
        logger,
    )
    
    // 钉钉Stream配置
    config := &dingtalk_stream.DingtalkStreamConfig{
        AppKey:       os.Getenv("DINGTALK_APP_KEY"),
        AppSecret:    os.Getenv("DINGTALK_APP_SECRET"),
        ClientID:     os.Getenv("DINGTALK_CLIENT_ID"),
        ClientSecret: os.Getenv("DINGTALK_CLIENT_SECRET"),
    }
    
    // 创建Stream控制器
    streamController := dingtalk_stream.NewStreamController(
        messageUsecase, // messageUsecase实现了MessageHandler接口
        logger,
        config,
        channelID,
    )
    
    // 启动Stream客户端
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    err := streamController.Start(ctx)
    if err != nil {
        log.Fatal("Failed to start dingtalk stream controller:", err)
    }
    
    // 优雅关闭
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
    
    <-stop
    logger.Info().Msg("Shutting down...")
    
    streamController.Stop()
}
```

### 自定义MessageHandler

如果需要自定义消息处理逻辑，可以实现MessageHandler接口：

```go
type CustomMessageHandler struct {
    logger zerolog.Logger
    // 其他依赖
}

func (h *CustomMessageHandler) ProcessInboundMessage(ctx context.Context, message *entity.Message) error {
    h.logger.Info().
        Str("message_id", message.MessageID).
        Str("content", message.Content).
        Msg("processing custom message")
    
    // 自定义处理逻辑
    switch message.MessageType {
    case "text":
        return h.processTextMessage(ctx, message)
    case "image":
        return h.processImageMessage(ctx, message)
    default:
        return h.processDefaultMessage(ctx, message)
    }
}

func (h *CustomMessageHandler) processTextMessage(ctx context.Context, message *entity.Message) error {
    // 处理文本消息
    return nil
}

func (h *CustomMessageHandler) processImageMessage(ctx context.Context, message *entity.Message) error {
    // 处理图片消息
    return nil
}

func (h *CustomMessageHandler) processDefaultMessage(ctx context.Context, message *entity.Message) error {
    // 处理其他类型消息
    return nil
}
``` 