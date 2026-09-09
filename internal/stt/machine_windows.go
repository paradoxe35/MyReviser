//go:build windows

package stt

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

func sysMemoryMB() int {
	var status windows.MemoryStatusEx
	status.Length = uint32(unsafe.Sizeof(status))
	if err := windows.GlobalMemoryStatusEx(&status); err != nil {
		return 0
	}
	return int(status.TotalPhys / (1 << 20))
}
