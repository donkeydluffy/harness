package builtin

import (
	"path"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// adr01：每个 ADR 文件恰好一条 ADR，含标题、Status、Relates 与正文三节（背景、决策、后果）。
type adr01 struct{}

func (adr01) ID() string       { return "adr.01" }
func (adr01) Covers() []string { return []string{"AC-010-1"} }
func (adr01) Fix() string {
	return "ADR 文件以 \"# ADR-001 标题\" 开头，文件头写 - Status: Proposed|Accepted|Superseded 与 - Relates: <需求或设计元素>，正文含 ## 背景、## 决策、## 后果"
}

var adrIDRe = regexp.MustCompile(`^ADR-\d{3,}$`)

func (adr01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, d := range s.Project.Docs {
		if !strings.HasPrefix(d.Path, "docs/adr/") {
			continue
		}
		var adrs []*model.Entity
		for _, e := range d.Entities {
			if e.Kind == "ADR" && e.Level == 1 {
				adrs = append(adrs, e)
			}
		}
		if len(adrs) != 1 {
			out = append(out, engine.Violation{File: d.Path, Line: 1, Msg: "ADR 文件必须以一个 \"# ADR-nnn 标题\" 开头，且一个文件只写一条 ADR"})
			continue
		}
		e := adrs[0]
		if !adrIDRe.MatchString(e.ID) {
			out = append(out, at(e, "编号 %s 格式不合法，应为 ADR-001 形式", e.ID))
		}
		if !strings.HasPrefix(path.Base(d.Path), e.ID) {
			out = append(out, at(e, "文件名 %s 应以编号 %s 开头", path.Base(d.Path), e.ID))
		}
		if strings.TrimSpace(e.Title) == "" {
			out = append(out, at(e, "%s 缺少标题", e.ID))
		}
		if st := e.Fields["Status"]; !oneOf(st, "Proposed", "Accepted", "Superseded") {
			out = append(out, at(e, "%s 的 Status 取值 %q 不合法，合法取值：Proposed、Accepted、Superseded", e.ID, st))
		}
		if len(e.List("Relates")) == 0 {
			out = append(out, at(e, "%s 缺少 Relates（关联的需求或设计元素）", e.ID))
		}
		for _, sec := range []string{"背景", "决策", "后果"} {
			if !hasHeading(d, 2, sec) {
				out = append(out, at(e, "%s 缺少章节 ## %s", e.ID, sec))
			}
		}
	}
	out = append(out, ownDiags(s, "adr.01")...)
	return out
}

func hasHeading(d *model.Doc, level int, text string) bool {
	for _, h := range d.Headings {
		if h.Level == level && strings.TrimSpace(h.Text) == text {
			return true
		}
	}
	return false
}
