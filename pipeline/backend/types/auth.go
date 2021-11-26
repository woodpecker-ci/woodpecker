package types

// Auth defines registry authentication credentials.
type Auth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Email    string `json:"email,omitempty"`
}
