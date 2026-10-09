package main

import (
	"fmt"
	"io"
)

// printReport 打印汇总（逐项进度已由 OnProgress 实时输出，这里不再重复列项）。
func printReport(out io.Writer, results []*Result, includeDeep, dryRun bool) {
	deleted, errs, total := totalsOf(results)

	mode := "普通"
	if includeDeep {
		mode = "深度"
	}
	if dryRun {
		mode = "演练-" + mode
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "-------------------- 汇总 --------------------")
	if dryRun {
		fmt.Fprintf(out, "  共 %d 项，预计可释放 %s\n", deleted, formatBytes(total))
	} else {
		fmt.Fprintf(out, "  共清理 %d 项，释放 %s\n", deleted, formatBytes(total))
	}
	if errs > 0 {
		fmt.Fprintf(out, "  注意: 有 %d 项因占用/权限无法清理（不影响其余结果）\n", errs)
	}
	fmt.Fprintln(out, "------------------------------------------------")
}
