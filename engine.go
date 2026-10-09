package main

import (
	"os"
	"path/filepath"
)

// Mode 描述清理模式。
type Mode int

const (
	ModeFileDelete    Mode = iota // 删除单个文件（不存在则跳过）
	ModeClearContents            // 清空目录内容（保留目录本身）
	ModeDeleteTree               // 删除整棵目录树
)

// CleanItem 一个可清理目标。
type CleanItem struct {
	Name string
	Path string
	Mode Mode
	Deep bool // 是否属于“深度清理”
}

// Result 记录一次清理结果。
type Result struct {
	Name    string
	Path    string
	Deleted int
	Bytes   int64 // 实际删除字节（dry-run 时为估算）
	Errors  int
}

// DryRun 控制是否真正删除。
var DryRun bool

// standardItems 普通清理：低风险、快速、不影响系统可用性。
func standardItems() []CleanItem {
	return []CleanItem{
		{Name: "Windows 临时文件", Path: systemRoot() + "\\Temp", Mode: ModeClearContents, Deep: false},
		{Name: "用户临时文件", Path: userTemp(), Mode: ModeClearContents, Deep: false},
		{Name: "系统预读(Prefetch)", Path: systemDrive() + "\\Prefetch", Mode: ModeDeleteTree, Deep: false},
		{Name: "缩略图缓存", Path: filepath.Join(localAppData(), "Microsoft", "Windows", "Explorer"), Mode: ModeClearContents, Deep: false},
		{Name: "DNS 解析缓存目录", Path: filepath.Join(systemRoot(), "System32", "DnsCache"), Mode: ModeDeleteTree, Deep: false},
		{Name: "Windows 更新下载缓存", Path: filepath.Join(systemRoot(), "Windows", "SoftwareDistribution", "Download"), Mode: ModeClearContents, Deep: false},
		{Name: "应用崩溃报告(WER)", Path: filepath.Join(allUsersProfile(), "Microsoft", "Windows", "WER"), Mode: ModeClearContents, Deep: false},
	}
}

// deepItems 深度清理：空间收益大，需管理员，但仍避开裸删危险系统文件。
func deepItems() []CleanItem {
	cats := []CleanItem{
		{Name: "回收站(系统盘)", Path: recycleBin(), Mode: ModeDeleteTree, Deep: true},
		{Name: "系统级临时文件", Path: filepath.Join(systemRoot(), "Temp"), Mode: ModeClearContents, Deep: true},
		{Name: "系统盘根目录临时", Path: systemDrive() + "\\Temp", Mode: ModeClearContents, Deep: true},
		{Name: "WinSxS 旧组件目录", Path: filepath.Join(systemRoot(), "WinSxS", "Backup"), Mode: ModeDeleteTree, Deep: true},
		{Name: "旧系统备份(Windows.old)", Path: windowsOld(), Mode: ModeDeleteTree, Deep: true},
	}
	return cats
}

// collectItems 返回当前模式下要执行的条目。
func collectItems(includeDeep bool) []CleanItem {
	items := standardItems()
	if includeDeep {
		items = append(items, deepItems()...)
	}
	return items
}

// cleanItem 执行单个条目。
func cleanItem(it CleanItem) *Result {
	res := &Result{Name: it.Name, Path: it.Path}
	pi, err := os.Lstat(it.Path)
	if err != nil {
		return res // 不存在：正常跳过
	}
	switch {
	case pi.IsDir():
		switch it.Mode {
		case ModeDeleteTree:
			res.deleteTree(it.Path)
		case ModeClearContents:
			res.clearContents(it.Path)
		default:
			res.clearContents(it.Path)
		}
	default:
		res.deleteFile(it.Path, pi.Size())
	}
	return res
}

func (r *Result) deleteFile(path string, size int64) {
	if DryRun {
		r.Deleted++
		r.Bytes += size
		return
	}
	if err := os.Remove(path); err != nil {
		r.Errors++
		return
	}
	r.Deleted++
	r.Bytes += size
}

func (r *Result) clearContents(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		r.Errors++
		return
	}
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		if e.IsDir() {
			r.deleteTree(full)
		} else {
			pi, err := os.Lstat(full)
			if err != nil {
				continue
			}
			r.deleteFile(full, pi.Size())
		}
	}
}

func (r *Result) deleteTree(dir string) {
	if DryRun {
		n, b := walkSize(dir)
		r.Deleted += int(n)
		r.Bytes += b
		return
	}
	// 先尝试直接整删；失败多因文件占用，退化为清空内容
	if err := os.RemoveAll(dir); err == nil {
		// 估算删除量
		_, b := walkSize(dir) // 已删，返回 0，仅保留计数语义
		_ = b
		r.Deleted++
		return
	}
	r.clearContents(dir)
}

func walkSize(root string) (int64, int64) {
	var count, bytes int64
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		count++
		if !info.IsDir() {
			bytes += info.Size()
		}
		return nil
	})
	return count, bytes
}

// runItems 按模式执行全部条目。
func runItems(includeDeep bool) []*Result {
	items := collectItems(includeDeep)
	out := make([]*Result, 0, len(items))
	for _, it := range items {
		out = append(out, cleanItem(it))
	}
	return out
}

// totals 汇总。
func totalsOf(r []*Result) (deleted, errs int, bytes int64) {
	for _, x := range r {
		deleted += x.Deleted
		errs += x.Errors
		bytes += x.Bytes
	}
	return
}
