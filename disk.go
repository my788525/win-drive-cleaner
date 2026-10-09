package main

import (
	"syscall"
	"unsafe"
)

// 通过 kernel32.GetDiskFreeSpaceExW 取真实磁盘剩余空间，纯标准库，Win8–11 通用。
var (
	kernel32   = syscall.NewLazyDLL("kernel32.dll")
	diskFreeEx = kernel32.NewProc("GetDiskFreeSpaceExW")
)

type ularge struct{ low, high uint32 }

// freeSpaceOnSystemDrive 返回系统盘当前可用空间（字节）。失败返回 -1。
func freeSpaceOnSystemDrive() int64 {
	path := syscall.StringToUTF16Ptr(systemDrive())
	var total, avail, free ularge
	ret, _, _ := diskFreeEx.Call(
		uintptr(unsafe.Pointer(path)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&avail)),
		uintptr(unsafe.Pointer(&free)),
	)
	if ret == 0 {
		return -1
	}
	// 可用空间 = avail（Low + High<<32）
	return int64(avail.low) | int64(avail.high)<<32
}
