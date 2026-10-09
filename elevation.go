package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// 通过纯标准库 syscall 调用 shell32，零外部依赖，Win8/10/11 通用。
var (
	shell32     = syscall.NewLazyDLL("shell32.dll")
	shellExecW  = shell32.NewProc("ShellExecuteW")
	isUserAdmin = shell32.NewProc("IsUserAnAdmin")
)

// isAdmin 判断当前进程是否以管理员权限运行。
func isAdmin() bool {
	r, _, _ := isUserAdmin.Call()
	return r != 0
}

// elevateSelf 以管理员身份重新启动自身，携带原有参数并追加 --elevated。
func elevateSelf() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法定位自身路径: %w", err)
	}

	var args []string
	for _, a := range os.Args[1:] {
		if a == "--no-uac" {
			continue
		}
		args = append(args, a)
	}
	args = append(args, "--elevated")

	verb := syscall.StringToUTF16Ptr("runas")
	file := syscall.StringToUTF16Ptr(exe)
	paramStr := syscall.StringToUTF16Ptr(strings.Join(args, " "))
	dir := syscall.StringToUTF16Ptr("")

	const (
		SE_ERR_OK     = 1
		SE_ERR_CANCEL = 5
	)
	ret, _, _ := shellExecW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(paramStr)),
		uintptr(unsafe.Pointer(dir)),
		1, // SW_SHOWNORMAL
	)

	if ret == uintptr(SE_ERR_CANCEL) {
		return fmt.Errorf("你已取消 UAC 权限请求")
	}
	if ret < uintptr(SE_ERR_OK) {
		return fmt.Errorf("启动提权进程失败 (code=%d)", ret)
	}
	return nil
}
