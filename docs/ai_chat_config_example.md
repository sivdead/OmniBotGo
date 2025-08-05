# AI聊天功能配置示例

## 概述

OmniBotGo现在支持使用eino框架进行AI聊天功能，支持多种AI供应商、工具调用和流式处理。

## 支持的AI供应商

- **OpenAI**: GPT-3.5, GPT-4等模型 ✅
- **Claude**: 计划支持（待实现）🚧
- **Gemini**: 计划支持（待实现）🚧

## 核心功能

- ✅ **多模型支持**: 支持不同AI供应商和模型
- ✅ **工具调用**: 支持Function Calling，让AI调用外部工具
- ✅ **流式处理**: 支持实时响应流
- ✅ **会话历史**: 自动管理对话上下文
- ✅ **自定义端点**: 支持使用代理或私有部署

## 基础配置示例

### 1. OpenAI基础配置

```json
{
  "provider": "openai",
  "model": "gpt-3.5-turbo",
  "system_prompt": "你是一个友好的AI助手，请用中文回答问题。",
  "temperature": 0.7,
  "max_tokens": 1024,
  "api_key": "sk-your-openai-api-key-here",
  "base_url": "https://api.openai.com"
}
```

### 2. GPT-4配置

```json
{
  "provider": "openai",
  "model": "gpt-4",
  "system_prompt": "你是一个专业的技术助手，擅长编程和系统架构。",
  "temperature": 0.8,
  "max_tokens": 2048,
  "api_key": "sk-your-openai-api-key-here"
}
```

### 3. 使用自定义端点

```json
{
  "provider": "openai",
  "model": "gpt-3.5-turbo",
  "system_prompt": "You are a helpful assistant.",
  "temperature": 0.7,
  "max_tokens": 1024,
  "api_key": "your-api-key",
  "base_url": "https://your-custom-openai-endpoint.com"
}
```

## 高级功能配置

### 4. 启用工具调用

```json
{
  "provider": "openai",
  "model": "gpt-3.5-turbo",
  "system_prompt": "你是一个智能助手，可以调用各种工具来帮助用户。",
  "temperature": 0.7,
  "max_tokens": 2048,
  "api_key": "sk-your-openai-api-key-here",
  "enable_tools": true,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_current_time",
        "description": "获取当前时间"
      }
    },
    {
      "type": "function", 
      "function": {
        "name": "get_weather",
        "description": "获取指定地点的天气信息",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "城市名称"
            }
          },
          "required": ["location"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "search_web",
        "description": "搜索网络信息",
        "parameters": {
          "type": "object",
          "properties": {
            "query": {
              "type": "string",
              "description": "搜索关键词"
            },
            "count": {
              "type": "integer",
              "description": "返回结果数量",
              "default": 3
            }
          },
          "required": ["query"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "send_message",
        "description": "发送消息到指定频道",
        "parameters": {
          "type": "object",
          "properties": {
            "channel_id": {
              "type": "integer",
              "description": "目标频道ID"
            },
            "content": {
              "type": "string",
              "description": "消息内容"
            }
          },
          "required": ["channel_id", "content"]
        }
      }
    }
  ]
}
```

### 5. 启用流式处理

```json
{
  "provider": "openai",
  "model": "gpt-4",
  "system_prompt": "你是一个实时响应的AI助手。",
  "temperature": 0.7,
  "max_tokens": 2048,
  "api_key": "sk-your-openai-api-key-here",
  "stream_mode": true,
  "enable_tools": false
}
```

### 6. 完整高级配置

```json
{
  "provider": "openai",
  "model": "gpt-4",
  "system_prompt": "你是一个全能的AI助手，可以回答问题、调用工具、提供实时响应。请根据用户需求灵活使用各种功能。",
  "temperature": 0.8,
  "max_tokens": 4096,
  "api_key": "sk-your-openai-api-key-here",
  "base_url": "https://api.openai.com",
  "stream_mode": true,
  "enable_tools": true,
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "get_current_time",
        "description": "获取当前时间"
      }
    },
    {
      "type": "function",
      "function": {
        "name": "get_weather", 
        "description": "获取天气信息",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "城市名称"
            }
          },
          "required": ["location"]
        }
      }
    },
    {
      "type": "function",
      "function": {
        "name": "search_web",
        "description": "搜索网络信息",
        "parameters": {
          "type": "object",
          "properties": {
            "query": {
              "type": "string",
              "description": "搜索关键词"
            },
            "count": {
              "type": "integer",
              "description": "返回结果数量",
              "default": 5
            }
          },
          "required": ["query"]
        }
      }
    }
  ]
}
```

## 可用工具列表

### 内置工具

| 工具名称 | 功能描述 | 参数 | 状态 |
|---------|---------|------|------|
| `get_current_time` | 获取当前时间 | 无 | ✅ 已实现 |
| `get_weather` | 获取天气信息 | location: 城市名称 | 🚧 模拟数据 |
| `search_web` | 网络搜索 | query: 搜索词, count: 结果数量 | 🚧 模拟数据 |
| `send_message` | 发送消息 | channel_id: 频道ID, content: 内容 | ✅ 已实现 |

### 扩展工具（计划中）

- `get_user_info` - 获取用户信息
- `create_reminder` - 创建提醒
- `translate_text` - 文本翻译
- `generate_image` - 图像生成
- `analyze_file` - 文件分析

## 配置参数说明

### 基础参数

- **provider**: AI供应商（`openai`, `claude`, `gemini`）
- **model**: 模型名称（如 `gpt-3.5-turbo`, `gpt-4`）
- **system_prompt**: 系统提示词，定义AI的角色和行为
- **temperature**: 创造性参数（0.0-2.0），越高越有创意
- **max_tokens**: 最大输出token数
- **api_key**: API密钥

### 高级参数

- **base_url**: 自定义API端点（可选）
- **stream_mode**: 是否启用流式响应（默认：false）
- **enable_tools**: 是否启用工具调用（默认：false）
- **tools**: 可用工具列表（当enable_tools为true时）

## 使用场景

### 1. 客服机器人
```json
{
  "provider": "openai",
  "model": "gpt-3.5-turbo",
  "system_prompt": "你是一个专业的客服助手，请耐心、友好地帮助用户解决问题。",
  "temperature": 0.3,
  "max_tokens": 1024,
  "enable_tools": true,
  "tools": [
    {"type": "function", "function": {"name": "get_current_time"}},
    {"type": "function", "function": {"name": "send_message"}}
  ]
}
```

### 2. 技术支持
```json
{
  "provider": "openai",
  "model": "gpt-4",
  "system_prompt": "你是一个技术支持专家，擅长解决各种技术问题。",
  "temperature": 0.2,
  "max_tokens": 2048,
  "enable_tools": true,
  "tools": [
    {"type": "function", "function": {"name": "search_web"}},
    {"type": "function", "function": {"name": "get_current_time"}}
  ]
}
```

### 3. 智能助手
```json
{
  "provider": "openai",
  "model": "gpt-4",
  "system_prompt": "你是一个全能的智能助手，可以帮助用户处理各种任务。",
  "temperature": 0.7,
  "max_tokens": 3072,
  "stream_mode": true,
  "enable_tools": true,
  "tools": [
    {"type": "function", "function": {"name": "get_current_time"}},
    {"type": "function", "function": {"name": "get_weather"}},
    {"type": "function", "function": {"name": "search_web"}},
    {"type": "function", "function": {"name": "send_message"}}
  ]
}
```

## 注意事项

1. **API密钥安全**: 请妥善保管API密钥，不要在配置中硬编码
2. **成本控制**: 设置合理的max_tokens避免过高费用
3. **工具权限**: 谨慎配置工具权限，避免安全风险
4. **流式处理**: 流式模式下暂不支持工具调用
5. **错误处理**: 系统会自动处理API调用失败和重试

## 性能优化建议

1. **合理设置max_tokens**: 根据实际需求设置，避免浪费
2. **优化system_prompt**: 简洁明确的提示词能提高响应质量
3. **选择合适的模型**: 根据任务复杂度选择模型
4. **启用缓存**: 对于重复查询启用缓存机制
5. **监控使用量**: 定期检查API使用情况和成本

## 创建AI聊天处理器

### 1. 通过API创建处理器

```bash
curl -X POST http://localhost:8080/api/v1/processors \
  -H "Content-Type: application/json" \
  -d '{
    "name": "OpenAI GPT-3.5 聊天助手",
    "processor_type": "ai_chat",
    "description": "使用GPT-3.5-turbo的AI聊天助手",
    "priority": 100,
    "is_active": true,
    "config": {
      "provider": "openai",
      "model": "gpt-3.5-turbo",
      "system_prompt": "你是一个友好的AI助手，请用中文回答问题。",
      "temperature": 0.7,
      "max_tokens": 1024,
      "api_key": "sk-your-openai-api-key-here"
    }
  }'
```

### 2. 创建路由规则

```bash
curl -X POST http://localhost:8080/api/v1/routing-rules \
  -H "Content-Type: application/json" \
  -d '{
    "processor_id": 1,
    "rule_name": "AI聊天路由",
    "priority": 100,
    "is_fallback": false,
    "conditions": {
      "message_types": ["text"],
      "platform_types": ["dingtalk", "wecom"]
    }
  }'
```

## 功能特性

### 1. 会话历史支持

AI聊天功能会自动获取会话历史，为AI提供上下文信息，实现连续对话。

### 2. 多模型支持

可以为不同的通道或场景配置不同的AI模型：

- 客服场景：使用温度较低的配置，提供准确回答
- 创意场景：使用温度较高的配置，提供更有创意的回答

### 3. 错误处理

- 自动重试网络错误
- 详细的错误日志记录
- 优雅的降级处理

### 4. 性能优化

- 使用eino框架的流式处理
- 支持并发请求
- 智能的token管理

## 使用建议

### 1. 系统提示词优化

- 明确指定回答语言
- 设定AI的角色和专业领域
- 包含必要的行为准则

### 2. 参数调优

- **temperature**: 0.1-0.3（准确回答），0.7-0.9（创意回答）
- **max_tokens**: 根据使用场景调整，避免过长回答
- **system_prompt**: 简洁明确，避免过于复杂

### 3. 成本控制

- 合理设置max_tokens限制
- 监控API调用频率
- 考虑使用较便宜的模型进行测试

## 故障排除

### 1. 常见错误

- **API key无效**: 检查API密钥是否正确
- **模型不存在**: 确认模型名称是否正确
- **配额不足**: 检查API账户余额
- **网络超时**: 检查网络连接和API端点

### 2. 调试方法

- 查看应用日志中的详细错误信息
- 使用API Explorer测试配置
- 检查处理器和路由规则配置

## 扩展开发

### 1. 添加新的AI供应商

1. 在`createChatModel`函数中添加新的case
2. 实现对应的ChatModel创建函数
3. 添加供应商特定的配置验证

### 2. 自定义处理逻辑

可以在`processAIChat`函数中添加：
- 消息预处理
- 响应后处理
- 特殊命令处理
- 多轮对话管理

## 更新日志

- **v1.0.0**: 初始版本，支持OpenAI模型
- **v1.1.0**: 集成eino框架，支持配置化模型和供应商
- **v1.2.0**: 计划支持Claude和Gemini模型 