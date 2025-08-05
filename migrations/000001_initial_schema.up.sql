-- OmniBotGo 初始数据库架构
-- 创建时间: 2024-12-19

-- ===========================
-- 1. Bot实例表 (核心表)
-- ===========================
CREATE TABLE bots (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    bot_name VARCHAR(100) NOT NULL COMMENT 'Bot名称',
    bot_type VARCHAR(50) DEFAULT 'standard' COMMENT 'Bot类型: standard, ai, custom',
    description TEXT COMMENT 'Bot描述',
    avatar_url VARCHAR(500) COMMENT 'Bot头像URL',
    
    -- Bot配置
    config JSON COMMENT 'Bot特定配置，如AI模型参数、回复策略等',
    
    -- 状态管理
    status TINYINT DEFAULT 1 COMMENT 'Bot状态: 0-禁用, 1-启用, 2-维护中',
    created_by VARCHAR(100) COMMENT '创建者',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_bot_type (bot_type),
    INDEX idx_status (status),
    INDEX idx_created_by (created_by),
    INDEX idx_deleted_at (deleted_at)
) COMMENT 'Bot实例表';

-- ===========================
-- 2. 消息通道配置表
-- ===========================
CREATE TABLE channels (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    bot_id BIGINT NOT NULL COMMENT '关联的Bot ID',
    platform_type VARCHAR(50) NOT NULL COMMENT '平台类型: wecom, dingtalk, wechat_official, feishu',
    channel_name VARCHAR(100) NOT NULL COMMENT '通道名称',
    webhook_path VARCHAR(200) UNIQUE COMMENT '该通道的唯一Webhook路径: /webhook/channel-{id}',
    
    -- 平台配置 (JSON格式，代码中校验)
    config JSON NOT NULL COMMENT '平台特定配置，如: {"corp_id": "xxx", "app_id": "xxx", "app_secret": "xxx"}',
    
    -- 访问令牌 (运行时动态获取和刷新)
    access_token VARCHAR(1000) COMMENT '访问令牌',
    access_token_expires_at TIMESTAMP COMMENT '访问令牌过期时间',
    
    -- 状态管理
    connection_status TINYINT DEFAULT 0 COMMENT '连接状态: 0-未连接, 1-已连接, 2-连接失败',
    last_connected_at TIMESTAMP NULL COMMENT '最后连接时间',
    status TINYINT DEFAULT 1 COMMENT '通道状态: 0-禁用, 1-启用',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (bot_id) REFERENCES bots(id),
    INDEX idx_bot_id (bot_id),
    INDEX idx_platform_type (platform_type),
    INDEX idx_connection_status (connection_status),
    INDEX idx_webhook_path (webhook_path),
    INDEX idx_deleted_at (deleted_at),
    UNIQUE KEY uk_bot_platform (bot_id, platform_type, channel_name)
) COMMENT '消息通道配置表';

-- ===========================
-- 3. 消息记录表 (核心表)
-- ===========================
CREATE TABLE messages (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id VARCHAR(100) NOT NULL COMMENT '统一消息ID',
    channel_id BIGINT NOT NULL COMMENT '关联通道',
    platform_message_id VARCHAR(200) COMMENT '平台原始消息ID',
    
    -- 消息基本信息
    direction TINYINT NOT NULL COMMENT '消息方向: 1-接收(上行), 2-发送(下行)',
    message_type VARCHAR(50) NOT NULL COMMENT '消息类型: text, image, audio, video, file, card, event',
    content_type VARCHAR(50) COMMENT '内容类型: text/plain, image/jpeg等',
    
    -- 发送者信息
    sender_id VARCHAR(100) COMMENT '发送者平台ID',
    sender_name VARCHAR(100) COMMENT '发送者名称',
    sender_type VARCHAR(50) COMMENT '发送者类型: user, group, system',
    
    -- 接收者信息
    receiver_id VARCHAR(100) COMMENT '接收者平台ID', 
    receiver_name VARCHAR(100) COMMENT '接收者名称',
    receiver_type VARCHAR(50) COMMENT '接收者类型: user, group, system',
    
    -- 消息内容
    content TEXT COMMENT '消息内容',
    raw_content JSON COMMENT '平台原始消息(完整JSON)',
    unified_content JSON COMMENT '统一格式消息内容',
    
    -- 媒体信息
    media_url VARCHAR(500) COMMENT '媒体文件URL',
    media_type VARCHAR(50) COMMENT '媒体类型',
    file_size BIGINT COMMENT '文件大小(字节)',
    
    -- 消息状态
    message_status TINYINT DEFAULT 1 COMMENT '消息状态: 1-已接收, 2-处理中, 3-已处理, 4-发送中, 5-已发送, 6-发送失败',
    retry_count INT DEFAULT 0 COMMENT '重试次数',
    error_message TEXT COMMENT '错误信息',
    
    -- 关联信息
    parent_message_id BIGINT COMMENT '父消息ID(回复消息)',
    conversation_id VARCHAR(100) COMMENT '会话ID',
    backend_request_id VARCHAR(100) COMMENT '后端请求ID',
    
    -- 时间戳
    platform_timestamp TIMESTAMP COMMENT '平台时间戳',
    received_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP COMMENT '处理时间',
    sent_at TIMESTAMP COMMENT '发送时间',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    FOREIGN KEY (parent_message_id) REFERENCES messages(id),
    
    UNIQUE KEY uk_message_id (message_id),
    INDEX idx_channel_message (channel_id, platform_message_id),
    INDEX idx_sender (sender_id, sender_type),
    INDEX idx_receiver (receiver_id, receiver_type),
    INDEX idx_message_type (message_type),
    INDEX idx_direction_status (direction, message_status),
    INDEX idx_conversation (conversation_id),
    INDEX idx_received_at (received_at),
    INDEX idx_backend_request (backend_request_id),
    INDEX idx_deleted_at (deleted_at)
) COMMENT '消息记录表';

-- ===========================
-- 4. 消息处理器配置表
-- ===========================
CREATE TABLE message_processors (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    processor_name VARCHAR(100) NOT NULL COMMENT '处理器名称',
    processor_type VARCHAR(50) NOT NULL COMMENT '处理器类型: openai_chat, webhook_forwarder, order_query等',
    
    -- 实例配置 (JSON格式，代码中校验)
    config JSON NOT NULL COMMENT '处理器特定配置，如: {"api_key": "xxx", "base_url": "xxx", "model": "gpt-3.5"}',
    
    -- 状态管理
    priority INT DEFAULT 100 COMMENT '处理器优先级',
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_processor_type (processor_type),
    INDEX idx_processor_name (processor_name),
    INDEX idx_priority (priority),
    INDEX idx_deleted_at (deleted_at)
) COMMENT '消息处理器配置表';

-- ===========================
-- 5. 消息路由规则表
-- ===========================
CREATE TABLE message_routing_rules (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    rule_description TEXT COMMENT '规则描述',
    
    -- 匹配条件
    bot_ids JSON COMMENT '匹配的Bot ID列表',
    platform_types JSON COMMENT '匹配的平台类型列表',
    channel_ids JSON COMMENT '匹配的通道ID列表',
    message_types JSON COMMENT '匹配的消息类型列表',
    sender_patterns JSON COMMENT '发送者匹配模式',
    content_patterns JSON COMMENT '内容匹配模式(正则)',
    
    -- 路由目标
    processor_id BIGINT NOT NULL COMMENT '目标处理器ID',
    route_type TINYINT DEFAULT 1 COMMENT '路由类型: 1-同步, 2-异步',
    
    -- 规则配置
    priority INT DEFAULT 100 COMMENT '优先级(数值越小优先级越高)',
    is_fallback TINYINT DEFAULT 0 COMMENT '是否为兜底规则',
    condition_logic VARCHAR(10) DEFAULT 'AND' COMMENT '条件逻辑: AND, OR',
    
    status TINYINT DEFAULT 1 COMMENT '状态: 0-禁用, 1-启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (processor_id) REFERENCES message_processors(id),
    INDEX idx_priority (priority),
    INDEX idx_processor (processor_id),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at)
) COMMENT '消息路由规则表';

-- ===========================
-- 6. 系统配置表
-- ===========================
CREATE TABLE system_configs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(100) UNIQUE NOT NULL COMMENT '配置键',
    config_value TEXT COMMENT '配置值',
    config_type VARCHAR(50) DEFAULT 'string' COMMENT '配置类型: string, int, bool, json',
    config_group VARCHAR(50) COMMENT '配置分组',
    description TEXT COMMENT '配置描述',
    is_encrypted TINYINT DEFAULT 0 COMMENT '是否加密存储',
    is_system TINYINT DEFAULT 0 COMMENT '是否为系统配置',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_config_group (config_group),
    INDEX idx_config_key (config_key),
    INDEX idx_deleted_at (deleted_at)
) COMMENT '系统配置表';

-- ===========================
-- 7. 消息队列表
-- ===========================
CREATE TABLE message_queue (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    queue_name VARCHAR(100) NOT NULL COMMENT '队列名称',
    message_id VARCHAR(100) NOT NULL COMMENT '消息ID',
    payload JSON NOT NULL COMMENT '消息载荷',
    priority INT DEFAULT 100 COMMENT '优先级',
    max_retries INT DEFAULT 3 COMMENT '最大重试次数',
    retry_count INT DEFAULT 0 COMMENT '当前重试次数',
    status TINYINT DEFAULT 0 COMMENT '状态: 0-待处理, 1-处理中, 2-已完成, 3-失败, 4-重试中, 5-已取消',
    scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '计划处理时间',
    started_at TIMESTAMP COMMENT '开始处理时间',
    completed_at TIMESTAMP COMMENT '完成时间',
    error_message TEXT COMMENT '错误信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_queue_status (queue_name, status),
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_message_id (message_id),
    INDEX idx_deleted_at (deleted_at)
) COMMENT '消息队列表';

-- ===========================
-- 8. 连接状态日志表
-- ===========================
CREATE TABLE connection_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    channel_id BIGINT NOT NULL COMMENT '通道ID',
    event_type VARCHAR(50) NOT NULL COMMENT '事件类型: connect, disconnect, reconnect, error',
    status TINYINT COMMENT '连接状态',
    error_code VARCHAR(50) COMMENT '错误代码',
    error_message TEXT COMMENT '错误信息',
    details JSON COMMENT '详细信息',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    INDEX idx_channel_id (channel_id),
    INDEX idx_event_type (event_type),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at)
) COMMENT '连接状态日志表';

-- ===========================
-- 9. API调用记录表
-- ===========================
CREATE TABLE api_call_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    request_id VARCHAR(100) COMMENT '请求ID',
    channel_id BIGINT COMMENT '消息通道ID',
    processor_id BIGINT COMMENT '消息处理器ID',
    
    -- 请求信息
    method VARCHAR(10) COMMENT 'HTTP方法',
    url VARCHAR(1000) COMMENT '请求URL',
    request_headers JSON COMMENT '请求头',
    request_body TEXT COMMENT '请求体',
    
    -- 响应信息
    response_status INT COMMENT '响应状态码',
    response_headers JSON COMMENT '响应头',
    response_body TEXT COMMENT '响应体',
    
    -- 性能信息
    start_time TIMESTAMP COMMENT '开始时间',
    end_time TIMESTAMP COMMENT '结束时间',
    duration_ms INT COMMENT '耗时(毫秒)',
    
    -- 结果信息
    success TINYINT COMMENT '是否成功',
    error_message TEXT COMMENT '错误信息',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (channel_id) REFERENCES channels(id),
    FOREIGN KEY (processor_id) REFERENCES message_processors(id),
    INDEX idx_request_id (request_id),
    INDEX idx_channel_id (channel_id),
    INDEX idx_processor (processor_id),
    INDEX idx_created_at (created_at),
    INDEX idx_success (success),
    INDEX idx_deleted_at (deleted_at)
) COMMENT 'API调用记录表';

-- ===========================
-- 插入初始数据
-- ===========================

-- 插入默认Bot实例
INSERT INTO bots (bot_name, bot_type, description, config, created_by) VALUES
('默认客服Bot', 'standard', '默认的客服机器人，处理常见问答', '{"auto_reply": true, "fallback_message": "您好，我是智能客服，请问有什么可以帮助您的？"}', 'system'),
('AI助手Bot', 'ai', '基于AI的智能助手，支持复杂对话', '{"model": "gpt-3.5-turbo", "temperature": 0.7, "max_tokens": 1000}', 'system');

-- 插入默认系统配置
INSERT INTO system_configs (config_key, config_value, config_type, config_group, description, is_system) VALUES
('server.port', '8080', 'int', 'server', '服务器端口', 1),
('server.host', '0.0.0.0', 'string', 'server', '服务器绑定地址', 1),
('log.level', 'info', 'string', 'logging', '日志级别', 0),
('message.max_size', '10485760', 'int', 'message', '最大消息大小(字节)', 0),
('webhook.timeout', '30', 'int', 'webhook', 'Webhook超时时间(秒)', 0),
('token.default_expire', '7200', 'int', 'auth', '默认Token过期时间(秒)', 0),
('message.retention_days', '30', 'int', 'message', '消息保留天数', 0),
('queue.max_size', '10000', 'int', 'queue', '队列最大长度', 0),
('webhook.base_url', 'http://localhost:8080', 'string', 'webhook', 'Webhook基础URL', 0),
('channel.webhook_path_template', '/webhook/channel-{id}', 'string', 'channel', '通道Webhook路径模板', 1);

-- 插入示例消息处理器
INSERT INTO message_processors (processor_name, processor_type, config, priority) VALUES
('OpenAI聊天处理器', 'openai_chat', '{"api_key": "sk-xxxxx", "model": "gpt-3.5-turbo", "temperature": 0.7, "base_url": "https://api.openai.com"}', 100),
('客服工单系统', 'webhook_forwarder', '{"webhook_url": "https://crm.company.com/api/webhook", "api_key": "crm_api_key", "timeout": 30}', 200),
('订单查询处理器', 'order_query', '{"api_base_url": "https://order.company.com/api", "api_key": "order_api_key"}', 300),
('邮件通知处理器', 'email_notifier', '{"smtp_host": "mail.company.com", "smtp_user": "bot@company.com", "smtp_pass": "xxxxx"}', 400); 