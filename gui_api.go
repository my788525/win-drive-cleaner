//go:build windows

package main

import (
	"encoding/json"

	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

// diskUsage 返回系统盘剩余/总字节。
func diskUsage() (free, total int64) {
	path := systemDrive() // 例如 C:
	var fr, tot, avail uint64
	if err := windows.GetDiskFreeSpaceEx(windows.StringToUTF16Ptr(path), &fr, &tot, &avail); err == nil {
		return int64(fr), int64(tot)
	}
	return freeSpaceOnSystemDrive(), 0
}

// apiDiskInfo 供前端读取磁盘与管理员状态。
func apiDiskInfo() (map[string]interface{}, error) {
	free, total := diskUsage()
	return map[string]interface{}{
		"free":  free,
		"total": total,
		"admin": isAdmin(),
	}, nil
}

// apiStartupList 返回全部开机自启项视图。
func apiStartupList() ([]map[string]interface{}, error) {
	items := enumerateStartups()
	admin := isAdmin()
	out := make([]map[string]interface{}, 0, len(items))
	for _, e := range items {
		out = append(out, map[string]interface{}{
			"id":       e.ID,
			"name":     e.Name,
			"target":   e.Target,
			"source":   e.Source.String(),
			"enabled":  e.Enabled,
			"needAdmin": e.Source.isLMHive() && !admin,
		})
	}
	return out, nil
}

// apiStartupToggle 对单条自启项执行 启用/禁用。
func apiStartupToggle(id string, on bool) (map[string]interface{}, error) {
	items := enumerateStartups()
	for i := range items {
		if items[i].ID == id {
			res := setValueEnabled(&items[i], on)
			return map[string]interface{}{"ok": res.OK, "msg": res.Msg, "error": res.Error}, nil
		}
	}
	return map[string]interface{}{"ok": false, "msg": "", "error": "未找到该自启项"}, nil
}

// apiCommsScan 通讯软件缓存（只读，含各目录预估大小）。
func apiCommsScan() ([]map[string]interface{}, error) {
	targets := commsTargets()
	out := make([]map[string]interface{}, 0, len(targets))
	for _, t := range targets {
		var total int64
		for _, root := range t.Roots {
			_, b := walkSize(root)
			total += b
		}
		out = append(out, map[string]interface{}{
			"name":  t.Name,
			"roots": t.Roots,
			"bytes": total,
		})
	}
	return out, nil
}

// apiCacheScan AppData 软件缓存（≥10MB）。
func apiCacheScan() ([]map[string]interface{}, error) {
	entries := scanCache(10 << 20)
	out := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		out = append(out, map[string]interface{}{
			"software": e.Software,
			"path":     e.Path,
			"bytes":    e.Bytes,
		})
	}
	return out, nil
}

// apiCacheDelete 删除选中序号对应的软件缓存目录（序号对齐 apiCacheScan 的返回顺序）。
func apiCacheDelete(indices []int) (map[string]interface{}, error) {
	entries := scanCache(10 << 20)
	var deleted int
	var total int64
	for _, i := range indices {
		if i < 0 || i >= len(entries) {
			continue
		}
		e := entries[i]
		res := &Result{Name: "软件缓存", Path: e.Path}
		res.deleteTree(e.Path)
		total += res.Bytes
		deleted += res.Deleted
	}
	return map[string]interface{}{"deleted": deleted, "bytes": total}, nil
}

// apiBigFiles 大文件检测（≥500MB, Top 20）。
func apiBigFiles() ([]map[string]interface{}, error) {
	list := scanBigFiles(500, 20)
	out := make([]map[string]interface{}, 0, len(list))
	for _, f := range list {
		out = append(out, map[string]interface{}{"path": f.Path, "bytes": f.Bytes})
	}
	return out, nil
}

// apiMediaScan 视频/录屏媒体文件。
func apiMediaScan() ([]map[string]interface{}, error) {
	list := scanMedia()
	out := make([]map[string]interface{}, 0, len(list))
	for _, m := range list {
		out = append(out, map[string]interface{}{"path": m.Path, "size": m.Size})
	}
	return out, nil
}

// apiOpenFolder 在资源管理器中打开指定路径。
func apiOpenFolder(path string) error {
	return OpenInExplorer(path)
}

// apiDism / apiCompact 系统工具（需管理员）。
func apiDism(dry bool) (map[string]interface{}, error) {
	RunDismCleanup(dry)
	return map[string]interface{}{"msg": "DISM 清理完成"}, nil
}

func apiCompact(dry bool) (map[string]interface{}, error) {
	RunCompact(dry)
	return map[string]interface{}{"msg": "系统盘压缩完成"}, nil
}

// doClean 在后台 goroutine 执行清理，并实时把进度推送到前端。
func doClean(w webview2.WebView, deep, dry bool) {
	DryRun = dry
	OnProgress = func(seq, total int, it CleanItem, res *Result) {
		b, _ := json.Marshal(map[string]interface{}{
			"seq":    seq,
			"total":  total,
			"name":   it.Name,
			"path":   it.Path,
			"deleted": res.Deleted,
			"bytes":  res.Bytes,
			"errors": res.Errors,
		})
		js := "window.__onCleanProgress(" + string(b) + ")"
		w.Dispatch(func() { w.Eval(js) })
	}
	results := runItems(deep)
	OnProgress = nil
	deleted, errs, bytes := totalsOf(results)
	b, _ := json.Marshal(map[string]interface{}{
		"deleted": deleted,
		"errs":    errs,
		"bytes":   bytes,
		"dry":     dry,
	})
	js := "window.__onCleanDone(" + string(b) + ")"
	w.Dispatch(func() { w.Eval(js) })
}
