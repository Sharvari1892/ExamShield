package domain

type SyncRequest struct {
	SessionID string        `json:"session_id" binding:"required"`
	Answers   []SyncAnswer  `json:"answers" binding:"required"`
}

type SyncAnswer struct {
	QuestionID string `json:"question_id" binding:"required"`
	Answer     string `json:"answer" binding:"required"`
	Version    int    `json:"version" binding:"required"`
}

type SyncResponse struct {
	Accepted       []string       `json:"accepted"`
	Rejected       []string       `json:"rejected"`
	ServerVersions map[string]int `json:"server_versions"`
}