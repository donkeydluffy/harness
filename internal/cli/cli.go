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
	"harness/internal/phase"
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
  check [规则或前缀...] [--commits <范围>] [--msg-file <文件>] [--json]
                                   运行检查；无参数时运行全部规则。
                                   commit 规则默认检查暂存区；--commits 检查范围内每个提交，
                                   也可写作 check commits origin/main..HEAD；
                                   --msg-file 传入即将提交的信息（commit-msg hook 使用）
  lock                             产物文档检查通过后更新 harness.lock
  approve <产物>                   由人在终端批准产物（prd 或产物路径），写入批准记录
  phase list                       列出流程阶段，标出当前阶段
  phase next                       由人在终端推进到下一阶段（退出与进入规则须通过）
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
	case "approve":
		return approveCmd(root, args[1:], stdout, stderr)
	case "phase":
		return phaseCmd(root, args[1:], stdout, stderr)
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
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--json":
			opt.JSON = true
		case a == "--commits" || a == "--msg-file":
			if i+1 >= len(args) {
				fmt.Fprintf(stderr, "harness check: %s 需要参数\n", a)
				return ExitInternal
			}
			i++
			if a == "--commits" {
				opt.Commits = args[i]
			} else {
				opt.MessageFile = args[i]
			}
		case strings.Contains(a, "..") && len(opt.Rules) > 0 && strings.HasPrefix(opt.Rules[len(opt.Rules)-1], "commit"):
			opt.Commits = a // check commits <范围>
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

func approveCmd(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "用法：harness approve <产物>（prd 或产物路径，如 docs/architecture/overview.md）")
		return ExitInternal
	}
	err := app.Approve(root, args[0], newTerminal(stdout))
	var ve *app.ViolationsError
	switch {
	case err == nil:
		fmt.Fprintln(stdout, "已批准，批准记录写入 harness.lock")
		return ExitOK
	case errors.As(err, &ve):
		printReport(ve.Report, false, stdout, stderr)
		fmt.Fprintln(stderr, "harness approve: 检查未通过，拒绝批准")
		return ExitViolated
	case errors.Is(err, app.ErrNotInteractive), errors.Is(err, app.ErrDeclined):
		fmt.Fprintln(stderr, "harness approve:", err)
		return ExitViolated
	default:
		fmt.Fprintln(stderr, "harness approve: 内部错误：", err)
		return ExitInternal
	}
}

func phaseCmd(root string, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 || (args[0] != "list" && args[0] != "next") {
		fmt.Fprintln(stderr, "用法：harness phase list|next")
		return ExitInternal
	}
	if args[0] == "list" {
		phases, cur, err := app.PhaseList(root)
		if err != nil {
			fmt.Fprintln(stderr, "harness phase list: 内部错误：", err)
			return ExitInternal
		}
		for _, p := range phases {
			mark := " "
			if p.Name == cur {
				mark = "*"
			}
			fmt.Fprintf(stdout, "%s %d %s  退出规则：%s\n", mark, p.Order, p.Name, strings.Join(p.Exit, ", "))
		}
		return ExitOK
	}
	err := app.PhaseNext(root, newTerminal(stdout))
	var ve *app.ViolationsError
	switch {
	case err == nil:
		_, cur, _ := app.PhaseList(root)
		fmt.Fprintf(stdout, "已推进到 %s\n", cur)
		return ExitOK
	case errors.As(err, &ve):
		printReport(ve.Report, false, stdout, stderr)
		fmt.Fprintln(stderr, "harness phase next: 规则未通过，拒绝推进")
		return ExitViolated
	case errors.Is(err, phase.ErrLastPhase), errors.Is(err, app.ErrDeclined), errors.Is(err, app.ErrNotInteractive):
		fmt.Fprintln(stderr, "harness phase next:", err)
		return ExitViolated
	default:
		fmt.Fprintln(stderr, "harness phase next: 内部错误：", err)
		return ExitInternal
	}
}
