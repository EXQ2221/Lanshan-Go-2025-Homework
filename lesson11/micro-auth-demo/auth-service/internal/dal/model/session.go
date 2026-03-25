package model

import "time"

type Session struct {
	ID                   int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID            string     `gorm:"size:64;uniqueIndex;not null" json:"session_id"`
	UserID               int64      `gorm:"index;not null" json:"user_id"`
	Status               string     `gorm:"size:32;index;not null" json:"status"`
	DeviceID             string     `gorm:"size:128;index" json:"device_id"`
	DeviceName           string     `gorm:"size:128" json:"device_name"`
	UserAgent            string     `gorm:"size:512" json:"user_agent"`
	LoginIP              string     `gorm:"size:64" json:"login_ip"`
	LastIP               string     `gorm:"size:64" json:"last_ip"`
	LastSeenAt           time.Time  `json:"last_seen_at"`
	CurrentAccessJTI     string     `gorm:"size:64;index" json:"current_access_jti"`
	CurrentAccessExpires time.Time  `json:"current_access_expires"`
	RevokedAt            *time.Time `json:"revoked_at,omitempty"`
	RevokeReason         string     `gorm:"size:128" json:"revoke_reason"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

func (Session) TableName() string {
	return "sessions"
}
