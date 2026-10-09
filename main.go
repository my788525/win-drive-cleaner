package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// 版本
const version = "1.2.0"

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
	scanComms := flag.Bool("scan-comms", false, "扫描通讯软件旧文件(默认30天前，只统计)")
	scanMedia := flag.Bool("scan-media", false, "扫描视频/录屏候选(只提示大小)")
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
	case *scanComms:
		CommsFlowReadOnly(30)
	case *scanMedia:
		MediaFlowReadOnly()
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
		// 一键清空回收站已在深度项完成；这里把"删不掉的残留临时文件"压缩成一个 zip，供用户手动转移/删除
		if packPath, n, _ := PackLeftovers(); n > 0 {
			fmt.Println()
			fmt.Printf("  已把 %d 个删不掉的残留临时文件压缩为: %s\n", n, packPath)
			fmt.Println("  （可手动转移或删除该压缩包；压缩包本身不计入释放空间）")
		}
		fmt.Println()
		fmt.Println("  深度清理已完成。以下可选进阶项（需管理员）：")
		fmt.Println("   ① DISM 组件清理：DISM /Online /Cleanup-Image /StartComponentCleanup")
		fmt.Println("      （清理 WinSxS 组件备份，通常可再释放数百 MB 到数 GB，但耗时较长）")
		fmt.Println("   ② 通讯软件旧文件 / 视频录屏：请在交互菜单选对应功能")
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
	fmt.Println("   5) 清理通讯软件旧文件(微信/QQ/企业微信)")
	fmt.Println("   6) 查看 视频/录屏(仅提示大小，可打开所在文件夹手动删)")
	fmt.Println("   7) DISM 组件清理(可选进阶)")
	fmt.Println("   q) 退出")
	fmt.Print("  请输入: ")

	var choice string
	_, _ = fmt.Scanln(&choice)
	switch choice {
	case "1":
		execute(false, skipUAC)
		return
	case "2":
		execute(true, skipUAC)
		return
	case "3":
		DryRun = true
		execute(false, skipUAC)
		return
	case "4":
		DryRun = true
		execute(true, skipUAC)
		return
	case "5":
		commsFlow()
	case "6":
		mediaFlow()
	case "7":
		RunDismCleanup(false)
	}
	// 交互菜单退出前也停一次，避免闪退（非提权、非 --no-pause）
	if !noPause {
		pause()
	}
}

// commsFlow 通讯软件清理流程：先按天数扫描显示大小，再确认删除。
func commsFlow() {
	days := askInt("  清理多少天前的通讯软件文件？(默认 30, 0=全部旧文件也保留只统计): ", 30)
	if days < 0 {
		days = 0
	}
	PrintCommsScan(days)
	if days == 0 {
		fmt.Println("  [0] 只统计，不删除。")
		return
	}
	if !askYesNo(fmt.Sprintf("  确定删除 %d 天前的通讯软件文件吗？", days), false) {
		fmt.Println("  已取消，未删除任何通讯文件。")
		return
	}
	for _, t := range commsTargets() {
		b, n, e := deleteCommsOlder(t, days)
		if b == 0 && n == 0 {
			continue
		}
		fmt.Printf("   %-18s 已删除 %d 个文件，释放 %s%s\n", t.Name, n, formatBytes(b), errNote(int(e)))
	}
}

// mediaFlow 视频/录屏流程：显示大小，用户可选打开所在文件夹手动清理。
func mediaFlow() {
	printMediaScan(0) // 默认 Top 10
	if len(mediaIndex) == 0 {
		return
	}
	for {
		fmt.Println("  输入 M 序号(1-10)打开所在文件夹，o=全部仅查看，q=返回")
		var c string
		_, _ = fmt.Scanln(&c)
		c = strings.ToLower(strings.TrimSpace(c))
		if c == "q" {
			return
		}
		if c == "o" {
			// 打开第一个候选所在目录（仅示例）
			if len(mediaIndex) > 0 {
				_ = OpenInExplorer(filepath.Dir(mediaIndex[0].Path))
			}
			continue
		}
		n, err := strconv.Atoi(c)
		if err != nil {
			continue
		}
		openMediaFolder(n)
	}
}

// errNote 当有错误时返回简短提示，否则返回空串。
func errNote(e int) string {
	if e > 0 {
		return fmt.Sprintf("（%d 项因占用未删）", e)
	}
	return ""
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
	fmt.Println("  --scan-comms          扫描通讯软件(微信/QQ/企业微信)旧文件, 只统计")
	fmt.Println("  --scan-media          扫描视频/录屏候选, 只提示大小")
	fmt.Println("  --version             打印版本")
	fmt.Println("  --help                本帮助")
	fmt.Println()
	fmt.Println("交互菜单另有:")
	fmt.Println("   5) 清理通讯软件旧文件(可指定 N 天)   6) 查看视频/录屏")
	fmt.Println("   7) DISM 组件清理(可选进阶)")
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

// CommsFlowReadOnly 只读扫描通讯软件旧文件（--scan-comms）。
func CommsFlowReadOnly(days int) {
	PrintCommsScan(days)
}

// MediaFlowReadOnly 只读扫描视频/录屏候选（--scan-media）。
func MediaFlowReadOnly() {
	printMediaScan(0)
}
