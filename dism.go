package main

import (
	"archive/zip"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
)

// RunDismCleanup 运行 DISM 组件清理（可选）。执行前打印固定提示并征求确认。
// dryRun 时只打印将执行什么，不真正运行。
func RunDismCleanup(dryRun bool) {
	fmt.Println()
	fmt.Println("  ┌─ DISM 组件清理（可选）─")
	fmt.Println("  │ 这会清理 WinSxS 组件备份，通常可再释放数百 MB 到数 GB，但耗时较长。")
	fmt.Println("  │ 命令: DISM /Online /Cleanup-Image /StartComponentCleanup")
	if dryRun {
		fmt.Println("  └ [演练] 未真正执行。")
		return
	}
	if !askYesNo("  确定执行 DISM 组件清理吗（可能需数分钟）?", false) {
		fmt.Println("  已取消 DISM 组件清理。")
		return
	}
	if !isAdmin() {
		fmt.Println("  DISM 需要管理员权限，请先以管理员身份运行本工具（或在菜单选深度清理触发提权）。")
		return
	}
	fmt.Println("  正在运行 DISM，请稍候…")
	cmd := exec.Command("DISM", "/Online", "/Cleanup-Image", "/StartComponentCleanup")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println("  DISM 执行失败或未完成:", err)
		return
	}
	fmt.Println("  DISM 组件清理完成。")
}

// PackLeftovers 把若干临时目录中"删不掉/仍残留"的小文件压缩成一个 zip，便于用户手动转移/删除。
// 典型用法：在深度清理删除后调用——此时目录里剩下的就是被占用或无权限的文件。
// 返回 (zip 路径, 打包条数, error)；无残留时路径为空。
func PackLeftovers() (string, int, error) {
	roots := []string{
		filepath.Join(systemRoot(), "Temp"),
		userTemp(),
		filepath.Join(allUsersProfile(), "Microsoft", "Windows", "WER"),
	}
	var keep []string
	for _, root := range roots {
		filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			// 只打包 50MB 以下的单个文件，避免超大文件
			if info.Size() <= 50<<20 {
				keep = append(keep, p)
			}
			return nil
		})
	}
	if len(keep) == 0 {
		return "", 0, nil
	}
	sort.Strings(keep)
	if len(keep) > 5000 {
		keep = keep[:5000]
	}

	ts := time.Now().Format("20060102_150405")
	outPath := filepath.Join(userTemp(), "CleanerPack_"+ts+".zip")
	f, err := os.Create(outPath)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	zw := zip.NewWriter(f)

	// 用相对子目录避免 zip 内文件名冲突
	count := 0
	for _, src := range keep {
		rel := filepath.Base(filepath.Dir(src)) + "_" + filepath.Base(src)
		w, err := zw.Create(rel)
		if err != nil {
			continue
		}
		if data, err := os.ReadFile(src); err == nil {
			_, _ = w.Write(data)
			count++
		}
	}
	_ = zw.Close()
	return outPath, count, nil
}
