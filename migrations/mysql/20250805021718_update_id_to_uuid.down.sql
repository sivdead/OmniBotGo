-- migrate:down
-- E06: Refuse unsafe UUID → BIGINT rollback.
-- Converting VARCHAR(36) UUID primary/foreign keys back to BIGINT is not
-- lossless and historically required DROP FOREIGN KEY + type MODIFY that can
-- corrupt or truncate relational integrity. Operators must restore from backup
-- instead of running this down migration.

SIGNAL SQLSTATE '45000'
    SET MESSAGE_TEXT = 'UUID down migration refused: unsafe UUID→BIGINT rollback; restore from backup';
