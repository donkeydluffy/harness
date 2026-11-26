// Package model 把 Markdown 产物解析为实体并组装为项目模型（MOD-03）。
// 格式定义见 docs/architecture/formats.md。只做结构解析，不做规则判断。
package model

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Doc 是一个已解析的 Markdown 文件。
type Doc struct {
	Path     string
	Header   map[string]string // 文件头字段，如 Status、Current-Release
	Entities []*Entity
	Tables   []Table
	Headings []Heading
	Lines    []string // 原始行，规则按行号取文本时使用
}

// Heading 是文件中的任意标题。
type Heading struct {
	Level int
	Text  string
	Line  int
}

// Entity 是带编号的标题及其字段、列表项和代码块。
type Entity struct {
	ID     string // 按书写原样保存，格式是否合法由规则判断
	Kind   string // 编号前缀：FR、NFR、OQ、MOD、EXT、IF、FLOW、ADR、T、DEBT
	Title  string
	Level  int
	Fields map[string]string
	Order  []string // 字段书写顺序
	Items  []*Item
	Code   []CodeBlock
	File   string
	Line   int
	Raw    string // 标题到实体范围结束的全部文本
}

// Item 是实体内形如 "- AC-001-1: 文字" 的列表项，Children 为其下缩进两格的用例。
type Item struct {
	ID       string
	Text     string
	Line     int
	Children []*Item
}

// CodeBlock 是围栏代码块。
type CodeBlock struct {
	Lang string
	Text string
	Line int // 开始围栏所在行
}

// Table 是 Markdown 表格；Section 为表格所在的最近一级标题。
type Table struct {
	Section string
	Header  []string
	Rows    []Row
	Line    int
}

// Row 是表格的一行数据。
type Row struct {
	Cells []string
	Line  int
}

// Diagnostic 是解析或组装阶段发现的结构问题。
type Diagnostic struct {
	File string
	Line int
	Msg  string
}

// Field 返回字段值及其是否存在。
func (e *Entity) Field(key string) (string, bool) {
	v, ok := e.Fields[key]
	return v, ok
}

// List 把逗号分隔的字段值拆为列表；空值返回空列表。
func (e *Entity) List(key string) []string {
	return SplitList(e.Fields[key])
}

// SplitList 按英文逗号拆分并去掉空白与空项。
func SplitList(v string) []string {
	var out []string
	for _, s := range strings.Split(v, ",") {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

var (
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)
	// 宽松匹配编号，格式错误（如 FR-3、FR003）也识别为实体，交给规则报告。
	entityRe = regexp.MustCompile(`^(FR|NFR|OQ|MOD|EXT|IF|FLOW|ADR|T|DEBT)(-?\d+)(?:\s+(.*))?$`)
	fieldRe  = regexp.MustCompile(`^- ([A-Za-z][A-Za-z-]*):(?:\s(.*))?$`)
	itemRe   = regexp.MustCompile(`^- (AC-[^:\s]+):\s*(.*)$`)
	childRe  = regexp.MustCompile(`^  - (AC-[^:\s]+):\s*(.*)$`)
	fenceRe  = regexp.MustCompile("^(```|~~~)\\s*(\\S*)")
)

// ParseMarkdown 把单个 Markdown 文件解析为实体列表，同时保留表格行和代码块（IF-001）。
//
// harness:IF-001
func ParseMarkdown(path string, src []byte) (*Doc, []Diagnostic) {
	text := strings.ReplaceAll(string(src), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	d := &Doc{Path: path, Header: map[string]string{}, Lines: lines}
	var diags []Diagnostic

	var cur *Entity
	var curItem *Item
	var table *Table
	inFields := false
	headerState := 0 // 0 未见标题；1 见到一级标题，等待文件头字段；2 文件头字段中；3 结束
	section := ""
	inCode := false
	var code *CodeBlock
	var codeLines []string
	fence := ""

	closeEntity := func(end int) {
		if cur == nil {
			return
		}
		cur.Raw = strings.Join(lines[cur.Line-1:end], "\n")
		cur = nil
		curItem = nil
	}

	for i, line := range lines {
		ln := i + 1
		if inCode {
			if strings.HasPrefix(strings.TrimSpace(line), fence) {
				inCode = false
				code.Text = strings.Join(codeLines, "\n")
				if cur != nil {
					cur.Code = append(cur.Code, *code)
				}
				code = nil
			} else {
				codeLines = append(codeLines, line)
			}
			continue
		}
		if m := fenceRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			inCode, fence = true, m[1]
			code = &CodeBlock{Lang: m[2], Line: ln}
			codeLines = nil
			inFields, table = false, nil
			continue
		}
		if m := headingRe.FindStringSubmatch(line); m != nil {
			level := len(m[1])
			d.Headings = append(d.Headings, Heading{Level: level, Text: m[2], Line: ln})
			table = nil
			if cur != nil && level <= cur.Level {
				closeEntity(i)
			}
			if level == 1 && headerState == 0 {
				headerState = 1
			} else if headerState < 3 {
				headerState = 3
			}
			// 实体标题为二到四级；一级标题只有 ADR 文件（一文件一条）是实体。
			em := entityRe.FindStringSubmatch(m[2])
			if em != nil && (level > 4 || (level == 1 && em[1] != "ADR")) {
				em = nil
			}
			if em != nil && cur == nil {
				cur = &Entity{
					ID: em[1] + em[2], Kind: em[1], Title: em[3], Level: level,
					Fields: map[string]string{}, File: path, Line: ln,
				}
				d.Entities = append(d.Entities, cur)
				inFields = true
				// 一级标题实体（ADR 文件）的字段同时是文件头。
				if level != 1 {
					headerState = 3
				}
			} else {
				inFields = false
				section = m[2]
			}
			continue
		}
		if fm := fieldRe.FindStringSubmatch(line); fm != nil && (inFields || headerState == 1 || headerState == 2) && !itemRe.MatchString(line) {
			if headerState == 1 || headerState == 2 {
				d.Header[fm[1]] = strings.TrimSpace(fm[2])
				headerState = 2
			}
			if inFields && cur != nil {
				cur.Fields[fm[1]] = strings.TrimSpace(fm[2])
				cur.Order = append(cur.Order, fm[1])
			}
			continue
		}
		if strings.TrimSpace(line) == "" {
			// 一级标题与文件头字段之间允许一个空行。
			if headerState == 2 {
				headerState = 3
			}
			if inFields && cur != nil && (len(cur.Fields) > 0 || cur.Level != 1) {
				inFields = false
			}
			table = nil
			continue
		}
		if headerState == 1 || headerState == 2 {
			headerState = 3
		}
		inFields = false
		if strings.HasPrefix(strings.TrimSpace(line), "|") {
			cells := splitRow(line)
			if table == nil {
				d.Tables = append(d.Tables, Table{Section: section, Header: cells, Line: ln})
				table = &d.Tables[len(d.Tables)-1]
			} else if !isSeparator(cells) {
				table.Rows = append(table.Rows, Row{Cells: cells, Line: ln})
			}
			continue
		}
		table = nil
		if cur != nil {
			if m := itemRe.FindStringSubmatch(line); m != nil {
				curItem = &Item{ID: m[1], Text: m[2], Line: ln}
				cur.Items = append(cur.Items, curItem)
				continue
			}
			if m := childRe.FindStringSubmatch(line); m != nil {
				if curItem == nil {
					diags = append(diags, Diagnostic{path, ln, fmt.Sprintf("用例 %s 不在任何验收标准之下", m[1])})
					continue
				}
				curItem.Children = append(curItem.Children, &Item{ID: m[1], Text: m[2], Line: ln})
				continue
			}
		}
	}
	if inCode {
		diags = append(diags, Diagnostic{path, code.Line, "代码块没有结束围栏"})
	}
	closeEntity(len(lines))
	// 一级标题实体（ADR）的文件头取实体字段。
	if len(d.Header) == 0 && len(d.Entities) > 0 && d.Entities[0].Level == 1 {
		for k, v := range d.Entities[0].Fields {
			d.Header[k] = v
		}
	}
	return d, diags
}

func splitRow(line string) []string {
	s := strings.TrimSpace(line)
	s = strings.TrimPrefix(s, "|")
	s = strings.TrimSuffix(s, "|")
	parts := strings.Split(s, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func isSeparator(cells []string) bool {
	for _, c := range cells {
		if strings.Trim(c, "-: ") != "" {
			return false
		}
	}
	return true
}

// Project 是全部产物组装后的模型。
type Project struct {
	Docs    []*Doc
	ByID    map[string]*Entity
	ByKind  map[string][]*Entity
	Cases   map[string]*Case
	Terms   []Term
	Changes []Change
}

// Case 是验收用例。
type Case struct {
	ID       string
	Category byte // N、B、E
	AC       string
	Req      string
	Text     string
	File     string
	Line     int
}

// Term 是统一语言表格的一行。
type Term struct {
	ID, Name, Code, Def string
	Banned              []string
	File                string
	Line                int
}

// Change 是变更日志表格的一行；同一 CHG 编号可占多行。
type Change struct {
	ID, Date, Action string
	Targets          []string
	Note             string
	File             string
	Line             int
}

var caseRe = regexp.MustCompile(`^(AC-N?\d+-\d+)/([NBE])\d+$`)

// Build 把多个 Doc 组装为项目模型，按编号建立索引；重复编号作为 Diagnostic 返回（IF-002）。
//
// harness:IF-002
func Build(docs []*Doc) (*Project, []Diagnostic) {
	p := &Project{Docs: docs, ByID: map[string]*Entity{}, ByKind: map[string][]*Entity{}, Cases: map[string]*Case{}}
	var diags []Diagnostic
	for _, d := range docs {
		for _, e := range d.Entities {
			if prev, ok := p.ByID[e.ID]; ok {
				diags = append(diags, Diagnostic{e.File, e.Line, fmt.Sprintf("编号 %s 重复，首次定义在 %s:%d", e.ID, prev.File, prev.Line)})
			} else {
				p.ByID[e.ID] = e
			}
			p.ByKind[e.Kind] = append(p.ByKind[e.Kind], e)
			for _, it := range e.Items {
				for _, c := range it.Children {
					m := caseRe.FindStringSubmatch(c.ID)
					cs := &Case{ID: c.ID, AC: it.ID, Req: e.ID, Text: c.Text, File: e.File, Line: c.Line}
					if m != nil {
						cs.Category = m[2][0]
					}
					if _, dup := p.Cases[c.ID]; dup {
						diags = append(diags, Diagnostic{e.File, c.Line, fmt.Sprintf("用例编号 %s 重复", c.ID)})
						continue
					}
					p.Cases[c.ID] = cs
				}
			}
		}
		for _, t := range d.Tables {
			if len(t.Header) == 0 {
				continue
			}
			for _, r := range t.Rows {
				if len(r.Cells) == 0 {
					continue
				}
				first := r.Cells[0]
				switch {
				case strings.HasPrefix(first, "TERM"):
					p.Terms = append(p.Terms, Term{
						ID: first, Name: cell(r, 1), Code: cell(r, 2), Def: cell(r, 3),
						Banned: splitBanned(cell(r, 4)), File: d.Path, Line: r.Line,
					})
				case strings.HasPrefix(first, "CHG"):
					p.Changes = append(p.Changes, Change{
						ID: first, Date: cell(r, 1), Action: cell(r, 2), Targets: SplitList(cell(r, 3)),
						Note: cell(r, 4), File: d.Path, Line: r.Line,
					})
				}
			}
		}
	}
	for k := range p.ByKind {
		sort.SliceStable(p.ByKind[k], func(i, j int) bool { return p.ByKind[k][i].ID < p.ByKind[k][j].ID })
	}
	return p, diags
}

func cell(r Row, i int) string {
	if i < len(r.Cells) {
		return r.Cells[i]
	}
	return ""
}

func splitBanned(v string) []string {
	var out []string
	for _, s := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == '、' || r == '，' }) {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}
