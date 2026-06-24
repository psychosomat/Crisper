package hardware

import (
	"os/exec"
	"strconv"
	"strings"
)

func detectRAMGB() float64 {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0
	}
	bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0
	}
	return float64(bytes) / (1024 * 1024 * 1024)
}
