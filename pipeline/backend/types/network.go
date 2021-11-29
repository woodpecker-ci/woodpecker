package types

// Network defines a container network.
type Network struct {
	Name       string            `json:"name,omitempty"`
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty"`
}
