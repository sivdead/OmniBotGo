-- OmniBotGo 数据库架构回滚
-- 按照依赖关系逆序删除表

-- 删除日志表
DROP TABLE IF EXISTS api_call_logs;
DROP TABLE IF EXISTS connection_logs;

-- 删除队列表
DROP TABLE IF EXISTS message_queue;

-- 删除路由规则表
DROP TABLE IF EXISTS message_routing_rules;

-- 删除处理器表
DROP TABLE IF EXISTS message_processors;

-- 删除系统配置表
DROP TABLE IF EXISTS system_configs;

-- 删除消息表
DROP TABLE IF EXISTS messages;

-- 删除通道表
DROP TABLE IF EXISTS channels;

-- 删除Bot表
DROP TABLE IF EXISTS bots; 