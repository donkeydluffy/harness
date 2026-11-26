package builtin

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/vcs"
	"github.com/donkeydluffy/harness/internal/workspace"
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
		cs := changeSet{Name: "暂存区", Files: files, Before: show("HEAD"), After: show("")}
		if s.Message != "" {
			cs.Commit = &vcs.Commit{Message: s.Message, Trailers: vcs.ParseTrailers(s.Message)}
		}
		return []changeSet{cs}, nil
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

var caseTokenRe = regexp.MustCompile(`AC_(N?\d+)_(\d+)_([NBE])(\d+)`)

// caseRefs 从测试名中提取引用的用例编号：AC_001_1_N1 → AC-001-1/N1（formats.md 第 5 节）。
func caseRefs(name string) []string {
	var out []string
	for _, m := range caseTokenRe.FindAllStringSubmatch(name, -1) {
		out = append(out, "AC-"+m[1]+"-"+m[2]+"/"+m[3]+m[4])
	}
	return out
}

// staticTest 是源码中静态声明的测试。
type staticTest struct {
	Name, File string
	Line       int
	Kind       lang.TestKind
}

// staticTests 返回测试文件中声明的全部测试。
func staticTests(s *workspace.Snapshot) []staticTest {
	var out []staticTest
	for _, f := range s.Sources {
		if f.Kind == lang.NotTest {
			continue
		}
		src, err := f.Content()
		if err != nil {
			continue
		}
		for _, n := range f.Lang.TestNames(src) {
			out = append(out, staticTest{Name: n, File: f.Path, Line: lineOf(src, n), Kind: f.Kind})
		}
	}
	return out
}

// lineOf 返回 name 首次出现的行号，找不到时为 1。
func lineOf(src []byte, name string) int {
	i := strings.Index(string(src), name)
	if i < 0 {
		return 1
	}
	return strings.Count(string(src[:i]), "\n") + 1
}

// reportKind 判断报告中的测试属于哪类测试文件：按顶层测试名匹配静态声明，
// 同名声明有多个时用报告中的包路径（Suite）与文件目录比对。找不到时为 NotTest。
func reportKind(tc testreport.TestCase, tests []staticTest) (staticTest, bool) {
	top, _, _ := strings.Cut(tc.Name, "/")
	var cands []staticTest
	for _, st := range tests {
		if st.Name == top && (tc.File == "" || tc.File == st.File) {
			cands = append(cands, st)
		}
	}
	if len(cands) == 1 {
		return cands[0], true
	}
	for _, c := range cands {
		dir := path.Dir(c.File)
		if tc.Suite == dir || strings.HasSuffix(tc.Suite, "/"+dir) {
			return c, true
		}
	}
	return staticTest{}, false
}

// sourceComment 是源码中的一行注释。
type sourceComment struct {
	File, Text string
	Line       int
}

// commentLines 按行返回全部源码文件的注释；块注释拆成多行，行号逐行递增。
func commentLines(s *workspace.Snapshot) []sourceComment {
	var out []sourceComment
	for _, f := range s.Sources {
		src, err := f.Content()
		if err != nil {
			continue
		}
		for _, c := range f.Lang.Comments(src) {
			for i, l := range strings.Split(c.Text, "\n") {
				out = append(out, sourceComment{File: f.Path, Text: l, Line: c.Line + i})
			}
		}
	}
	return out
}

var entityRefRe = regexp.MustCompile(`\b(DEBT|T)-(\d+)\b`)

// entityRefs 返回文本中引用的 DEBT 与 T 实体，编号按数值归一（DEBT-2 与 DEBT-002 相同）。
func entityRefs(s *workspace.Snapshot, text string) []*model.Entity {
	var out []*model.Entity
	for _, m := range entityRefRe.FindAllStringSubmatch(text, -1) {
		n, _ := strconv.Atoi(m[2])
		for _, e := range s.Project.ByKind[m[1]] {
			if num, err := strconv.Atoi(strings.TrimPrefix(e.ID, m[1]+"-")); err == nil && num == n {
				out = append(out, e)
			}
		}
	}
	return out
}

const tasksFile = "docs/tasks.md"

// tasks 返回任务列表中的任务。
func tasks(s *workspace.Snapshot) []*model.Entity {
	var out []*model.Entity
	if d := s.Doc(tasksFile); d != nil {
		for _, e := range d.Entities {
			if e.Kind == "T" {
				out = append(out, e)
			}
		}
	}
	return out
}

// acItem 返回编号为 id 的验收标准及其所属需求。
func acItem(s *workspace.Snapshot, id string) (*model.Item, *model.Entity) {
	for _, e := range requirements(s) {
		for _, it := range e.Items {
			if it.ID == id {
				return it, e
			}
		}
	}
	return nil, nil
}

// flowModules 返回调用流程涉及的模块：调用方与被调接口的所属模块。
func flowModules(s *workspace.Snapshot, flow string) map[string]bool {
	out := map[string]bool{}
	f := s.Project.ByID[flow]
	if f == nil || f.Kind != "FLOW" {
		return out
	}
	steps, _ := parseSteps(f.Fields["Steps"])
	for _, st := range steps {
		if st.Caller != "" {
			out[st.Caller] = true
		}
		if i := s.Project.ByID[st.IF]; i != nil && i.Fields["Module"] != "" {
			out[i.Fields["Module"]] = true
		}
	}
	return out
}

// taskFlows 返回任务引用的调用流程（Flow 字段与 Design 中的 FLOW）。
func taskFlows(t *model.Entity) []string {
	var out []string
	for _, d := range append(t.List("Flow"), t.List("Design")...) {
		if strings.HasPrefix(d, "FLOW-") && !contains(out, d) {
			out = append(out, d)
		}
	}
	return out
}

// taskModules 返回任务的关联模块：Design 中的模块、接口所属模块、所引用流程涉及的模块。
func taskModules(s *workspace.Snapshot, t *model.Entity) map[string]bool {
	out := map[string]bool{}
	for _, d := range t.List("Design") {
		e := s.Project.ByID[d]
		switch {
		case e == nil:
		case e.Kind == "MOD":
			out[d] = true
		case e.Kind == "IF" && e.Fields["Module"] != "":
			out[e.Fields["Module"]] = true
		}
	}
	for _, f := range taskFlows(t) {
		for m := range flowModules(s, f) {
			out[m] = true
		}
	}
	return out
}

// globWithin 判断范围 glob p 匹配的文件是否都在路径 glob m 内（按前缀近似判断）。
func globWithin(p, m string) bool {
	if p == m {
		return true
	}
	if strings.HasSuffix(m, "/**") {
		dir := strings.TrimSuffix(m, "**")
		return strings.HasPrefix(p, dir) || p == strings.TrimSuffix(dir, "/")
	}
	return globBase(p) == p && globBase(m) != m && workspace.Match(m, p)
}

// scopeOwners 返回路径包含范围 glob p 的模块，按编号排序。
func scopeOwners(s *workspace.Snapshot, p string) []string {
	var out []string
	for _, m := range s.Project.ByKind["MOD"] {
		if workspace.Removed(m) {
			continue
		}
		for _, g := range m.List("Paths") {
			if globWithin(p, g) {
				out = append(out, m.ID)
				break
			}
		}
	}
	sort.Strings(out)
	return out
}

// generatedFiles 是 harness 命令写入项目的生成文件（TERM-026），提交范围与行数检查豁免。
var generatedFiles = []string{"harness.json", "harness.lock", ".claude/settings.json", "CODEOWNERS", ".github/CODEOWNERS",
	".github/workflows/harness.yml", ".gitlab-ci.yml"}

// generatedDirs 下的文件同样是生成文件：harness init 写入的 skill。
var generatedDirs = []string{".claude/skills/"}

// exemptFile 判断文件是否为产物文件（docs/ 下）或生成文件。
func exemptFile(p string) bool {
	if strings.HasPrefix(p, "docs/") || contains(generatedFiles, p) {
		return true
	}
	for _, d := range generatedDirs {
		if strings.HasPrefix(p, d) {
			return true
		}
	}
	return false
}

// commitTask 返回提交的 Task 尾注对应的任务；尾注缺失、重复或任务不存在时返回 nil。
func commitTask(s *workspace.Snapshot, cs changeSet) *model.Entity {
	if cs.Commit == nil || len(cs.Commit.Trailers["Task"]) != 1 {
		return nil
	}
	for _, e := range entityRefs(s, cs.Commit.Trailers["Task"][0]) { // T-3 与 T-003 相同
		if e.Kind == "T" {
			return e
		}
	}
	return nil
}

func commitViolation(cs changeSet, file, format string, args ...any) engine.Violation {
	if file == "" {
		file = "."
	}
	return engine.Violation{File: file, Line: 1, Msg: cs.Name + "：" + fmt.Sprintf(format, args...)}
}

// caseResult 汇总测试报告中每个用例的集成测试结果。
type caseResult struct {
	passed   map[string]bool
	failed   map[string][]string
	other    map[string]string // 未计入覆盖的原因
	inReport map[string]bool   // 文件 + "\x00" + 测试名
}

func caseResults(s *workspace.Snapshot, statics []staticTest) caseResult {
	r := caseResult{map[string]bool{}, map[string][]string{}, map[string]string{}, map[string]bool{}}
	for _, tc := range s.Tests {
		st, ok := reportKind(tc, statics)
		if ok {
			r.inReport[st.File+"\x00"+st.Name] = true
		}
		for _, c := range caseRefs(tc.Name) {
			switch {
			case !ok || st.Kind != lang.Integration:
				r.other[c] = "只被单元测试或无法识别类型的测试引用"
			case tc.Status == testreport.Passed:
				r.passed[c] = true
			case tc.Status == testreport.Failed:
				r.failed[c] = append(r.failed[c], tc.Name)
			default:
				r.other[c] = "引用它的测试状态为 skipped"
			}
		}
	}
	return r
}

// violation 返回用例未被通过的集成测试覆盖时的违规；覆盖时返回 nil。
func (r caseResult) violation(s *workspace.Snapshot, id string) *engine.Violation {
	c := s.Project.Cases[id]
	switch {
	case len(r.failed[id]) > 0:
		return &engine.Violation{File: c.File, Line: c.Line,
			Msg: fmt.Sprintf("用例 %s 的集成测试失败：%s", id, strings.Join(r.failed[id], ", "))}
	case !r.passed[id]:
		why := "没有引用它的集成测试"
		if r.other[id] != "" {
			why = r.other[id]
		}
		return &engine.Violation{File: c.File, Line: c.Line,
			Msg: fmt.Sprintf("用例 %s 未被通过的集成测试覆盖（%s）", id, why)}
	}
	return nil
}

// effectivelyApproved 判断产物处于有效的批准状态：文件头为 Approved（ADR 为 Accepted），且批准记录与当前内容一致。
func effectivelyApproved(s *workspace.Snapshot, rel string) bool {
	d := s.Doc(rel)
	if d == nil || workspace.HeaderStatus(d) != workspace.ApprovedStatus(rel) {
		return false
	}
	a := s.Lock.LastApproval(rel)
	return a != nil && a.Hash == s.ApprovalHash(rel)
}

// coveredRequirement 返回任务关联条目所属的需求编号（AC 取其所属需求）。
func coveredRequirement(s *workspace.Snapshot, ref string) string {
	if strings.HasPrefix(ref, "AC-") {
		if _, req := acItem(s, ref); req != nil {
			return req.ID
		}
		return ""
	}
	return ref
}

// taskCases 返回任务关联条目涉及的有效用例：关联 FR 或 NFR 时为其下全部用例，关联 AC 时为该 AC 下的用例（ADR-008）。
func taskCases(s *workspace.Snapshot, t *model.Entity) []string {
	var out []string
	add := func(it *model.Item) {
		for _, c := range it.Children {
			if !workspace.CaseRemoved(c.Text) && !contains(out, c.ID) {
				out = append(out, c.ID)
			}
		}
	}
	for _, ref := range t.List("Covers") {
		if strings.HasPrefix(ref, "AC-") {
			if it, _ := acItem(s, ref); it != nil {
				add(it)
			}
			continue
		}
		if e := s.Project.ByID[ref]; e != nil && (e.Kind == "FR" || e.Kind == "NFR") {
			for _, it := range e.Items {
				add(it)
			}
		}
	}
	return out
}
