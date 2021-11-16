package models

type Session struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	Status   string `json:"status"`
}

func NewSession(username string, token string) *Session {
	return &Session{Username: username, Token: token, Status: "active"}
}
