package gate

import (
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"github.com/donkeydluffy/harness/internal/workspace"
)

// decideCommand 解析 Bash 命令（ADR-003）：拒绝 harness approve / phase next，
// 对能静态判定的写入目标逐个按写入判定。无法静态判定的命令放行，由命令后比对（IF-018）兜底。
func decideCommand(s *workspace.Snapshot, cmd string) Decision {
	if id, desc := findSecret(cmd); id != "" {
		return denySecret("命令", id, desc)
	}
	if pat := dangerous(s, cmd); pat != "" {
		return deny("gate.command", "命令命中危险命令模式 %q（harness.json 的 dangerousCommands）", pat)
	}
	if regenRemovals(s, cmd) != nil {
		return allow() // 删除可重新生成的生成文件，随后由单独一次 harness init 写回
	}
	f, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil {
		return allow() // 语法无法解析时同样交给命令后比对
	}
	var d = allow()
	syntax.Walk(f, func(n syntax.Node) bool {
		if !d.Allow {
			return false
		}
		switch n := n.(type) {
		case *syntax.Redirect:
			if p, ok := literal(n.Word); ok && sensitive(s, Resolve(s.Root, s.Root, p)) {
				d = denySensitive(Resolve(s.Root, s.Root, p))
				return false
			}
			if n.Op == syntax.RdrOut || n.Op == syntax.AppOut || n.Op == syntax.RdrAll || n.Op == syntax.AppAll || n.Op == syntax.ClbOut {
				if p, ok := literal(n.Word); ok {
					d = decideWrite(s, Resolve(s.Root, s.Root, p), "")
				}
			}
		case *syntax.CallExpr:
			args := words(n.Args)
			if len(args) == 0 {
				return true // 只有变量赋值（如 d=$(mktemp -d)），赋值中的命令替换由 Walk 继续访问
			}
			if humanOnly(args) {
				d = deny("gate.protect", "harness approve 与 harness phase next 只能由人在终端执行")
				return false
			}
			if what, ok := recursiveRemoveOutside(s.Root, n); ok {
				d = deny("gate.command", "递归删除仓库外或无法静态判定的路径：%s", what)
				return false
			}
			for _, a := range args[1:] {
				if a != "" && !strings.HasPrefix(a, "-") && sensitive(s, Resolve(s.Root, s.Root, a)) {
					d = denySensitive(Resolve(s.Root, s.Root, a))
					return false
				}
			}
			for _, p := range writeTargets(args) {
				if d = decideWrite(s, Resolve(s.Root, s.Root, p), ""); !d.Allow {
					return false
				}
			}
		}
		return true
	})
	return d
}

// literal 返回只由字面量组成的词；含变量、命令替换、通配符时 ok 为 false。
func literal(w *syntax.Word) (string, bool) {
	if w == nil {
		return "", false
	}
	v := w.Lit()
	if v == "" {
		// 带引号的字面量
		var b strings.Builder
		for _, part := range w.Parts {
			switch p := part.(type) {
			case *syntax.SglQuoted:
				b.WriteString(p.Value)
			case *syntax.DblQuoted:
				for _, q := range p.Parts {
					lit, ok := q.(*syntax.Lit)
					if !ok {
						return "", false
					}
					b.WriteString(lit.Value)
				}
			case *syntax.Lit:
				b.WriteString(p.Value)
			default:
				return "", false
			}
		}
		v = b.String()
	}
	if v == "" || strings.ContainsAny(v, "*?[") {
		return "", false
	}
	return v, true
}

// words 返回命令的各个词；无法静态判定的词为空串。
func words(ws []*syntax.Word) []string {
	out := make([]string, len(ws))
	for i, w := range ws {
		out[i], _ = literal(w)
	}
	return out
}

// humanOnly 判断命令是否调用 harness approve 或 harness phase next（含 go run ./cmd/harness 形式）。
func humanOnly(args []string) bool {
	for i, a := range args {
		if filepath.Base(a) != "harness" || i+1 >= len(args) {
			continue
		}
		if args[i+1] == "approve" || (args[i+1] == "phase" && i+2 < len(args) && args[i+2] == "next") {
			return true
		}
	}
	return false
}

// onlyHarness 判断整条命令只有一个简单命令且调用 harness。
func onlyHarness(cmd string) bool {
	f, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil || len(f.Stmts) != 1 || f.Stmts[0].Redirs != nil {
		return false
	}
	call, ok := f.Stmts[0].Cmd.(*syntax.CallExpr)
	if !ok {
		return false
	}
	args := words(call.Args)
	for i, a := range args {
		if filepath.Base(a) == "harness" {
			return i == 0 || (i == 2 && args[0] == "go" && args[1] == "run")
		}
	}
	return false
}

// regenRemovals 返回整条命令只是 rm 可重新生成的生成文件时被删除的文件（仓库相对路径）；
// 命令含其他内容、其他路径或通配符时返回 nil（DEBT-009）。
func regenRemovals(s *workspace.Snapshot, cmd string) map[string]bool {
	f, err := syntax.NewParser().Parse(strings.NewReader(cmd), "")
	if err != nil || len(f.Stmts) != 1 || f.Stmts[0].Redirs != nil {
		return nil
	}
	call, ok := f.Stmts[0].Cmd.(*syntax.CallExpr)
	if !ok || len(call.Assigns) > 0 || len(call.Args) < 2 {
		return nil
	}
	if name, ok := literal(call.Args[0]); !ok || name != "rm" {
		return nil
	}
	out := map[string]bool{}
	for _, w := range call.Args[1:] {
		a, ok := literal(w)
		switch {
		case !ok:
			return nil
		case a == "-f":
			continue
		case strings.HasPrefix(a, "-"):
			return nil // -r 等其他选项不在允许范围内
		}
		p := Resolve(s.Root, s.Root, a)
		if !contains(workspace.RegeneratedFiles, p) {
			return nil
		}
		out[p] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// writeTargets 返回常见写文件命令的路径参数。
func writeTargets(args []string) []string {
	if len(args) == 0 {
		return nil
	}
	name := filepath.Base(args[0])
	operands := func(from int) []string {
		var out []string
		for _, a := range args[from:] {
			if a != "" && !strings.HasPrefix(a, "-") {
				out = append(out, a)
			}
		}
		return out
	}
	switch name {
	case "tee", "rm", "touch", "mkdir", "rmdir", "truncate":
		return operands(1)
	case "cp", "install", "ln":
		if ops := operands(1); len(ops) > 0 {
			return ops[len(ops)-1:]
		}
	case "mv":
		return operands(1) // 来源被删除，目标被写入
	case "sed", "gsed":
		inPlace, script := false, false
		var files []string
		for _, a := range args[1:] {
			switch {
			case a == "-i" || strings.HasPrefix(a, "-i") || a == "--in-place" || strings.HasPrefix(a, "--in-place="):
				inPlace = true
			case a == "-e" || a == "-f" || strings.HasPrefix(a, "--expression"):
				script = true
			case strings.HasPrefix(a, "-") || a == "":
			case !script:
				script = true // 第一个非选项参数是脚本
			default:
				files = append(files, a)
			}
		}
		if inPlace {
			return files
		}
	case "git":
		if len(args) > 2 && (args[1] == "checkout" || args[1] == "restore") {
			for i, a := range args {
				if a == "--" {
					return operands(i + 1)
				}
			}
			if args[1] == "restore" {
				return operands(2)
			}
		}
	}
	return nil
}

// Resolve 把路径解析为仓库相对路径：相对路径以 base 为基准，解析 ".." 与符号链接；
// 落在仓库外时返回以 "../" 开头的路径。
func Resolve(root, base, p string) string {
	if !filepath.IsAbs(p) {
		p = filepath.Join(base, p)
	}
	p = evalExisting(filepath.Clean(p))
	r := evalExisting(root)
	rel, err := filepath.Rel(r, p)
	if err != nil {
		return "../" + filepath.ToSlash(p)
	}
	return filepath.ToSlash(rel)
}

// evalExisting 解析路径中已存在部分的符号链接，不存在的后缀原样保留。
func evalExisting(p string) string {
	rest := ""
	for cur := p; ; {
		if r, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(r, rest)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return p
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}
