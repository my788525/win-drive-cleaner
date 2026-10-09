package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// compactGo 系统盘压缩（Windows 10/11 自带 LZX 压缩）。
// 把 C:\Windows 内文件就地 LZX 压缩存储，通常释放 1-3GB（机械硬盘收益最大）。
// 可逆：compact /decompress 可恢复。需管理员。
// 注意：只压缩系统文件，不影响用户数据；SSD 会轻微影响随机读寿命。

// compactSupported 检测 compact /Compact 是否可用（Win10/11 有，Win8/7 没有 LZX 系统压缩）。
func compactSupported() bool {
	out, err := exec.Command("compact", "/Compact", "/?").CombinedOutput()
	if err != nil {
		return false
	}
	// 可用时会显示 help，含 "LZX" 或 "compress"
	s := strings.ToLower(string(out))
	return strings.Contains(s, "lzx") || strings.Contains(s, "compress")
}

// RunCompact 交互式执行系统盘压缩。dryRun 时只打印将做什么。
func RunCompact(dryRun bool) {
	fmt.Println()
	fmt.Println("  ┌─ 系统盘压缩（可选进阶）─")
	fmt.Println("  │ 命令: compact /Compact /BaseFile:C:\\compact.sys /EssentialDirectories")
	fmt.Println("  │ 效果: 把 C:\\Windows 内文件就地 LZX 压缩，通常释放 1-3GB（机械硬盘收益最大）。")
	fmt.Println("  │ 可逆: 日后运行 compact /decompress 可恢复原状。")

	if !compactSupported() {
		fmt.Println("  └ [不支持] 本系统（Win8/Win7）没有 LZX 系统压缩，跳过。")
		return
	}

	if dryRun {
		fmt.Println("  └ [演练] 未真正执行。")
		return
	}

	if !isAdmin() {
		fmt.Println("  系统盘压缩需要管理员权限。请：")
		fmt.Println("   - 直接双击本工具并选择“以管理员身份运行”，或")
		fmt.Println("   - 在交互菜单先选 2) 深度清理 触发 UAC 提权后再回来")
		return
	}

	if !askYesNo("  确定压缩系统盘（可能需数分钟，期间勿强制断电）?", false) {
		fmt.Println("  已取消系统盘压缩。")
		return
	}

	fmt.Println("  正在压缩系统文件，请稍候（首次约 1-5 分钟）…")
	cmd := exec.Command("compact", "/Compact", "/BaseFile:C:\\compact.sys", "/EssentialDirectories")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("  compact 执行完成或有警告:", err)
	} else {
		fmt.Println("  系统盘压缩完成。")
	}
}

// CompactFlowReadOnly 只读（--scan-compact 用）：只检测是否支持，不执行。
func CompactFlowReadOnly() {
	if compactSupported() {
		fmt.Println("  本系统支持 LZX 系统压缩（compact /Compact）。在交互菜单 11) 可执行。")
	} else {
		fmt.Println("  本系统不支持 LZX 系统压缩（Win8/7 无此功能）。")
	}
}
