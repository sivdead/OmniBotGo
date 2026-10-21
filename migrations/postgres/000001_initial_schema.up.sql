-- OmniBotGo PostgreSQL initial schema (UUID/varchar IDs from day one; E02).
-- Verified path for new Postgres installs; not a mechanical MySQL dump.

CREATE TABLE bots (
    id VARCHAR(36) PRIMARY KEY,
    bot_name VARCHAR(100) NOT NULL,
    bot_type VARCHAR(50) DEFAULT 'standard',
    description TEXT,
    avatar_url VARCHAR(500),
    config JSONB,
    status SMALLINT DEFAULT 1,
    created_by VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX idx_bots_bot_type ON bots (bot_type);
CREATE INDEX idx_bots_status ON bots (status);
CREATE INDEX idx_bots_deleted_at ON bots (deleted_at);

CREATE TABLE channels (
    id VARCHAR(36) PRIMARY KEY,
    bot_id VARCHAR(36) NOT NULL REFERENCES bots (id),
    platform_type VARCHAR(50) NOT NULL,
    channel_name VARCHAR(100) NOT NULL,
    webhook_path VARCHAR(200) UNIQUE,
    config JSONB NOT NULL,
    access_token VARCHAR(1000),
    access_token_expires_at TIMESTAMPTZ,
    connection_status SMALLINT DEFAULT 0,
    last_connected_at TIMESTAMPTZ NULL,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    UNIQUE (bot_id, platform_type, channel_name)
);
CREATE INDEX idx_channels_bot_id ON channels (bot_id);
CREATE INDEX idx_channels_platform_type ON channels (platform_type);
CREATE INDEX idx_channels_deleted_at ON channels (deleted_at);

CREATE TABLE messages (
    id VARCHAR(36) PRIMARY KEY,
    message_id VARCHAR(100) NOT NULL UNIQUE,
    channel_id VARCHAR(36) NOT NULL REFERENCES channels (id),
    platform_message_id VARCHAR(200),
    direction SMALLINT NOT NULL,
    message_type VARCHAR(50) NOT NULL,
    content_type VARCHAR(50),
    sender_id VARCHAR(100),
    sender_name VARCHAR(100),
    sender_type VARCHAR(50),
    receiver_id VARCHAR(100),
    receiver_name VARCHAR(100),
    receiver_type VARCHAR(50),
    content TEXT,
    raw_content JSONB,
    unified_content JSONB,
    media_url VARCHAR(500),
    media_type VARCHAR(50),
    file_size BIGINT,
    message_status SMALLINT DEFAULT 0,
    retry_count INT DEFAULT 0,
    error_message TEXT,
    parent_message_id VARCHAR(36) REFERENCES messages (id),
    conversation_id VARCHAR(100),
    backend_request_id VARCHAR(100),
    platform_timestamp TIMESTAMPTZ,
    received_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX idx_messages_channel_message ON messages (channel_id, platform_message_id);
CREATE INDEX idx_messages_direction_status ON messages (direction, message_status);
CREATE INDEX idx_messages_deleted_at ON messages (deleted_at);

CREATE TABLE message_processors (
    id VARCHAR(36) PRIMARY KEY,
    processor_name VARCHAR(100) NOT NULL,
    processor_type VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    priority INT DEFAULT 100,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE message_routing_rules (
    id VARCHAR(36) PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL,
    rule_description TEXT,
    bot_ids JSONB,
    platform_types JSONB,
    channel_ids JSONB,
    message_types JSONB,
    sender_patterns JSONB,
    content_patterns JSONB,
    processor_id VARCHAR(36) NOT NULL REFERENCES message_processors (id),
    route_type SMALLINT DEFAULT 1,
    priority INT DEFAULT 100,
    is_fallback SMALLINT DEFAULT 0,
    condition_logic VARCHAR(10) DEFAULT 'AND',
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE system_configs (
    id VARCHAR(36) PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    config_type VARCHAR(50) DEFAULT 'string',
    config_group VARCHAR(50),
    description TEXT,
    is_encrypted SMALLINT DEFAULT 0,
    is_system SMALLINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE message_queue (
    id VARCHAR(36) PRIMARY KEY,
    queue_name VARCHAR(100) NOT NULL,
    message_id VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    priority INT DEFAULT 100,
    max_retries INT DEFAULT 3,
    retry_count INT DEFAULT 0,
    status SMALLINT DEFAULT 0,
    scheduled_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
CREATE INDEX idx_message_queue_status ON message_queue (queue_name, status);

CREATE TABLE connection_logs (
    id VARCHAR(36) PRIMARY KEY,
    channel_id VARCHAR(36) NOT NULL REFERENCES channels (id),
    event_type VARCHAR(50) NOT NULL,
    status SMALLINT,
    error_code VARCHAR(50),
    error_message TEXT,
    details JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE TABLE api_call_logs (
    id VARCHAR(36) PRIMARY KEY,
    request_id VARCHAR(100),
    channel_id VARCHAR(36) REFERENCES channels (id),
    processor_id VARCHAR(36) REFERENCES message_processors (id),
    method VARCHAR(10),
    url VARCHAR(1000),
    request_headers JSONB,
    request_body TEXT,
    response_status INT,
    response_headers JSONB,
    response_body TEXT,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    duration_ms INT,
    success SMALLINT,
    error_message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);
