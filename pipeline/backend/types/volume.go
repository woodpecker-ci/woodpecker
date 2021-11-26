package types

// Volume defines a container volume.
type Volume struct {
	Name       string            `json:"name,omitempty"`
	Driver     string            `json:"driver,omitempty"`
	DriverOpts map[string]string `json:"driver_opts,omitempty"`
}
