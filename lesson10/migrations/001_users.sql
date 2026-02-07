CREATE TABLE users (
                       id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                       username VARCHAR(64) NOT NULL,
                       password_hash VARCHAR(255) NOT NULL,
                       token_version INT NOT NULL DEFAULT 0,

                       avatar_url VARCHAR(255) NULL,
                       profile VARCHAR(255) NULL,

                       role TINYINT UNSIGNED NOT NULL DEFAULT 0, -- 0=normal 1=vip 2=admin
                       vip_expires_at DATETIME NULL,

                       created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                       PRIMARY KEY (id),
                       UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE posts (
                       id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                       type TINYINT UNSIGNED NOT NULL,          -- 1=article 2=question
                       author_id BIGINT UNSIGNED NOT NULL,

                       title VARCHAR(200) NOT NULL,
                       content LONGTEXT NOT NULL,               -- markdown，图片用 ![](url) 方式插入即可
                       is_deleted TINYINT UNSIGNED NOT NULL DEFAULT 0,

                       created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                       PRIMARY KEY (id),
                       KEY idx_posts_author_created (author_id, created_at),
                       KEY idx_posts_type_created (type, created_at),
                       FULLTEXT KEY ft_posts_title_content (title, content),

                       CONSTRAINT fk_posts_author FOREIGN KEY (author_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE answers (
                         id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                         question_id BIGINT UNSIGNED NOT NULL,    -- 指向 posts.id 且 posts.type=2
                         author_id BIGINT UNSIGNED NOT NULL,

                         content LONGTEXT NOT NULL,
                         is_deleted TINYINT UNSIGNED NOT NULL DEFAULT 0,

                         created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
                         updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                         PRIMARY KEY (id),
                         KEY idx_answers_question_created (question_id, created_at),
                         KEY idx_answers_author_created (author_id, created_at),
                         FULLTEXT KEY ft_answers_content (content),

                         CONSTRAINT fk_answers_question FOREIGN KEY (question_id) REFERENCES posts(id),
                         CONSTRAINT fk_answers_author FOREIGN KEY (author_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE comments (
                          id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                          target_type TINYINT UNSIGNED NOT NULL,   -- 1=post 2=answer
                          target_id BIGINT UNSIGNED NOT NULL,

                          author_id BIGINT UNSIGNED NOT NULL,
                          parent_id BIGINT UNSIGNED NULL,          -- 回复某条评论(可空)
                          content TEXT NOT NULL,

                          is_deleted TINYINT UNSIGNED NOT NULL DEFAULT 0,
                          created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                          PRIMARY KEY (id),
                          KEY idx_comments_target_created (target_type, target_id, created_at),
                          KEY idx_comments_author_created (author_id, created_at),
                          KEY idx_comments_parent (parent_id),

                          CONSTRAINT fk_comments_author FOREIGN KEY (author_id) REFERENCES users(id),
                          CONSTRAINT fk_comments_parent
                              FOREIGN KEY (parent_id) REFERENCES comments(id)
                                  ON DELETE SET NULL

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE post_images (
                             id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                             post_id BIGINT UNSIGNED NOT NULL,
                             uploader_id BIGINT UNSIGNED NOT NULL,

                             url VARCHAR(512) NOT NULL,
                             created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                             PRIMARY KEY (id),
                             KEY idx_images_post (post_id),
                             KEY idx_images_uploader (uploader_id),

                             CONSTRAINT fk_images_post FOREIGN KEY (post_id) REFERENCES posts(id),
                             CONSTRAINT fk_images_uploader FOREIGN KEY (uploader_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE user_follows (
                              id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                              follower_id BIGINT UNSIGNED NOT NULL,
                              followee_id BIGINT UNSIGNED NOT NULL,
                              created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                              PRIMARY KEY (id),
                              UNIQUE KEY uk_user_follow_pair (follower_id, followee_id),
                              KEY idx_user_follows_followee (followee_id, created_at),

                              CONSTRAINT fk_user_follows_follower FOREIGN KEY (follower_id) REFERENCES users(id),
                              CONSTRAINT fk_user_follows_followee FOREIGN KEY (followee_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE question_follows (
                                  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                  user_id BIGINT UNSIGNED NOT NULL,
                                  question_id BIGINT UNSIGNED NOT NULL,
                                  created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                                  PRIMARY KEY (id),
                                  UNIQUE KEY uk_question_follow (user_id, question_id),
                                  KEY idx_qf_question (question_id, created_at),

                                  CONSTRAINT fk_qf_user FOREIGN KEY (user_id) REFERENCES users(id),
                                  CONSTRAINT fk_qf_question FOREIGN KEY (question_id) REFERENCES posts(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE reactions (
                           id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                           user_id BIGINT UNSIGNED NOT NULL,

                           target_type TINYINT UNSIGNED NOT NULL,   -- 1=post 2=answer 3=comment
                           target_id BIGINT UNSIGNED NOT NULL,

                           created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                           PRIMARY KEY (id),
                           UNIQUE KEY uk_reaction (user_id, target_type, target_id),
                           KEY idx_reactions_target (target_type, target_id, created_at),

                           CONSTRAINT fk_reactions_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE favorites (
                           id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                           user_id BIGINT UNSIGNED NOT NULL,

                           target_type TINYINT UNSIGNED NOT NULL,   -- 1=post 2=answer
                           target_id BIGINT UNSIGNED NOT NULL,

                           created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                           PRIMARY KEY (id),
                           UNIQUE KEY uk_favorite (user_id, target_type, target_id),
                           KEY idx_favorites_target (target_type, target_id, created_at),

                           CONSTRAINT fk_favorites_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE activities (
                            id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                            actor_id BIGINT UNSIGNED NOT NULL,

                            action TINYINT UNSIGNED NOT NULL,        -- 1=post 2=answer 3=comment 4=like 5=favorite 6=follow_user 7=follow_question
                            target_type TINYINT UNSIGNED NOT NULL,   -- 1=post 2=answer 3=comment 4=user 5=question(post)
                            target_id BIGINT UNSIGNED NOT NULL,

                            created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                            PRIMARY KEY (id),
                            KEY idx_activities_actor_created (actor_id, created_at),

                            CONSTRAINT fk_activities_actor FOREIGN KEY (actor_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE notifications (
                               id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                               user_id BIGINT UNSIGNED NOT NULL,        -- 接收者

                               type TINYINT UNSIGNED NOT NULL,          -- 1=comment 2=answer 3=follow_question 4=dm(可选)
                               actor_id BIGINT UNSIGNED NULL,           -- 触发者
                               target_type TINYINT UNSIGNED NULL,       -- 1=post 2=answer 3=comment
                               target_id BIGINT UNSIGNED NULL,

                               content VARCHAR(255) NOT NULL,
                               is_read TINYINT UNSIGNED NOT NULL DEFAULT 0,

                               created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                               PRIMARY KEY (id),
                               KEY idx_notifications_user_read (user_id, is_read, created_at),

                               CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE conversations (
                               id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                               created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
                               PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE conversation_members (
                                      id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                      conversation_id BIGINT UNSIGNED NOT NULL,
                                      user_id BIGINT UNSIGNED NOT NULL,

                                      created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                                      PRIMARY KEY (id),
                                      UNIQUE KEY uk_conv_member (conversation_id, user_id),
                                      KEY idx_conv_member_user (user_id),

                                      CONSTRAINT fk_cm_conv FOREIGN KEY (conversation_id) REFERENCES conversations(id),
                                      CONSTRAINT fk_cm_user FOREIGN KEY (user_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE messages (
                          id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                          conversation_id BIGINT UNSIGNED NOT NULL,
                          sender_id BIGINT UNSIGNED NOT NULL,
                          content TEXT NOT NULL,
                          created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,

                          PRIMARY KEY (id),
                          KEY idx_messages_conv_created (conversation_id, created_at),

                          CONSTRAINT fk_msg_conv FOREIGN KEY (conversation_id) REFERENCES conversations(id),
                          CONSTRAINT fk_msg_sender FOREIGN KEY (sender_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
