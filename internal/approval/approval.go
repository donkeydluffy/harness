// Package approval 按批准单位计算内容哈希、批准链与批准状态，生成批准记录（MOD-15，ADR-009）。
// 只读快照，不写文件。
package approval

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// UnitKind 是批准单位的类型（TERM-027）。
type UnitKind string

const (
	Requirement UnitKind = "requirement"
	Term        UnitKind = "term"
	Question    UnitKind = "question"
	Design      UnitKind = "design"
	ADR         UnitKind = "adr"
	Task        UnitKind = "task"
	Body        UnitKind = "body"
)

// BodySuffix 是正文单位编号的后缀：<文件路径>#正文。
const BodySuffix = "#正文"

// Unit 是一个批准单位。
type Unit struct {
	ID       string
	Kind     UnitKind
	File     string
	Line     int
	End      int    // 单位最后一行；正文单位为 0，由 Owns 判定
	Hash     string // 规范化内容哈希，不含 Status 字段；PRD 正文不含变更日志
	Text     string // 规范化内容
	Upstream []string
	lines    map[int]bool // 正文单位包含的行
}

// Owns 判断文件中的某一行是否属于本单位。
func (u Unit) Owns(file string, line int) bool {
	if file != u.File {
		return false
	}
	if u.Kind == Body {
		return u.lines[line] || line <= 1
	}
	return line >= u.Line && line <= u.End
}

var kinds = map[string]UnitKind{
	"FR": Requirement, "NFR": Requirement, "OQ": Question,
	"MOD": Design, "EXT": Design, "IF": Design, "FLOW": Design, "ADR": ADR, "T": Task,
}

// unitFile 判断文件中的内容是否按单位批准：PRD、设计、ADR、任务列表。
func unitFile(p string) bool {
	return p == workspace.PRDIndex || strings.HasPrefix(p, "docs/prd/") ||
		strings.HasPrefix(p, "docs/architecture/") || strings.HasPrefix(p, "docs/adr/") || p == "docs/tasks.md"
}

func isPRD(p string) bool { return p == workspace.PRDIndex || strings.HasPrefix(p, "docs/prd/") }

var statusRe = regexp.MustCompile(`^- Status:`)

// Units 返回快照中的全部批准单位，按文件、行号排序（IF-027）。
//
// harness:IF-027
func Units(s *workspace.Snapshot) []Unit {
	var out []Unit
	owner := map[string]string{} // 验收标准 -> 所属需求
	for _, d := range s.Project.Docs {
		if !unitFile(d.Path) {
			continue
		}
		taken := map[int]bool{}
		for _, e := range d.Entities {
			k, ok := kinds[e.Kind]
			if !ok {
				continue
			}
			raw := strings.Split(e.Raw, "\n")
			end := e.Line + len(raw) - 1
			for l := e.Line; l <= end; l++ {
				taken[l] = true
			}
			text := lock.Normalize(strings.Join(dropStatus(raw), "\n"))
			out = append(out, Unit{ID: e.ID, Kind: k, File: d.Path, Line: e.Line, End: end, Hash: lock.Hash(text), Text: text, Upstream: upstreamRefs(e, k)})
			for _, it := range e.Items {
				owner[it.ID] = e.ID
			}
		}
		for _, t := range s.Project.Terms {
			if t.File == d.Path && t.ID != "" {
				text := lock.Normalize(d.Lines[t.Line-1])
				taken[t.Line] = true
				out = append(out, Unit{ID: t.ID, Kind: Term, File: d.Path, Line: t.Line, End: t.Line, Hash: lock.Hash(text), Text: text})
			}
		}
		if b, ok := body(d, taken); ok {
			out = append(out, b)
		}
	}
	ids := map[string]bool{}
	for _, u := range out {
		ids[u.ID] = true
	}
	for i := range out {
		var ups []string
		seen := map[string]bool{}
		for _, r := range out[i].Upstream {
			if o, ok := owner[r]; ok {
				r = o
			}
			if ids[r] && r != out[i].ID && !seen[r] {
				seen[r] = true
				ups = append(ups, r)
			}
		}
		sort.Strings(ups)
		out[i].Upstream = ups
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Line < out[j].Line
	})
	return out
}

// dropStatus 去掉实体字段块中的 Status 行（标题之后、第一个空行之前）。
func dropStatus(raw []string) []string {
	out := []string{}
	for i, l := range raw {
		if i > 0 && l == "" {
			return append(out, raw[i:]...)
		}
		if i > 0 && statusRe.MatchString(l) {
			continue
		}
		out = append(out, l)
	}
	return out
}

func upstreamRefs(e *model.Entity, k UnitKind) []string {
	switch k {
	case Design, ADR:
		return append(e.List("Covers"), e.List("Relates")...)
	case Task:
		return append(append(e.List("Covers"), e.List("Design")...), e.List("Flow")...)
	}
	return nil
}

// body 返回文件中条目之外的文字组成的正文单位：不含文件头 Status 行，PRD 不含变更日志章节。
func body(d *model.Doc, taken map[int]bool) (Unit, bool) {
	skip := map[int]bool{}
	if i := workspace.HeaderStatusLine(d.Lines); i >= 0 {
		skip[i+1] = true
	}
	if isPRD(d.Path) {
		for i, h := range d.Headings {
			if !strings.Contains(h.Text, "变更日志") {
				continue
			}
			end := len(d.Lines)
			for _, n := range d.Headings[i+1:] {
				if n.Level <= h.Level {
					end = n.Line - 1
					break
				}
			}
			for l := h.Line; l <= end; l++ {
				skip[l] = true
			}
		}
	}
	lines := map[int]bool{}
	var kept []string
	for i, l := range d.Lines {
		ln := i + 1
		if taken[ln] || skip[ln] {
			continue
		}
		lines[ln] = true
		kept = append(kept, l)
	}
	text := lock.Normalize(strings.Join(kept, "\n"))
	if text == "" {
		return Unit{}, false
	}
	return Unit{ID: d.Path + BodySuffix, Kind: Body, File: d.Path, Line: 1, Hash: lock.Hash(text), Text: text, lines: lines}, true
}

// State 是批准状态。
type State string

const (
	Approved   State = "approved"
	Unapproved State = "unapproved"
	Stale      State = "stale"
)

// Status 是一个单位的批准状态；Cause 为引起失效的单位编号（自身内容变化时含自身）。
type Status struct {
	Unit  Unit
	State State
	Cause []string
}

// Evaluate 返回每个单位的批准状态（IF-028）：批准记录存在、自身与记录的上游哈希都与当前一致、
// 且每个上游为 approved 时为 approved；没有记录为 unapproved；其余为 stale。
//
// harness:IF-028
func Evaluate(s *workspace.Snapshot) map[string]Status {
	units := Units(s)
	byID := map[string]Unit{}
	for _, u := range units {
		byID[u.ID] = u
	}
	out := map[string]Status{}
	visiting := map[string]bool{}
	var eval func(id string) Status
	eval = func(id string) Status {
		if st, ok := out[id]; ok {
			return st
		}
		u := byID[id]
		rec, ok := s.Lock.Units[id]
		if !ok {
			out[id] = Status{Unit: u, State: Unapproved}
			return out[id]
		}
		visiting[id] = true
		st := Status{Unit: u, State: Approved}
		if rec.Hash != u.Hash {
			st.Cause = append(st.Cause, id)
		}
		for _, up := range u.Upstream {
			if visiting[up] {
				continue // 引用成环时不沿环判定
			}
			if rec.Upstream[up] != byID[up].Hash || eval(up).State != Approved {
				st.Cause = append(st.Cause, up)
			}
		}
		delete(visiting, id)
		if len(st.Cause) > 0 {
			st.State = Stale
		}
		out[id] = st
		return st
	}
	for _, u := range units {
		eval(u.ID)
	}
	return out
}

// Chain 返回单位及其全部上游（递归）的状态，按编号排序（IF-028）。
func Chain(s *workspace.Snapshot, id string) []Status {
	all := Evaluate(s)
	seen := map[string]bool{}
	var out []Status
	var walk func(id string)
	walk = func(id string) {
		st, ok := all[id]
		if !ok || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, st)
		for _, up := range st.Unit.Upstream {
			walk(up)
		}
	}
	walk(id)
	sort.Slice(out, func(i, j int) bool { return out[i].Unit.ID < out[j].Unit.ID })
	return out
}

// Pending 是待批准的单位；Before 为上次批准时的内容，新增单位为空。
type Pending struct {
	Unit   Unit
	Before string
}

// PendingIn 返回 target 范围内新增、修改、批准失效的单位（IF-029）。target 为产物路径、prd 或单位编号。
//
// harness:IF-029
func PendingIn(s *workspace.Snapshot, target string) ([]Pending, error) {
	units := Units(s)
	var in func(u Unit) bool
	t := path.Clean(strings.TrimPrefix(target, "./"))
	switch {
	case strings.EqualFold(target, "prd"):
		in = func(u Unit) bool { return isPRD(u.File) }
	case unitFile(t) && s.Doc(t) != nil:
		in = func(u Unit) bool { return u.File == t }
	default:
		found := false
		for _, u := range units {
			found = found || u.ID == target
		}
		if !found {
			return nil, fmt.Errorf("无法识别 %s：应为 prd、产物路径（docs/ 下的 PRD、设计、ADR、任务列表）或单位编号", target)
		}
		in = func(u Unit) bool { return u.ID == target }
	}
	states := Evaluate(s)
	var out []Pending
	for _, u := range units {
		if !in(u) || states[u.ID].State == Approved {
			continue
		}
		out = append(out, Pending{Unit: u, Before: s.Lock.Units[u.ID].Text})
	}
	return out, nil
}

// Record 为确认批准的单位生成批准记录，记录各上游单位当前的哈希（IF-029）。
func Record(s *workspace.Snapshot, units []Unit, by string, at time.Time) map[string]lock.UnitApproval {
	hash := map[string]string{}
	for _, u := range Units(s) {
		hash[u.ID] = u.Hash
	}
	out := map[string]lock.UnitApproval{}
	for _, u := range units {
		ups := map[string]string{}
		for _, up := range u.Upstream {
			ups[up] = hash[up]
		}
		out[u.ID] = lock.UnitApproval{Hash: u.Hash, Text: u.Text, By: by, At: at, Upstream: ups}
	}
	return out
}
