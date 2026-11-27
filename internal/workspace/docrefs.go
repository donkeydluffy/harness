package workspace

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// DocLine 是受检 Markdown 中代码块之外的一行（AC-030-1）。
type DocLine struct {
	File string // 仓库相对路径
	Line int
	Text string
}

// 用例行（测试输入）与字段行（由结构规则检查）不检查（AC-030-1/B3）。
var (
	caseLineRe  = regexp.MustCompile(`^\s*- AC-N?\d+-\d+/[NEB]\d+:`)
	fieldLineRe = regexp.MustCompile(`^- [A-Z][A-Za-z-]*:`)
)

// excludedSections 是 PRD 中不检查的章节（AC-030-1、AC-030-3/B1）。
var excludedSections = []string{"澄清记录", "变更日志"}

// CheckedDocs 返回受检 Markdown 的全部文件：README.md、AGENTS.md、CLAUDE.md、.claude/skills/**/SKILL.md 与 docs/**/*.md，
// 不含进度文件。
func (s *Snapshot) CheckedDocs() []string {
	var out []string
	for _, f := range []string{"README.md", "AGENTS.md", "CLAUDE.md"} {
		if fi, err := os.Stat(filepath.Join(s.Root, f)); err == nil && !fi.IsDir() {
			out = append(out, f)
		}
	}
	walk := func(dir string, keep func(rel string) bool) {
		filepath.WalkDir(filepath.Join(s.Root, filepath.FromSlash(dir)), func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(s.Root, p)
			if rel = filepath.ToSlash(rel); keep(rel) {
				out = append(out, rel)
			}
			return nil
		})
	}
	walk(".claude/skills", func(rel string) bool { return path.Base(rel) == "SKILL.md" })
	walk("docs", func(rel string) bool { return strings.HasSuffix(rel, ".md") && rel != "docs/progress.md" })
	sort.Strings(out)
	return out
}

// DocLines 返回受检 Markdown 中代码块之外的行，跳过 PRD 的澄清记录与变更日志章节、
// Release 大于 Current-Release 的需求、用例行与字段行（AC-030-1、AC-030-2/B1）。
func (s *Snapshot) DocLines() []DocLine {
	var out []DocLine
	future := s.futureLines()
	for _, f := range s.CheckedDocs() {
		b, err := os.ReadFile(filepath.Join(s.Root, filepath.FromSlash(f)))
		if err != nil {
			continue
		}
		fence, skip := "", false
		for i, l := range strings.Split(string(b), "\n") {
			t := strings.TrimSpace(l)
			if fence != "" {
				if strings.HasPrefix(t, fence) {
					fence = ""
				}
				continue
			}
			if strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~") {
				fence = t[:3]
				continue
			}
			if strings.HasPrefix(t, "## ") {
				title := strings.TrimSpace(strings.TrimPrefix(t, "## "))
				skip = false
				for _, x := range excludedSections {
					if strings.Contains(title, x) {
						skip = true
					}
				}
			}
			if !skip && !future[f][i+1] && !caseLineRe.MatchString(l) && !fieldLineRe.MatchString(l) {
				out = append(out, DocLine{File: f, Line: i + 1, Text: l})
			}
		}
	}
	return out
}

// InlineCode 返回行中的行内代码片段（反引号之间的文字）。
func (l DocLine) InlineCode() []string {
	var out []string
	parts := strings.Split(l.Text, "`")
	for i := 1; i+1 < len(parts); i += 2 {
		if c := strings.TrimSpace(parts[i]); c != "" {
			out = append(out, c)
		}
	}
	return out
}

// futureLines 返回 Release 大于 Current-Release 的需求所占的行（文件 -> 行号）。
func (s *Snapshot) futureLines() map[string]map[int]bool {
	out := map[string]map[int]bool{}
	idx := s.Doc(PRDIndex)
	if idx == nil {
		return out
	}
	cur, err := strconv.Atoi(idx.Header["Current-Release"])
	if err != nil {
		return out
	}
	for _, d := range s.PRDDocs() {
		for _, e := range d.Entities {
			rel, err := strconv.Atoi(e.Fields["Release"])
			if err != nil || rel <= cur {
				continue
			}
			if out[d.Path] == nil {
				out[d.Path] = map[int]bool{}
			}
			for n := 0; n < strings.Count(e.Raw, "\n")+1; n++ {
				out[d.Path][e.Line+n] = true
			}
		}
	}
	return out
}
