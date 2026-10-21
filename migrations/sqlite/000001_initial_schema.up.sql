-- OmniBotGo SQLite initial schema (UUID/varchar IDs; E02/E04).
-- Enable foreign_keys via DSN _pragma on every connection (E08).

CREATE TABLE bots (
    id TEXT PRIMARY KEY,
    bot_name TEXT NOT NULL,
    bot_type TEXT DEFAULT 'standard',
    description TEXT,
    avatar_url TEXT,
    config TEXT,
    status INTEGER DEFAULT 1,
    created_by TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

CREATE TABLE channels (
    id TEXT PRIMARY KEY,
    bot_id TEXT NOT NULL REFERENCES bots (id),
    platform_type TEXT NOT NULL,
    channel_name TEXT NOT NULL,
    webhook_path TEXT UNIQUE,
    config TEXT NOT NULL,
    access_token TEXT,
    access_token_expires_at DATETIME,
    connection_status INTEGER DEFAULT 0,
    last_connected_at DATETIME NULL,
    status INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE (bot_id, platform_type, channel_name)
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL UNIQUE,
    channel_id TEXT NOT NULL REFERENCES channels (id),
    platform_message_id TEXT,
    direction INTEGER NOT NULL,
    message_type TEXT NOT NULL,
    content_type TEXT,
    sender_id TEXT,
    sender_name TEXT,
    sender_type TEXT,
    receiver_id TEXT,
    receiver_name TEXT,
    receiver_type TEXT,
    content TEXT,
    raw_content TEXT,
    unified_content TEXT,
    media_url TEXT,
    media_type TEXT,
    file_size INTEGER,
    message_status INTEGER DEFAULT 0,
    retry_count INTEGER DEFAULT 0,
    error_message TEXT,
    parent_message_id TEXT REFERENCES messages (id),
    conversation_id TEXT,
    backend_request_id TEXT,
    platform_timestamp DATETIME,
    received_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME,
    sent_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

CREATE TABLE message_processors (
    id TEXT PRIMARY KEY,
    processor_name TEXT NOT NULL,
    processor_type TEXT NOT NULL,
    config TEXT NOT NULL,
    priority INTEGER DEFAULT 100,
    status INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

CREATE TABLE message_routing_rules (
    id TEXT PRIMARY KEY,
    rule_name TEXT NOT NULL,
    rule_description TEXT,
    bot_ids TEXT,
    platform_types TEXT,
    channel_ids TEXT,
    message_types TEXT,
    sender_patterns TEXT,
    content_patterns TEXT,
    processor_id TEXT NOT NULL REFERENCES message_processors (id),
    route_type INTEGER DEFAULT 1,
    priority INTEGER DEFAULT 100,
    is_fallback INTEGER DEFAULT 0,
    condition_logic TEXT DEFAULT 'AND',
    status INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

CREATE TABLE system_configs (
    id TEXT PRIMARY KEY,
    config_key TEXT UNIQUE NOT NULL,
    config_value TEXT,
    config_type TEXT DEFAULT 'string',
    config_group TEXT,
    description TEXT,
    is_encrypted INTEGER DEFAULT 0,
    is_system INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

CREATE TABLE message_queue (
    id TEXT PRIMARY KEY,
    queue_name TEXT NOT NULL,
    message_id TEXT NOT NULL,
    payload TEXT NOT NULL,
    priority INTEGER DEFAULT 100,
    max_retries INTEGER DEFAULT 3,
    retry_count INTEGER DEFAULT 0,
    status INTEGER DEFAULT 0,
    scheduled_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    completed_at DATETIME,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

CREATE TABLE connection_logs (
    id TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES channels (id),
    event_type TEXT NOT NULL,
    status INTEGER,
    error_code TEXT,
    error_message TEXT,
    details TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);

CREATE TABLE api_call_logs (
    id TEXT PRIMARY KEY,
    request_id TEXT,
    channel_id TEXT REFERENCES channels (id),
    processor_id TEXT REFERENCES message_processors (id),
    method TEXT,
    url TEXT,
    request_headers TEXT,
    request_body TEXT,
    response_status INTEGER,
    response_headers TEXT,
    response_body TEXT,
    start_time DATETIME,
    end_time DATETIME,
    duration_ms INTEGER,
    success INTEGER,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL
);
