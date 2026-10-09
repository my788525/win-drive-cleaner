package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// OpenInExplorer 用资源管理器打开目标；文件则定位选中，目录则打开。
func OpenInExplorer(path string) error {
	if info, err := os.Lstat(path); err == nil && !info.IsDir() {
		return exec.Command("explorer.exe", "/select", path).Start()
	}
	return exec.Command("explorer.exe", path).Start()
}

// askInt 读取一个正整数；解析失败或 0 时返回 def。
func askInt(prompt string, def int) int {
	fmt.Print(prompt)
	var s string
	_, _ = fmt.Scanln(&s)
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// askYesNo 读取 y/n；默认返回 def（空回车视为默认）。
func askYesNo(prompt string, def bool) bool {
	fmt.Print(prompt + " [y/N] ")
	var s string
	_, _ = fmt.Scanln(&s)
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return def
	}
	return s == "y" || s == "yes" || s == "是"
}

// uniqueStrings 去掉重复并保持顺序。
func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
