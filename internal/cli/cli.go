// Package cli 是命令行入口：参数解析、输出格式、退出码（MOD-01）。
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"harness/internal/app"
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
	if opt.JSON {
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
