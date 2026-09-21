-- migrate:down
SET FOREIGN_KEY_CHECKS = 0;

ALTER TABLE `channels` DROP FOREIGN KEY `channels_ibfk_1`;
ALTER TABLE `messages` DROP FOREIGN KEY `messages_ibfk_1`;
ALTER TABLE `messages` DROP FOREIGN KEY `messages_ibfk_2`;
ALTER TABLE `message_routing_rules` DROP FOREIGN KEY `message_routing_rules_ibfk_1`;
ALTER TABLE `connection_logs` DROP FOREIGN KEY `connection_logs_ibfk_1`;
ALTER TABLE `api_call_logs` DROP FOREIGN KEY `api_call_logs_ibfk_1`;
ALTER TABLE `api_call_logs` DROP FOREIGN KEY `api_call_logs_ibfk_2`;

ALTER TABLE `bots` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT;
ALTER TABLE `channels` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT, MODIFY `bot_id` BIGINT NOT NULL;
ALTER TABLE `messages` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT, MODIFY `channel_id` BIGINT NOT NULL, MODIFY `parent_message_id` BIGINT;
ALTER TABLE `message_processors` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT;
ALTER TABLE `message_routing_rules` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT, MODIFY `processor_id` BIGINT NOT NULL;
ALTER TABLE `system_configs` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT;
ALTER TABLE `message_queue` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT;
ALTER TABLE `connection_logs` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT, MODIFY `channel_id` BIGINT NOT NULL;
ALTER TABLE `api_call_logs` MODIFY `id` BIGINT NOT NULL AUTO_INCREMENT, MODIFY `channel_id` BIGINT, MODIFY `processor_id` BIGINT;

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
