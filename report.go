package main

import (
	"fmt"
	"io"
)

// printReport 在终端/日志中打印每个分类的释放情况与总计。
func printReport(out io.Writer, results []*Result, includeDeep, dryRun bool) {
	deleted, errs, total := totalsOf(results)

	verb := "预计释放"
	if !dryRun {
		verb = "已释放"
	}

	mode := "普通"
	if includeDeep {
		mode = "深度"
	}
	if dryRun {
		mode = "演练-" + mode
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "==================== 清理结果 ====================")
	fmt.Fprintf(out, " 模式: %s | 项数: %d | 错误: %d\n", mode, deleted, errs)
	fmt.Fprintln(out)

	for _, r := range results {
		if r.Bytes == 0 && r.Deleted == 0 {
			continue
		}
		fmt.Fprintf(out, "  %-20s  %s  %s\n", r.Name, verb, formatBytes(r.Bytes))
		if r.Path != "" {
			fmt.Fprintf(out, "      └ %s\n", r.Path)
		}
	}

	fmt.Fprintln(out, "---------------------------------------------")
	if dryRun {
		fmt.Fprintf(out, "  [演练] 共可清理 %d 项，预计释放 %s\n", deleted, formatBytes(total))
	} else {
		fmt.Fprintf(out, "  共清理 %d 项，释放 %s\n", deleted, formatBytes(total))
	}
	if errs > 0 {
		fmt.Fprintf(out, "  注意: 有 %d 项因占用/权限无法清理（不影响其余结果）\n", errs)
	}
	fmt.Fprintln(out, "==================================================")
}
