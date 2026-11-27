package builtin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// supp01：lint 与类型检查的抑制注释必须在同一行引用一个 Open 的技术债（AC-042-1）。
// 注释由语言配置包提取，字符串字面量中的内容不计入；Java 的抑制注解不是注释，按去掉字符串后的代码行识别。
type supp01 struct{}

func (supp01) Needs() engine.Input { return engine.InputSources }
func (supp01) ID() string          { return "supp.01" }
func (supp01) Covers() []string    { return []string{"AC-042-1"} }
func (supp01) Fix() string {
	return "在同一行引用 Open 的技术债，如 `//nolint:errcheck // DEBT-014`；没有对应项时先在 docs/debt.md 登记，或去掉抑制并修复问题"
}

var (
	suppCommentRe = regexp.MustCompile(`eslint-disable|@ts-ignore|@ts-expect-error|(?:^|[^A-Za-z])noqa\b|type:\s*ignore|pylint:\s*disable|(?:^|[^A-Za-z])nolint\b|NOLINT`)
	suppAnnotRe   = regexp.MustCompile(`@SuppressWarnings\b`)
	stringLitRe   = regexp.MustCompile("\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'|`[^`]*`")
)

func (supp01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	check := func(file string, line int, marker, text string) {
		var closed []string
		for _, e := range entityRefs(s, text) {
			if e.Kind != "DEBT" {
				continue
			}
			if e.Fields["Status"] == "Open" {
				return
			}
			closed = append(closed, fmt.Sprintf("%s（%s）", e.ID, e.Fields["Status"]))
		}
		msg := fmt.Sprintf("抑制注释 %s 没有在同一行引用 Open 的技术债", marker)
		if len(closed) > 0 {
			msg = fmt.Sprintf("抑制注释 %s 引用的 %s 已关闭，应引用 Open 的技术债", marker, strings.Join(closed, "、"))
		}
		out = append(out, engine.Violation{File: file, Line: line, Msg: msg})
	}
	for _, c := range commentLines(s) {
		if m := suppCommentRe.FindString(c.Text); m != "" {
			check(c.File, c.Line, strings.TrimLeft(m, " \t/#*"), c.Text)
		}
	}
	for _, f := range s.Sources {
		src, err := f.Content()
		if err != nil {
			continue
		}
		for i, l := range strings.Split(string(src), "\n") {
			code := stringLitRe.ReplaceAllString(l, `""`)
			if j := strings.Index(code, "//"); j >= 0 {
				code = code[:j]
			}
			if suppAnnotRe.MatchString(code) {
				check(f.Path, i+1, "@SuppressWarnings", l)
			}
		}
	}
	return out
}
