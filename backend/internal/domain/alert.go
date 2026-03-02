package domain

type IntegrityAlert struct {
	SessionID string   `json:"session_id"`
	Score     int      `json:"score"`
	Flags     []string `json:"flags"`
	Type      string   `json:"type"` // RISK_ALERT
}