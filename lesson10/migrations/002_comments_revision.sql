DROP TABLE IF EXISTS answers;

ALTER TABLE comments DROP FOREIGN KEY fk_comments_parent;

ALTER TABLE comments DROP INDEX idx_comments_parent;
ALTER TABLE comments DROP COLUMN parent_id;

ALTER TABLE posts ADD INDEX idx_posts_is_deleted (is_deleted);
ALTER TABLE comments ADD INDEX idx_comments_is_deleted (is_deleted);