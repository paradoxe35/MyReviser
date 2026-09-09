//go:build !darwin && !windows

package stt

// Linux reads /proc/meminfo; anything else declines to guess.
func sysMemoryMB() int { return 0 }
