-- migrate:up
-- Disable foreign key checks to avoid errors during modification
SET FOREIGN_KEY_CHECKS = 0;

-- Bots Table
ALTER TABLE `bots`
    MODIFY `id` VARCHAR(36) NOT NULL,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- Channels Table
ALTER TABLE `channels`
    MODIFY `id` VARCHAR(36) NOT NULL,
    MODIFY `bot_id` VARCHAR(36) NOT NULL,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- Messages Table
ALTER TABLE `messages`
    MODIFY `id` VARCHAR(36) NOT NULL,
    MODIFY `channel_id` VARCHAR(36) NOT NULL,
    MODIFY `parent_message_id` VARCHAR(36),
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- Message Processors Table
ALTER TABLE `message_processors`
    MODIFY `id` VARCHAR(36) NOT NULL,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- Message Routing Rules Table
ALTER TABLE `message_routing_rules`
    MODIFY `id` VARCHAR(36) NOT NULL,
    MODIFY `processor_id` VARCHAR(36) NOT NULL,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- System Config Table
ALTER TABLE `system_configs`
    MODIFY `id` VARCHAR(36) NOT NULL,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- Message Queue Table
ALTER TABLE `message_queues`
    MODIFY `id` VARCHAR(36) NOT NULL,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- Connection Logs Table
ALTER TABLE `connection_logs`
    MODIFY `id` VARCHAR(36) NOT NULL,
    MODIFY `channel_id` VARCHAR(36) NOT NULL,
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- API Call Logs Table
ALTER TABLE `api_call_logs`
    MODIFY `id` VARCHAR(36) NOT NULL,
    MODIFY `channel_id` VARCHAR(36),
    MODIFY `processor_id` VARCHAR(36),
    DROP PRIMARY KEY,
    ADD PRIMARY KEY (`id`);

-- Re-enable foreign key checks
SET FOREIGN_KEY_CHECKS = 1;

-- Note: In a real-world scenario with existing data, you would also need to:
-- 1. Add new UUID columns.
-- 2. Populate them with UUIDs (potentially generating UUIDv7 from created_at timestamps).
-- 3. Update foreign key columns with the new UUIDs.
-- 4. Drop old integer columns and rename new UUID columns.
-- Since this is a development environment, a simple type modification is sufficient for now.

