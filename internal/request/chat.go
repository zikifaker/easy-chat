package request

type ChatRequest struct {
	Username  string `json:"username" binding:"required"`
	SessionID string `json:"session_id" binding:"required"`
	Query     string `json:"query" binding:"required"`
	Model     string `json:"model"`
	Mode      string `json:"mode"`
}
