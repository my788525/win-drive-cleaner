//go:build windows

package main

import (
	_ "embed"

	"github.com/jchv/go-webview2"
)

//go:embed gui.html
var guiHTML string

// runGUI 启动 WebView2 原生窗口并承载图形界面。
func runGUI() {
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug: false,
		WindowOptions: webview2.WindowOptions{
			Title:  "Win 盘清理",
			Width:  1120,
			Height: 760,
			Center: true,
		},
	})
	if w == nil {
		return
	}
	defer w.Destroy()

	// 数据/操作接口（前端通过 window.apiX(...) 调用，返回 Promise）。
	_ = w.Bind("apiDiskInfo", apiDiskInfo)
	_ = w.Bind("apiStartupList", apiStartupList)
	_ = w.Bind("apiStartupToggle", apiStartupToggle)
	_ = w.Bind("apiCommsScan", apiCommsScan)
	_ = w.Bind("apiCacheScan", apiCacheScan)
	_ = w.Bind("apiCacheDelete", apiCacheDelete)
	_ = w.Bind("apiBigFiles", apiBigFiles)
	_ = w.Bind("apiMediaScan", apiMediaScan)
	_ = w.Bind("apiOpenFolder", apiOpenFolder)
	_ = w.Bind("apiDism", apiDism)
	_ = w.Bind("apiCompact", apiCompact)
	_ = w.Bind("apiClean", func(deep, dry bool) (map[string]interface{}, error) {
		go doClean(w, deep, dry)
		return map[string]interface{}{"ok": true}, nil
	})

	w.SetHtml(guiHTML)
	w.Run()
}
