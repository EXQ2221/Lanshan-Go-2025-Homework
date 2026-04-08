package kafka

// Topic 常量，统一管理
const (
	TopicSecurityEvent = "auth.security.events"
)

// EventType 事件类型
// EventType 增加异地登录类型
const (
	EventTypeLogin            = "login"
	EventTypeLogout           = "logout"
	EventTypeLogoutAll        = "logout_all"
	EventTypeTokenRefresh     = "token_refresh"
	EventTypeTokenRevoke      = "token_revoke"
	EventTypeLoginFailed      = "login_failed"
	EventTypeAbnormalGeoLogin = "abnormal_geo_login" // 异地登录
)

// SecurityEvent 增加城市字段
type SecurityEvent struct {
	EventType string `json:"event_type"`
	UserID    int64  `json:"user_id"`
	SessionID string `json:"session_id"`
	IP        string `json:"ip"`
	DeviceID  string `json:"device_id"`
	UserAgent string `json:"user_agent"`
	Detail    string `json:"detail,omitempty"`
	OccurAt   int64  `json:"occur_at"`

	// 异地登录相关字段
	CurrentCity  string `json:"current_city,omitempty"`
	PreviousCity string `json:"previous_city,omitempty"`
	PreviousIP   string `json:"previous_ip,omitempty"`
	PreviousTime int64  `json:"previous_time,omitempty"`
}
