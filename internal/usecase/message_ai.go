package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/sivdead/OmniBotGo/internal/entity"
	"github.com/sivdead/OmniBotGo/internal/usecase/port"
)

// AI chat processing (T03/T06)

// processAIChat 处理AI聊天功能.
//
//nolint:gocognit,gocyclo,cyclop,funlen // AI chat orchestration retained from pre-split message.go
func (uc *messageUseCase) processAIChat(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor) error {
	// 解析AI配置
	var aiConfig struct {
		Provider     string                   `json:"provider"`           // AI供应商：openai, claude, gemini
		Model        string                   `json:"model"`              // 模型名称
		SystemPrompt string                   `json:"system_prompt"`      // 系统提示词
		Temperature  float64                  `json:"temperature"`        // 温度参数
		MaxTokens    int                      `json:"max_tokens"`         // 最大token数
		APIKey       string                   `json:"api_key"`            // API密钥
		BaseURL      string                   `json:"base_url,omitempty"` // 自定义API地址
		EnableTools  bool                     `json:"enable_tools"`       // 是否启用工具调用
		Tools        []map[string]interface{} `json:"tools,omitempty"`    // 可用工具列表
		StreamMode   bool                     `json:"stream_mode"`        // 是否启用流式响应
	}

	// 将JSONField转换为字节数组
	configBytes, err := json.Marshal(processor.Config)
	if err != nil {
		uc.logger.Error("Failed to marshal processor config", "error", err)

		return fmt.Errorf("invalid processor config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &aiConfig); err != nil {
		uc.logger.Error("Failed to parse AI config", "error", err)

		return fmt.Errorf("invalid AI config: %w", err)
	}

	// 验证必要参数
	if aiConfig.Provider == "" {
		return errAIProviderRequired
	}

	if aiConfig.Model == "" {
		return errAIModelRequired
	}

	if aiConfig.APIKey == "" {
		return errAIAPIKeyRequired
	}

	// 设置默认值
	if aiConfig.Temperature == 0 {
		aiConfig.Temperature = 0.7
	}

	if aiConfig.MaxTokens == 0 {
		aiConfig.MaxTokens = 1024
	}

	if aiConfig.SystemPrompt == "" {
		aiConfig.SystemPrompt = "You are a helpful assistant."
	}

	// 获取会话历史（T06：按 channel_id + conversation_id 隔离）
	history, err := uc.getConversationHistory(ctx, msg.ChannelID, msg.ConversationID, defaultAIHistoryLimit)
	if err != nil {
		uc.logger.Error("Failed to get conversation history", "error", err)

		return fmt.Errorf("failed to get conversation history: %w", err)
	}

	// 构建消息列表
	messages := []*schema.Message{
		schema.SystemMessage(aiConfig.SystemPrompt),
	}

	// 添加历史消息
	for _, historyMsg := range history {
		if historyMsg.ID == msg.ID {
			continue // 跳过当前消息
		}

		if historyMsg.Direction == entity.MessageDirectionInbound {
			messages = append(messages, schema.UserMessage(historyMsg.Content))
		} else {
			// 检查是否是工具调用响应
			if toolCallData := uc.extractToolCallFromMessage(historyMsg); toolCallData != nil {
				messages = append(messages, schema.AssistantMessage(historyMsg.Content, toolCallData))
			} else {
				messages = append(messages, schema.AssistantMessage(historyMsg.Content, []schema.ToolCall{}))
			}
		}
	}

	// 添加当前用户消息
	messages = append(messages, schema.UserMessage(msg.Content))

	// 通过AdapterManager获取AI模型
	aiModelManager := uc.adapterManager.GetAIModelManager()

	// 将配置转换为map格式
	configMap := map[string]interface{}{
		"provider":          aiConfig.Provider,
		"model":             aiConfig.Model,
		"system_prompt":     aiConfig.SystemPrompt,
		"temperature":       aiConfig.Temperature,
		"max_tokens":        aiConfig.MaxTokens,
		"api_key":           aiConfig.APIKey,
		"base_url":          aiConfig.BaseURL,
		"enable_tools":      aiConfig.EnableTools,
		"tools":             aiConfig.Tools,
		configKeyStreamMode: aiConfig.StreamMode,
	}

	chatModelInterface, err := aiModelManager.CreateChatModel(ctx, aiConfig.Provider, configMap)
	if err != nil {
		uc.logger.Error("Failed to create chat model", "error", err)

		return fmt.Errorf("failed to create chat model: %w", err)
	}

	// 类型断言为ToolCallingChatModel
	chatModel, ok := chatModelInterface.(model.ToolCallingChatModel)
	if !ok {
		return errAIModelNotToolCalling
	}

	// 根据配置选择处理方式
	if aiConfig.StreamMode {
		return uc.processAIChatStream(ctx, msg, processor, chatModel, messages, &aiConfig)
	}

	return uc.processAIChatSync(ctx, msg, processor, chatModel, messages, &aiConfig)
}

// processAIChatSync 同步处理AI聊天.
func (uc *messageUseCase) processAIChatSync(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor, chatModel model.ToolCallingChatModel, messages []*schema.Message, aiConfig *struct {
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"system_prompt"`
	Temperature  float64                  `json:"temperature"`
	MaxTokens    int                      `json:"max_tokens"`
	APIKey       string                   `json:"api_key"`
	BaseURL      string                   `json:"base_url,omitempty"`
	EnableTools  bool                     `json:"enable_tools"`
	Tools        []map[string]interface{} `json:"tools,omitempty"`
	StreamMode   bool                     `json:"stream_mode"`
},
) error {
	// 调用AI模型
	response, err := chatModel.Generate(ctx, messages)
	if err != nil {
		uc.logger.Error("Failed to call AI model", "error", err)

		return fmt.Errorf("failed to call AI model: %w", err)
	}

	// 处理工具调用
	if aiConfig.EnableTools && len(response.ToolCalls) > 0 {
		return uc.handleToolCalls(ctx, msg, processor, chatModel, messages, response, aiConfig)
	}

	// 获取AI响应内容
	aiResponse := response.Content
	if aiResponse == "" {
		return errAIEmptyResponse
	}

	// 创建AI回复消息
	replyMsg := uc.createAIReplyMessage(msg, aiResponse, aiConfig, processor, nil)

	// 保存AI回复消息
	if err := uc.messageRepo.Create(ctx, replyMsg); err != nil {
		uc.logger.Error("Failed to save AI reply message", "error", err)

		return fmt.Errorf("failed to save AI reply message: %w", err)
	}

	uc.logger.Info("AI chat processed successfully",
		"provider", aiConfig.Provider,
		"model", aiConfig.Model,
		"message_id", msg.ID,
		"reply_id", replyMsg.ID,
		"response_length", len(aiResponse))

	return nil
}

// processAIChatStream 流式处理AI聊天.
//
//nolint:gocognit,gocyclo,cyclop,funlen // stream path retained from pre-split message.go
func (uc *messageUseCase) processAIChatStream(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor, chatModel model.ToolCallingChatModel, messages []*schema.Message, aiConfig *struct {
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"system_prompt"`
	Temperature  float64                  `json:"temperature"`
	MaxTokens    int                      `json:"max_tokens"`
	APIKey       string                   `json:"api_key"`
	BaseURL      string                   `json:"base_url,omitempty"`
	EnableTools  bool                     `json:"enable_tools"`
	Tools        []map[string]interface{} `json:"tools,omitempty"`
	StreamMode   bool                     `json:"stream_mode"`
},
) error {
	// 调用流式AI模型
	streamReader, err := chatModel.Stream(ctx, messages)
	if err != nil {
		uc.logger.Error("Failed to call AI model stream", "error", err)

		return fmt.Errorf("failed to call AI model stream: %w", err)
	}
	defer streamReader.Close()

	var fullResponse strings.Builder

	var toolCalls []schema.ToolCall

	// 读取流式响应
	for {
		chunk, err := streamReader.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}

			uc.logger.Error("Failed to read stream chunk", "error", err)

			return fmt.Errorf("failed to read stream chunk: %w", err)
		}

		// 累积响应内容
		if chunk.Content != "" {
			fullResponse.WriteString(chunk.Content)
		}

		// 收集工具调用
		if len(chunk.ToolCalls) > 0 {
			toolCalls = append(toolCalls, chunk.ToolCalls...)
		}

		// Real-time client push (WebSocket/SSE) can be added later.
		uc.logger.Debug("Received stream chunk",
			"content", chunk.Content,
			"tool_calls", len(chunk.ToolCalls))
	}

	finalResponse := fullResponse.String()
	if finalResponse == "" && len(toolCalls) == 0 {
		return errAIEmptyStreamResponse
	}

	// 处理工具调用
	if aiConfig.EnableTools && len(toolCalls) > 0 {
		// 创建带工具调用的响应
		response := &schema.Message{
			Content:   finalResponse,
			ToolCalls: toolCalls,
		}

		return uc.handleToolCalls(ctx, msg, processor, chatModel, messages, response, aiConfig)
	}

	// 创建AI回复消息
	replyMsg := uc.createAIReplyMessage(msg, finalResponse, aiConfig, processor, nil)

	// 保存AI回复消息
	if err := uc.messageRepo.Create(ctx, replyMsg); err != nil {
		uc.logger.Error("Failed to save AI reply message", "error", err)

		return fmt.Errorf("failed to save AI reply message: %w", err)
	}

	uc.logger.Info("AI chat stream processed successfully",
		"provider", aiConfig.Provider,
		"model", aiConfig.Model,
		"message_id", msg.ID,
		"reply_id", replyMsg.ID,
		"response_length", len(finalResponse))

	return nil
}

// handleToolCalls 处理工具调用.
func (uc *messageUseCase) handleToolCalls(ctx context.Context, msg *entity.Message, processor *entity.MessageProcessor, chatModel model.ToolCallingChatModel, messages []*schema.Message, response *schema.Message, aiConfig *struct {
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"system_prompt"`
	Temperature  float64                  `json:"temperature"`
	MaxTokens    int                      `json:"max_tokens"`
	APIKey       string                   `json:"api_key"`
	BaseURL      string                   `json:"base_url,omitempty"`
	EnableTools  bool                     `json:"enable_tools"`
	Tools        []map[string]interface{} `json:"tools,omitempty"`
	StreamMode   bool                     `json:"stream_mode"`
},
) error {
	uc.logger.Info("Processing tool calls", "count", len(response.ToolCalls))

	// 添加助手的工具调用消息
	messages = append(messages, response)

	// 执行工具调用
	for _, toolCall := range response.ToolCalls {
		toolResult, err := uc.executeToolCall(ctx, &toolCall)
		if err != nil {
			uc.logger.Error("Failed to execute tool call", "tool", toolCall.Function.Name, "error", err)
			toolResult = fmt.Sprintf("Error executing tool %s: %v", toolCall.Function.Name, err)
		}

		// 添加工具结果消息
		messages = append(messages, schema.ToolMessage(toolResult, toolCall.ID))
	}

	// 再次调用AI模型获取最终响应
	finalResponse, err := chatModel.Generate(ctx, messages)
	if err != nil {
		uc.logger.Error("Failed to get final response after tool calls", "error", err)

		return fmt.Errorf("failed to get final response after tool calls: %w", err)
	}

	// 创建包含工具调用信息的回复消息
	toolCallsData := make([]map[string]interface{}, len(response.ToolCalls))
	for i, toolCall := range response.ToolCalls {
		toolCallsData[i] = map[string]interface{}{
			jsonKeyID:       toolCall.ID,
			jsonKeyFunction: toolCall.Function.Name,
			jsonKeyArgs:     toolCall.Function.Arguments,
		}
	}

	replyMsg := uc.createAIReplyMessage(msg, finalResponse.Content, aiConfig, processor, toolCallsData)

	// 保存AI回复消息
	if err := uc.messageRepo.Create(ctx, replyMsg); err != nil {
		uc.logger.Error("Failed to save AI reply message with tool calls", "error", err)

		return fmt.Errorf("failed to save AI reply message with tool calls: %w", err)
	}

	uc.logger.Info("AI chat with tool calls processed successfully",
		"provider", aiConfig.Provider,
		"model", aiConfig.Model,
		"message_id", msg.ID,
		"reply_id", replyMsg.ID,
		"tool_calls", len(response.ToolCalls))

	return nil
}

// executeToolCall 执行工具调用.
func (uc *messageUseCase) executeToolCall(ctx context.Context, toolCall *schema.ToolCall) (string, error) {
	uc.logger.Info("Executing tool call", "function", toolCall.Function.Name, "args", toolCall.Function.Arguments)

	// 根据工具名称执行相应的功能
	switch toolCall.Function.Name {
	case "get_current_time":
		return uc.toolGetCurrentTime(ctx, toolCall.Function.Arguments)
	case "get_weather":
		return uc.toolGetWeather(ctx, toolCall.Function.Arguments)
	case "search_web":
		return uc.toolSearchWeb(ctx, toolCall.Function.Arguments)
	case "send_message":
		return uc.toolSendMessage(ctx, toolCall.Function.Arguments)
	default:
		return "", fmt.Errorf("%w: %s", errUnknownToolFunction, toolCall.Function.Name)
	}
}

// toolGetCurrentTime 获取当前时间工具.
func (uc *messageUseCase) toolGetCurrentTime(_ context.Context, _ string) (string, error) {
	now := time.Now()

	return fmt.Sprintf("Current time: %s", now.Format("2006-01-02 15:04:05 MST")), nil
}

// toolGetWeather 获取天气工具.
func (uc *messageUseCase) toolGetWeather(_ context.Context, args string) (string, error) {
	// 解析参数
	var params struct {
		Location string `json:"location"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid weather tool arguments: %w", err)
	}

	// Placeholder until a real weather provider is wired.
	return fmt.Sprintf("Weather in %s: Sunny, 25°C", params.Location), nil
}

// toolSearchWeb 网络搜索工具.
func (uc *messageUseCase) toolSearchWeb(_ context.Context, args string) (string, error) {
	// 解析参数
	var params struct {
		Query string `json:"query"`
		Count int    `json:"count,omitempty"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid search tool arguments: %w", err)
	}

	if params.Count == 0 {
		params.Count = 3
	}

	// Placeholder until a real search provider is wired.
	return fmt.Sprintf("Search results for '%s': Found %d relevant results", params.Query, params.Count), nil
}

// toolSendMessage 发送消息工具.
func (uc *messageUseCase) toolSendMessage(ctx context.Context, args string) (string, error) {
	// 解析参数
	var params struct {
		ChannelID string `json:"channel_id"`
		Content   string `json:"content"`
	}

	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", fmt.Errorf("invalid send message tool arguments: %w", err)
	}

	// 创建新消息
	newMsg := &entity.Message{
		ChannelID:         params.ChannelID,
		PlatformMessageID: fmt.Sprintf("tool_msg_%d", time.Now().Unix()),
		Direction:         entity.MessageDirectionOutbound,
		MessageType:       entity.MessageTypeText,
		ContentType:       entity.MessageTypeText,
		SenderID:          "ai_tool",
		SenderName:        "AI Tool",
		SenderType:        "bot",
		Content:           params.Content,
		MessageStatus:     entity.MessageStatusProcessed,
		PlatformTimestamp: time.Now(),
	}

	// 发送消息
	if err := uc.SendMessage(ctx, newMsg); err != nil {
		return "", fmt.Errorf("failed to send message via tool: %w", err)
	}

	return fmt.Sprintf("Message sent successfully to channel %s", params.ChannelID), nil
}

// createAIReplyMessage 创建AI回复消息.
func (uc *messageUseCase) createAIReplyMessage(msg *entity.Message, content string, aiConfig *struct {
	Provider     string                   `json:"provider"`
	Model        string                   `json:"model"`
	SystemPrompt string                   `json:"system_prompt"`
	Temperature  float64                  `json:"temperature"`
	MaxTokens    int                      `json:"max_tokens"`
	APIKey       string                   `json:"api_key"`
	BaseURL      string                   `json:"base_url,omitempty"`
	EnableTools  bool                     `json:"enable_tools"`
	Tools        []map[string]interface{} `json:"tools,omitempty"`
	StreamMode   bool                     `json:"stream_mode"`
}, processor *entity.MessageProcessor, toolCalls []map[string]interface{},
) *entity.Message {
	unifiedContent := entity.JSONField{
		"ai_provider":         aiConfig.Provider,
		"ai_model":            aiConfig.Model,
		payloadKeyProcessorID: processor.ID,
		configKeyStreamMode:   aiConfig.StreamMode,
	}

	if toolCalls != nil {
		unifiedContent["tool_calls"] = toolCalls
	}

	// T04: reuse DM/group targeting (do not always reply to SenderID).
	reply := BuildOutboundReplyFromInbound(msg, content, msg.MessageType)
	reply.PlatformMessageID = fmt.Sprintf("ai_reply_%s_%d", msg.ID, time.Now().Unix())
	reply.ContentType = entity.MessageTypeText
	reply.SenderID = "ai_bot"
	reply.SenderName = fmt.Sprintf("AI Bot (%s)", aiConfig.Model)
	reply.SenderType = entity.SenderTypeBot
	reply.UnifiedContent = unifiedContent
	// Persist-only AI path historically marked processed without deliver (T10).
	reply.MessageStatus = entity.MessageStatusProcessed
	now := time.Now()
	reply.ProcessedAt = &now

	return reply
}

// extractToolCallFromMessage 从消息中提取工具调用信息.
func (uc *messageUseCase) extractToolCallFromMessage(msg *entity.Message) []schema.ToolCall {
	if msg.UnifiedContent == nil {
		return nil
	}

	toolCallsData, exists := msg.UnifiedContent["tool_calls"]
	if !exists {
		return nil
	}

	// 尝试转换为工具调用列表
	toolCallsSlice, ok := toolCallsData.([]interface{})
	if !ok {
		return nil
	}

	var toolCalls []schema.ToolCall

	for _, item := range toolCallsSlice {
		toolCallMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		id, ok := toolCallMap[jsonKeyID].(string)
		if !ok {
			continue
		}

		functionName, ok := toolCallMap[jsonKeyFunction].(string)
		if !ok {
			continue
		}

		args, ok := toolCallMap[jsonKeyArgs].(string)
		if !ok {
			continue
		}

		toolCalls = append(toolCalls, schema.ToolCall{
			ID: id,
			Function: schema.FunctionCall{
				Name:      functionName,
				Arguments: args,
			},
		})
	}

	return toolCalls
}

// getConversationHistory 获取会话历史消息（T06：必须带 channelID，避免跨 bot 历史串扰）.
func (uc *messageUseCase) getConversationHistory(ctx context.Context, channelID, conversationID string, limit int) ([]*entity.Message, error) {
	if conversationID == "" || channelID == "" {
		return []*entity.Message{}, nil
	}

	params := port.ListParams{
		Page:     1,
		PageSize: limit,
		OrderBy:  orderByCreatedAt,
	}

	result, err := uc.messageRepo.GetByConversationID(ctx, channelID, conversationID, params)
	if err != nil {
		return nil, err
	}

	return result.Items, nil
}
