package model

func (Session) TableName() string {
	return "sessions"
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

func (SecurityEvent) TableName() string {
	return "security_events"
}
