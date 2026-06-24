package hardware

import "runtime"

type Specs struct {
	CPUThreads int    `json:"cpu_threads"`
	TotalRAMGB float64 `json:"total_ram_gb"`
}

func Detect() Specs {
	s := Specs{
		CPUThreads: runtime.NumCPU(),
		TotalRAMGB: detectRAMGB(),
	}
	return s
}
