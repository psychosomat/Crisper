package hardware

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func detectRAMGB() float64 {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					return float64(kb) / (1024 * 1024)
				}
			}
			break
		}
	}
	return 0
}
