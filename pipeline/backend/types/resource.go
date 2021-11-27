package types

// Resources defines pod resources limits.
type Resources struct {
	CPULimit    string `json:"cpu_limit"`
	MemoryLimit string `json:"mem_limit"`
}
