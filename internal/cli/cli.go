// Package cli 是命令行入口：参数解析、输出格式、退出码（MOD-01）。
package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"harness/internal/app"
	"harness/internal/rules/engine"
)

// 退出码（AC-N004-2）。
const (
	ExitOK       = 0
	ExitViolated = 1
	ExitInternal = 2
)

const usage = `用法：harness <命令> [参数]

命令：
  check [规则或前缀...] [--json]   运行检查；无参数时运行全部规则
  lock                             全量检查通过后更新 harness.lock
  trace [--format md|json]         输出需求 × 设计元素覆盖矩阵
`

// Main 执行命令并返回退出码。
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return ExitInternal
	}
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, "harness:", err)
		return ExitInternal
	}
	switch args[0] {
	case "check":
		return check(root, args[1:], stdout, stderr)
	case "lock":
		return lockCmd(root, stdout, stderr)
	case "trace":
		return traceCmd(root, args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return ExitOK
	default:
		fmt.Fprintf(stderr, "harness: 未知命令 %q\n\n%s", args[0], usage)
		return ExitInternal
	}
}

func check(root string, args []string, stdout, stderr io.Writer) int {
	var opt app.CheckOptions
	for _, a := range args {
		switch {
		case a == "--json":
			opt.JSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(stderr, "harness check: 未知参数 %q\n", a)
			return ExitInternal
		default:
			opt.Rules = append(opt.Rules, a)
		}
	}
	rep, err := app.Check(root, opt)
	if err != nil {
		fmt.Fprintln(stderr, "harness check: 内部错误：", err)
		return ExitInternal
	}
	return printReport(rep, opt.JSON, stdout, stderr)
}

func traceCmd(root string, args []string, stdout, stderr io.Writer) int {
	format := "md"
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case a == "--format" && i+1 < len(args):
			format = args[i+1]
			i++
		case strings.HasPrefix(a, "--format="):
			format = strings.TrimPrefix(a, "--format=")
		default:
			fmt.Fprintf(stderr, "harness trace: 未知参数 %q\n", a)
			return ExitInternal
		}
	}
	out, err := app.Trace(root, format)
	if err != nil {
		fmt.Fprintln(stderr, "harness trace:", err)
		return ExitInternal
	}
	stdout.Write(out)
	return ExitOK
}

func lockCmd(root string, stdout, stderr io.Writer) int {
	err := app.Lock(root)
	var ve *app.ViolationsError
	if errors.As(err, &ve) {
		printReport(ve.Report, false, stdout, stderr)
		fmt.Fprintln(stderr, "harness lock: 检查未通过，锁文件未更新")
		return ExitViolated
	}
	if err != nil {
		fmt.Fprintln(stderr, "harness lock: 内部错误：", err)
		return ExitInternal
	}
	fmt.Fprintln(stdout, "已更新 harness.lock")
	return ExitOK
}

// printReport 按文本或 JSON 输出报告，返回退出码。
func printReport(rep engine.Report, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		if err := enc.Encode(rep); err != nil {
			fmt.Fprintln(stderr, "harness check:", err)
			return ExitInternal
		}
	} else {
		for _, v := range rep.Violations {
			fmt.Fprintf(stdout, "%s:%d: [%s] %s\n  修复：%s\n", v.File, v.Line, v.Rule, v.Msg, v.Fix)
		}
		fmt.Fprintf(stdout, "%d 条违规，运行 %d 条规则\n", len(rep.Violations), len(rep.Ran))
	}
	if len(rep.Violations) > 0 {
		return ExitViolated
	}
	return ExitOK
}
