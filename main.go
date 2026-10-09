package main

import (
	"flag"
	"fmt"
	"os"
)

// 版本
const version = "1.1.0"

// 运行参数（由主流程填充）
var (
	noUAC    bool
	elevated bool
	noPause  bool // 不暂停，直接退出（脚本/CI 场景）
)

func main() {
	stdFlag := flag.Bool("standard", false, "执行普通清理")
	deepFlag := flag.Bool("deep", false, "执行深度清理")
	dryRunFlag := flag.Bool("dry-run", false, "演练模式：只统计，不删除")
	verFlag := flag.Bool("version", false, "打印版本")
	noUACFlag := flag.Bool("no-uac", false, "跳过 UAC 提权（以当前权限运行）")
	elevatedFlag := flag.Bool("elevated", false, "内部标记：已提权子进程，结束不暂停")
	noPauseFlag := flag.Bool("no-pause", false, "不暂停直接退出（脚本/批处理场景）")
	showHelp := flag.Bool("h", false, "显示帮助")
	flag.Parse()

	noUAC = *noUACFlag
	elevated = *elevatedFlag
	noPause = *noPauseFlag
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
	// 深度清理需管理员；非管理员时自动请求 UAC 提权
	if includeDeep && !skipUAC && !isAdmin() {
		fmt.Println("  深度清理需要管理员权限，正在请求 UAC 提权…")
		if err := elevateSelf(); err != nil {
			fmt.Println("  提权失败:", err)
			// 提权失败也要暂停，让用户看到原因
			pause()
			os.Exit(2)
		}
		// 提权成功：父进程结束，子进程（elevated=true）接管并负责暂停
		return
	}

	fmt.Printf(" [%s] %s\n", nowStamp(), osDescription())
	if DryRun {
		fmt.Println("  [演练模式] 不会真正删除任何文件")
	}
	fmt.Println()

	// 清理前剩余空间（真实值，用于前后对比）
	freeBefore := freeSpaceOnSystemDrive()

	results := runItems(includeDeep)
	printReport(os.Stdout, results, includeDeep, DryRun)

	// 清理后剩余空间 + 对比
	if !DryRun {
		freeAfter := freeSpaceOnSystemDrive()
		if freeBefore >= 0 && freeAfter >= 0 {
			fmt.Fprintf(os.Stdout, "  磁盘剩余空间: %s  ->  %s\n", formatBytes(freeBefore), formatBytes(freeAfter))
			if freeAfter > freeBefore {
				fmt.Fprintf(os.Stdout, "  实测释放: %s\n", formatBytes(freeAfter-freeBefore))
			} else {
				fmt.Fprintf(os.Stdout, "  （文件删除后磁盘计数可能略有延迟，数值以“已释放”统计为准）\n")
			}
		}
	}

	if !DryRun && includeDeep {
		fmt.Println()
		fmt.Println("  提示: 追加运行 DISM 可进一步释放空间（可选、需管理员、耗时较长）：")
		fmt.Println("      DISM /Online /Cleanup-Image /StartComponentCleanup")
	}

	// 走到这里说明本进程真正执行了清理/演练（父进程提权成功后会在上面 return，不会到这里）
	// 因此默认暂停，方便看清结果；仅 --no-pause 时自动结束
	if !noPause {
		pause()
	}
}

// pause 停住等待回车，避免窗口一闪而过。
func pause() {
	fmt.Println()
	fmt.Print("  按回车键退出…")
	var dummy string
	_, _ = fmt.Scanln(&dummy)
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
	// 交互菜单退出前也停一次，避免闪退（非提权、非 --no-pause）
	if !elevated && !noPause {
		pause()
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
	fmt.Println("  --standard            普通清理(低风险)")
	fmt.Println("  --deep                深度清理(需管理员, 释放空间更多)")
	fmt.Println("  --dry-run             演练: 只统计大小, 不删除")
	fmt.Println("  --no-uac              跳过 UAC 提权")
	fmt.Println("  --no-pause            不暂停直接退出(脚本/批处理场景)")
	fmt.Println("  --version             打印版本")
	fmt.Println("  --help                本帮助")
	fmt.Println()
	fmt.Println("说明: 每次清理结束后默认停在“按回车键退出”，方便看清结果；")
	fmt.Println("      脚本自动化时加 --no-pause 可自动结束。")
}

func adminStateLabel() string {
	if isAdmin() {
		return "管理员"
	}
	return "普通用户"
}
