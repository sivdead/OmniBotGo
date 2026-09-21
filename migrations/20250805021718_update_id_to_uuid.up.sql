-- migrate:up
-- MySQL Error 3780: ALTER of PK/FK columns fails while FKs exist even with
-- FOREIGN_KEY_CHECKS=0. Drop FKs, convert types, then recreate FKs.

SET FOREIGN_KEY_CHECKS = 0;

-- Drop foreign keys created by 000001_initial_schema (InnoDB default names)
ALTER TABLE `channels` DROP FOREIGN KEY `channels_ibfk_1`;
ALTER TABLE `messages` DROP FOREIGN KEY `messages_ibfk_1`;
ALTER TABLE `messages` DROP FOREIGN KEY `messages_ibfk_2`;
ALTER TABLE `message_routing_rules` DROP FOREIGN KEY `message_routing_rules_ibfk_1`;
ALTER TABLE `connection_logs` DROP FOREIGN KEY `connection_logs_ibfk_1`;
ALTER TABLE `api_call_logs` DROP FOREIGN KEY `api_call_logs_ibfk_1`;
ALTER TABLE `api_call_logs` DROP FOREIGN KEY `api_call_logs_ibfk_2`;

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

-- Message Queue Table (initial schema uses singular `message_queue`)
ALTER TABLE `message_queue`
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

-- Recreate foreign keys with matching VARCHAR(36) types
ALTER TABLE `channels`
    ADD CONSTRAINT `channels_ibfk_1` FOREIGN KEY (`bot_id`) REFERENCES `bots` (`id`);

ALTER TABLE `messages`
    ADD CONSTRAINT `messages_ibfk_1` FOREIGN KEY (`channel_id`) REFERENCES `channels` (`id`),
    ADD CONSTRAINT `messages_ibfk_2` FOREIGN KEY (`parent_message_id`) REFERENCES `messages` (`id`);

ALTER TABLE `message_routing_rules`
    ADD CONSTRAINT `message_routing_rules_ibfk_1` FOREIGN KEY (`processor_id`) REFERENCES `message_processors` (`id`);

ALTER TABLE `connection_logs`
    ADD CONSTRAINT `connection_logs_ibfk_1` FOREIGN KEY (`channel_id`) REFERENCES `channels` (`id`);

ALTER TABLE `api_call_logs`
    ADD CONSTRAINT `api_call_logs_ibfk_1` FOREIGN KEY (`channel_id`) REFERENCES `channels` (`id`),
    ADD CONSTRAINT `api_call_logs_ibfk_2` FOREIGN KEY (`processor_id`) REFERENCES `message_processors` (`id`);

SET FOREIGN_KEY_CHECKS = 1;
