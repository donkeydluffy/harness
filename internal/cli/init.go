package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"harness/internal/app"
	"harness/internal/scaffold"
)

// initCmd：harness init --lang <语言列表> [--agent <代理列表>]。
func initCmd(root string, args []string, stdout, stderr io.Writer) int {
	fl := flag.NewFlagSet("init", flag.ContinueOnError)
	fl.SetOutput(stderr)
	langs := fl.String("lang", "", "项目语言，逗号分隔（"+strings.Join(scaffold.Languages(), ", ")+"）")
	agents := fl.String("agent", "claudecode", "编程代理，逗号分隔")
	ci := fl.String("ci", "", "CI 配置模板：github（默认）或 gitlab")
	if err := fl.Parse(args); err != nil {
		return ExitInternal
	}
	res, err := app.Init(root, scaffold.InitOptions{Languages: split(*langs), Agents: split(*agents), CI: *ci})
	for _, p := range res.Written {
		fmt.Fprintln(stdout, "写入", p)
	}
	for _, p := range res.Skipped {
		fmt.Fprintln(stdout, "跳过", p, "（已存在且内容不同，未覆盖）")
	}
	if err != nil {
		var ue *scaffold.UnsupportedError
		if errors.As(err, &ue) {
			fmt.Fprintln(stderr, "harness init:", err)
			return ExitViolated
		}
		fmt.Fprintln(stderr, "harness init:", err)
		return ExitInternal
	}
	if _, ok := scaffold.GitHooksDir(root); !ok {
		fmt.Fprintln(stdout, "不是 git 仓库（或 .git 不是目录），跳过 git hook 安装；git init 后重新运行 harness init 安装")
	}
	fmt.Fprintf(stdout, "写入 %d 个文件，跳过 %d 个\n", len(res.Written), len(res.Skipped))
	return ExitOK
}

func split(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
