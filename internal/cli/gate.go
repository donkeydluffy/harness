package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/donkeydluffy/harness/internal/adapter"
	"github.com/donkeydluffy/harness/internal/app"
	"github.com/donkeydluffy/harness/internal/gate"
)

// gateCmd 是代理 hook 入口：适配器解码输入，app.Gate 判定，适配器编码输出。
// 输入无法解析时判定为拒绝（AC-012-5/E1）。
func gateCmd(root string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintf(stderr, "用法：harness gate <代理> <事件>（代理：%s）\n", strings.Join(adapter.Names(), ", "))
		return ExitInternal
	}
	a, ok := adapter.Get(args[0])
	if !ok {
		fmt.Fprintf(stderr, "harness gate：未知代理 %q（可用：%s）\n", args[0], strings.Join(adapter.Names(), ", "))
		return ExitInternal
	}
	in, err := io.ReadAll(stdin)
	var d gate.Decision
	if err == nil {
		var req gate.Request
		if req, err = a.Decode(args[1], in); err == nil {
			if req.Root != "" {
				root = req.Root // 项目根由适配器查找；hook 进程的当前目录可能是仓库的子目录
			}
			d = app.Gate(root, req)
		}
	}
	if err != nil {
		d = gate.Decision{Rule: "gate.input", Reason: err.Error()}
	}
	out, code := a.Encode(args[1], d)
	stdout.Write(out)
	if !d.Allow {
		fmt.Fprintln(stderr, "harness gate: 拒绝：["+d.Rule+"] "+d.Reason)
	}
	return code
}
