package types

// Conn defines a container network connection.
type Conn struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}
