package hardware

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var kernel32 = windows.NewLazySystemDLL("kernel32.dll")
var globalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")

type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func detectRAMGB() float64 {
	var m memoryStatusEx
	m.Length = uint32(unsafe.Sizeof(m))
	ret, _, _ := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&m)))
	if ret == 0 {
		return 0
	}
	return float64(m.TotalPhys) / (1024 * 1024 * 1024)
}
