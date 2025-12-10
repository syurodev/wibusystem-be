package user

// SessionResponse represents a user session
type SessionResponse struct {
	ID         string `json:"id"`
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	Device     string `json:"device"`
	ClientOS   string `json:"client_os"`
	Browser    string `json:"browser"`
	LastActive string `json:"last_active"`
	Current    bool   `json:"current"`
}
