package builtin

import (
	"fmt"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// doc03：受检 Markdown 行内代码中的 `harness <子命令>` 必须是存在的子命令（AC-030-2）。
// 快照中没有子命令名（不经 cli 运行）时不检查。
type doc03 struct{}

func (doc03) ID() string       { return "doc.03" }
func (doc03) Covers() []string { return []string{"AC-030-2"} }
func (doc03) Fix() string {
	return "改为现有的子命令（harness help 列出全部子命令），或删掉过期的用法"
}

func (doc03) Check(s *workspace.Snapshot) []engine.Violation {
	if len(s.Commands) == 0 {
		return nil
	}
	var out []engine.Violation
	for _, l := range s.DocLines() {
		for _, c := range l.InlineCode() {
			f := strings.Fields(c)
			if len(f) < 2 || f[0] != "harness" || strings.HasPrefix(f[1], "-") || strings.ContainsAny(f[1], "<>[]") {
				continue
			}
			if !contains(s.Commands, f[1]) {
				out = append(out, engine.Violation{File: l.File, Line: l.Line,
					Msg: fmt.Sprintf("%s 中的子命令 %s 不存在（现有：%s）", c, f[1], strings.Join(s.Commands, "、"))})
			}
		}
	}
	return out
}
