package builtin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"harness/internal/model"
	"harness/internal/rules/engine"
	"harness/internal/vcs"
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

// archEntities 返回架构文档（docs/architecture/）中指定类型的实体。
func archEntities(s *workspace.Snapshot, kind string) []*model.Entity {
	var out []*model.Entity
	for _, d := range s.Project.Docs {
		if !strings.HasPrefix(d.Path, "docs/architecture/") {
			continue
		}
		for _, e := range d.Entities {
			if e.Kind == kind {
				out = append(out, e)
			}
		}
	}
	return out
}

// globBase 返回 glob 中第一个通配符之前的目录前缀，如 "internal/app/**" → "internal/app/"。
func globBase(g string) string {
	if i := strings.IndexAny(g, "*?[<"); i >= 0 {
		return g[:i]
	}
	return g
}

// globsOverlap 判断两个路径 glob 是否可能匹配同一文件（按前缀近似判断）。
func globsOverlap(a, b string) bool {
	ba, bb := globBase(a), globBase(b)
	wa, wb := ba != a, bb != b
	switch {
	case !wa && !wb:
		return a == b
	case wa && wb:
		return strings.HasPrefix(ba, bb) || strings.HasPrefix(bb, ba)
	case wa:
		return strings.HasPrefix(b, ba)
	default:
		return strings.HasPrefix(a, bb)
	}
}

// flowStep 是调用流程的一步：调用方模块 -> 被调用接口。
type flowStep struct{ Caller, IF string }

// parseSteps 解析 Steps 字段；格式错误的步骤返回其序号（从 1 开始）。
func parseSteps(v string) (steps []flowStep, bad []int) {
	for i, st := range model.SplitList(v) {
		parts := strings.Split(st, "->")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			bad = append(bad, i+1)
			steps = append(steps, flowStep{})
			continue
		}
		steps = append(steps, flowStep{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])})
	}
	return steps, bad
}

// designKinds 是设计元素的类型（TERM-011，含扩展点）。
var designKinds = []string{"MOD", "IF", "FLOW", "EXT", "ADR"}

// designElements 返回架构文档与 ADR 中的全部设计元素。
func designElements(s *workspace.Snapshot) []*model.Entity {
	var out []*model.Entity
	for _, d := range s.Project.Docs {
		if !strings.HasPrefix(d.Path, "docs/architecture/") && !strings.HasPrefix(d.Path, "docs/adr/") {
			continue
		}
		for _, e := range d.Entities {
			if contains(designKinds, e.Kind) {
				out = append(out, e)
			}
		}
	}
	return out
}

// effective 判断设计元素是否有效：ADR 不是 Superseded，其他元素不是 Removed。
func effective(e *model.Entity) bool {
	st := e.Fields["Status"]
	return st != "Superseded" && st != "Removed"
}

// traceRefs 返回设计元素声明的追溯引用（Covers 与 Relates）。
func traceRefs(e *model.Entity) []string {
	return append(e.List("Covers"), e.List("Relates")...)
}

// changeSet 是一个待检查的提交或暂存区改动。
type changeSet struct {
	Name   string // 提交短哈希或"暂存区"
	Commit *vcs.Commit
	Files  []vcs.FileChange
	Before func(path string) []byte // 改动前内容，文件不存在时为 nil
	After  func(path string) []byte // 改动后内容，文件被删除时为 nil
}

// changeSets 返回要检查的改动：指定提交范围时为范围内每个提交，否则为暂存区。
// 指定了范围却不是 git 仓库时返回"输入缺失"违规；未指定范围且不是 git 仓库时无可检查内容。
func changeSets(s *workspace.Snapshot) ([]changeSet, *engine.Violation) {
	missing := func(msg string) *engine.Violation {
		return &engine.Violation{File: ".", Line: 1, Msg: "输入缺失：" + msg,
			Fix: "在 git 仓库中运行，并确认提交范围有效（如 origin/main..HEAD）"}
	}
	if s.Repo == nil {
		if s.CommitRange != "" {
			return nil, missing("指定了提交范围 " + s.CommitRange + "，但当前目录不是 git 仓库")
		}
		return nil, nil
	}
	show := func(rev string) func(string) []byte {
		return func(p string) []byte {
			b, err := s.Repo.FileAt(rev, p)
			if err != nil {
				return nil
			}
			return b
		}
	}
	if s.CommitRange == "" {
		files, err := s.Repo.Staged()
		if err != nil {
			return nil, missing(err.Error())
		}
		if len(files) == 0 {
			return nil, nil
		}
		return []changeSet{{Name: "暂存区", Files: files, Before: show("HEAD"), After: show("")}}, nil
	}
	commits, err := s.Repo.Commits(s.CommitRange)
	if err != nil {
		return nil, missing(err.Error())
	}
	var out []changeSet
	for i := range commits {
		c := commits[i]
		files, err := s.Repo.Changes(c)
		if err != nil {
			return nil, missing(err.Error())
		}
		out = append(out, changeSet{Name: c.Hash[:min(8, len(c.Hash))], Commit: &c, Files: files,
			Before: show(c.Hash + "^"), After: show(c.Hash)})
	}
	return out, nil
}

// parseAt 解析某一版本的 Markdown 文件，内容为空时返回 nil。
func parseAt(path string, src []byte) *model.Doc {
	if src == nil {
		return nil
	}
	d, _ := model.ParseMarkdown(path, src)
	return d
}
