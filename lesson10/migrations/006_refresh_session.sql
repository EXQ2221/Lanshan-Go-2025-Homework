CREATE TABLE refresh_sessions (
                                  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
                                  sid CHAR(36) NOT NULL,                     -- refresh token 的会话ID(jti/uuid)
                                  user_id BIGINT UNSIGNED NOT NULL,
                                  token_version INT NOT NULL,                -- 与用户表 token_version 对齐
                                  refresh_token_hash CHAR(64) NOT NULL,      -- 建议存 SHA-256(hex)，不存明文
                                  expires_at DATETIME NOT NULL,
                                  revoked_at DATETIME NULL,
                                  replaced_by_sid CHAR(36) NULL,             -- 轮换后指向新 sid
                                  created_ip VARCHAR(45) NULL,
                                  created_ua VARCHAR(255) NULL,
                                  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

                                  PRIMARY KEY (id),
                                  UNIQUE KEY uk_refresh_sid (sid),
                                  KEY idx_refresh_user (user_id),
                                  KEY idx_refresh_user_revoked (user_id, revoked_at),
                                  KEY idx_refresh_expires (expires_at),

                                  CONSTRAINT fk_refresh_user
                                      FOREIGN KEY (user_id) REFERENCES users(id)
                                          ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
