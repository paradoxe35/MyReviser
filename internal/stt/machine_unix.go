//go:build darwin

package stt

import (
	"os/exec"
	"strconv"
	"strings"
)

func sysMemoryMB() int {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0
	}
	bytes, err := strconv.ParseInt(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0
	}
	return int(bytes / (1 << 20))
}
