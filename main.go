package main

import (
	"flag"
	"fmt"
	"os"
)

// 版本
const version = "1.0.0"

// 运行参数（由主流程填充）
var (
	noUAC    bool
	elevated bool
)

func main() {
	stdFlag := flag.Bool("standard", false, "执行普通清理")
	deepFlag := flag.Bool("deep", false, "执行深度清理")
	dryRunFlag := flag.Bool("dry-run", false, "演练模式：只统计，不删除")
	verFlag := flag.Bool("version", false, "打印版本")
	noUACFlag := flag.Bool("no-uac", false, "跳过 UAC 提权（以当前权限运行）")
	elevatedFlag := flag.Bool("elevated", false, "内部标记：已提权进程，结束时暂停")
	showHelp := flag.Bool("h", false, "显示帮助")
	flag.Parse()

	noUAC = *noUACFlag
	elevated = *elevatedFlag
	stand, deep, dry, ver, help := *stdFlag, *deepFlag, *dryRunFlag, *verFlag, *showHelp

	switch {
	case ver:
		printVersion()
	case help:
		printHelp()
	case deep:
		DryRun = dry
		execute(true, noUAC)
	case stand:
		DryRun = dry
		execute(false, noUAC)
	case dry:
		DryRun = true
		execute(false, noUAC)
	default:
		interactive(noUAC)
	}
}

// execute 按模式执行清理（含 UAC 提权判断）。
func execute(includeDeep, skipUAC bool) {
	// 深度清理需管理员；普通清理尽量无需提权，但系统级目录可能需要。
	needElev := includeDeep
	if !skipUAC && needElev && !isAdmin() {
		fmt.Println("  深度清理需要管理员权限，正在请求 UAC 提权…")
		if err := elevateSelf(); err != nil {
			fmt.Println("  提权失败:", err)
			os.Exit(2)
		}
		return // 提权成功后，由新进程继续
	}

	fmt.Printf(" [%s] %s\n", nowStamp(), osDescription())
	if DryRun {
		fmt.Println("  [演练模式] 不会真正删除任何文件")
	}
	fmt.Println()

	results := runItems(includeDeep)
	printReport(os.Stdout, results, includeDeep, DryRun)

	if !DryRun && includeDeep {
		fmt.Println()
		fmt.Println("  提示: 追加运行 DISM 可进一步释放空间（可选、需管理员、耗时较长）：")
		fmt.Println("      DISM /Online /Cleanup-Image /StartComponentCleanup")
	}

	// 提权重启的进程结束后暂停，避免窗口一闪而过
	if elevated {
		fmt.Println()
		fmt.Print("  按回车键退出…")
		fmt.Scanln()
	}
}

func interactive(skipUAC bool) {
	fmt.Println()
	fmt.Println("================ Win 盘清理工具 ================")
	fmt.Printf(" 系统: %s\n", osDescription())
	fmt.Printf(" 权限: %s\n", adminStateLabel())
	fmt.Println()
	fmt.Println("  选择模式:")
	fmt.Println("   1) 普通清理")
	fmt.Println("   2) 深度清理(需管理员)")
	fmt.Println("   3) 演练普通清理(不删除)")
	fmt.Println("   4) 演练深度清理(不删除)")
	fmt.Println("   q) 退出")
	fmt.Print("  请输入: ")

	var choice string
	fmt.Scanln(&choice)
	switch choice {
	case "1":
		execute(false, skipUAC)
	case "2":
		execute(true, skipUAC)
	case "3":
		DryRun = true
		execute(false, skipUAC)
	case "4":
		DryRun = true
		execute(true, skipUAC)
	}
}

func printVersion() {
	fmt.Println("Win 盘清理工具  v" + version)
}

func printHelp() {
	fmt.Println("Win 盘清理工具  v" + version)
	fmt.Println("  安全清理 Windows C 盘垃圾，零安装、绿色单文件，兼容 Windows 8/10/11。")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  直接运行              进入交互菜单")
	fmt.Println("  --standard           普通清理(低风险)")
	fmt.Println("  --deep               深度清理(需管理员, 释放空间更多)")
	fmt.Println("  --dry-run            演练: 只统计大小, 不删除")
	fmt.Println("  --no-uac             跳过 UAC 提权")
	fmt.Println("  --version            打印版本")
	fmt.Println("  --help               本帮助")
}

func adminStateLabel() string {
	if isAdmin() {
		return "管理员"
	}
	return "普通用户"
}
