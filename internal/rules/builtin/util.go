package builtin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"harness/internal/model"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// 产物编号的合法格式（docs/architecture/formats.md 第 2 节）。
var idFormat = map[string]*regexp.Regexp{
	"FR":   regexp.MustCompile(`^FR-\d{3,}$`),
	"NFR":  regexp.MustCompile(`^NFR-\d{3,}$`),
	"OQ":   regexp.MustCompile(`^OQ-\d{3,}$`),
	"TERM": regexp.MustCompile(`^TERM-\d{3,}$`),
	"CHG":  regexp.MustCompile(`^CHG-\d{3,}$`),
}

// requirements 返回全部 PRD 文件中的 FR 与 NFR，按文件和行号顺序。
func requirements(s *workspace.Snapshot) []*model.Entity {
	var out []*model.Entity
	for _, d := range s.PRDDocs() {
		for _, e := range d.Entities {
			if e.Kind == "FR" || e.Kind == "NFR" {
				out = append(out, e)
			}
		}
	}
	return out
}

func at(e *model.Entity, format string, args ...any) engine.Violation {
	return engine.Violation{File: e.File, Line: e.Line, Msg: fmt.Sprintf(format, args...)}
}

func positiveInt(v string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	return err == nil && n > 0
}

func oneOf(v string, allowed ...string) bool {
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}

// hasAC 判断列表项中是否至少有一条验收标准。
func hasAC(items []*model.Item) bool {
	for _, it := range items {
		if strings.HasPrefix(it.ID, "AC-") {
			return true
		}
	}
	return false
}

var chapterNumRe = regexp.MustCompile(`^\d+(\.\d+)*\.?\s*`)

// chapterName 去掉章节标题的序号和括注，如 "2. 统一语言（Glossary）" → "统一语言"。
func chapterName(heading string) string {
	h := chapterNumRe.ReplaceAllString(strings.TrimSpace(heading), "")
	if i := strings.IndexAny(h, "（("); i > 0 {
		h = h[:i]
	}
	return strings.TrimSpace(h)
}

// chapter 返回文档中名为 name 的二级章节的行号范围 [start, end)，start 为标题行。
func chapter(d *model.Doc, name string) (start, end int, ok bool) {
	for i, h := range d.Headings {
		if h.Level != 2 || chapterName(h.Text) != name {
			continue
		}
		end = len(d.Lines) + 1
		for _, n := range d.Headings[i+1:] {
			if n.Level <= 2 {
				end = n.Line
				break
			}
		}
		return h.Line, end, true
	}
	return 0, 0, false
}

// inScope 判断需求的 Release 是否不大于 PRD 的 Current-Release（AC-003-5）。
func inScope(s *workspace.Snapshot, e *model.Entity) bool {
	idx := s.Doc(workspace.PRDIndex)
	if idx == nil {
		return true
	}
	cur, err1 := strconv.Atoi(idx.Header["Current-Release"])
	rel, err2 := strconv.Atoi(e.Fields["Release"])
	return err1 != nil || err2 != nil || rel <= cur
}

var inlineCodeRe = regexp.MustCompile("`[^`]*`")

// bodyLines 返回实体正文（不含标题、字段行、代码块）的行及行号，行内代码替换为空白。
func bodyLines(e *model.Entity) (lines []string, nums []int) {
	all, ln := proseLines(strings.Split(e.Raw, "\n"), e.Line)
	for i, l := range all {
		idx := ln[i] - e.Line
		if idx == 0 || (idx <= len(e.Order) && strings.HasPrefix(l, "- ")) {
			continue
		}
		lines = append(lines, l)
		nums = append(nums, ln[i])
	}
	return lines, nums
}

// proseLines 去掉围栏代码块，并把行内代码替换为等长空白；first 为第一行的行号。
func proseLines(src []string, first int) (lines []string, nums []int) {
	inCode := false
	for i, l := range src {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		lines = append(lines, inlineCodeRe.ReplaceAllStringFunc(l, func(m string) string { return strings.Repeat(" ", len(m)) }))
		nums = append(nums, first+i)
	}
	return lines, nums
}

// anyEntity 判断行号范围 [start, end) 内是否有指定类型的实体。
func anyEntity(d *model.Doc, start, end int, kind string) bool {
	for _, e := range d.Entities {
		if e.Kind == kind && e.Line > start && e.Line < end {
			return true
		}
	}
	return false
}

// anyRow 判断行号范围 [start, end) 内是否有首列以 prefix 开头的表格行。
func anyRow(d *model.Doc, start, end int, prefix string) bool {
	for _, t := range d.Tables {
		for _, r := range t.Rows {
			if r.Line > start && r.Line < end && len(r.Cells) > 0 && strings.HasPrefix(r.Cells[0], prefix) {
				return true
			}
		}
	}
	return false
}
