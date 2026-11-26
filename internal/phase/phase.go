// Package phase 定义流程阶段：内置阶段与项目阶段同一格式，合并后按顺序排列（MOD-10）。
// 新增内置阶段是扩展点 EXT-04：在 builtin/ 新增阶段文件并在 order.json 登记一行。
package phase

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Phase 是一个流程阶段（IF-015）。Entry、Exit 中每项是规则编号、"区域.*" 或 "*"（全部规则）。
type Phase struct {
	Name     string   // requirements / architecture / planning / implementation
	Order    int      // 从 1 开始
	Entry    []string // 进入规则
	Exit     []string // 退出规则
	Writable []string // 产物路径 glob
	Skill    string
	Builtin  bool
}

//go:embed builtin/*.json
var builtinFS embed.FS

// def 是阶段定义文件或 harness.json phases 中的一项。指针字段用来区分"缺少"与"为空"。
type def struct {
	Name     string    `json:"name"`
	Entry    *[]string `json:"entry"`
	Exit     *[]string `json:"exit"`
	Writable *[]string `json:"writable"`
	Skill    *string   `json:"skill"`
	// 项目新增阶段：排在哪个阶段之后。
	After string `json:"after"`
	// 修改已有阶段：追加或移除规则；移除时须引用 ADR。
	AddEntry    []string `json:"addEntry"`
	AddExit     []string `json:"addExit"`
	RemoveEntry []string `json:"removeEntry"`
	RemoveExit  []string `json:"removeExit"`
	ADR         string   `json:"adr"`
}

func (d def) modifies() bool {
	return d.Entry == nil && d.Exit == nil && d.Writable == nil && d.Skill == nil && d.After == ""
}

// Problem 是阶段定义中的一个问题，由 phase.01 报告。
type Problem struct{ File, Msg string }

// Load 合并内置阶段（internal/phase/builtin/*.json，顺序见 order.json）与 harness.json 的 phases（IF-015）。
// 定义有问题时返回第一个问题。
func Load(s *workspace.Snapshot) ([]Phase, error) {
	phases, probs := load(builtinFS, s)
	if len(probs) > 0 {
		return nil, fmt.Errorf("%s：%s", probs[0].File, probs[0].Msg)
	}
	return phases, nil
}

func load(fsys fs.FS, s *workspace.Snapshot) ([]Phase, []Problem) {
	var probs []Problem
	bad := func(file, format string, a ...any) { probs = append(probs, Problem{file, fmt.Sprintf(format, a...)}) }
	var phases []Phase
	var order []string
	if b, err := fs.ReadFile(fsys, "builtin/order.json"); err != nil || json.Unmarshal(b, &order) != nil {
		bad("internal/phase/builtin/order.json", "无法读取内置阶段顺序")
	}
	for _, name := range order {
		file := "internal/phase/builtin/" + name + ".json"
		var d def
		b, err := fs.ReadFile(fsys, "builtin/"+name+".json")
		if err == nil {
			err = json.Unmarshal(b, &d)
		}
		if err != nil {
			bad(file, "无法读取阶段定义：%v", err)
			continue
		}
		if p, ok := complete(d, file, bad); ok {
			p.Builtin = true
			phases = append(phases, p)
		}
	}
	for i, raw := range s.Config.Phases {
		file := workspace.ConfigFile
		var d def
		if err := json.Unmarshal(raw, &d); err != nil {
			bad(file, "phases 第 %d 项格式错误：%v", i+1, err)
			continue
		}
		idx := index(phases, d.Name)
		switch {
		case d.Name == "":
			bad(file, "phases 第 %d 项缺少 name", i+1)
		case idx >= 0 && d.modifies():
			modify(&phases[idx], d, s, file, bad)
		case idx >= 0:
			bad(file, "新增阶段 %s 与已有阶段同名；修改已有阶段只能用 addEntry、addExit、removeEntry、removeExit", d.Name)
		default:
			after := index(phases, d.After)
			if after < 0 {
				bad(file, "新增阶段 %s 的 after 引用的阶段 %q 不存在", d.Name, d.After)
				continue
			}
			if p, ok := complete(d, file, bad); ok {
				phases = append(phases[:after+1], append([]Phase{p}, phases[after+1:]...)...)
			}
		}
	}
	for i := range phases {
		phases[i].Order = i + 1
		for _, sel := range append(append([]string{}, phases[i].Entry...), phases[i].Exit...) {
			if !matchesRule(sel) {
				bad(source(phases[i]), "阶段 %s 引用的规则 %s 不存在", phases[i].Name, sel)
			}
		}
	}
	return phases, probs
}

func source(p Phase) string {
	if p.Builtin {
		return "internal/phase/builtin/" + p.Name + ".json"
	}
	return workspace.ConfigFile
}

// complete 校验完整的阶段定义：名称、进入规则、退出规则、可写路径、skill 均须出现（进入规则可为空列表）。
func complete(d def, file string, bad func(string, string, ...any)) (Phase, bool) {
	ok := true
	for _, f := range []struct {
		name    string
		present bool
	}{{"entry", d.Entry != nil}, {"exit", d.Exit != nil}, {"writable", d.Writable != nil}, {"skill", d.Skill != nil}} {
		if !f.present {
			bad(file, "阶段 %s 缺少 %s 字段", d.Name, f.name)
			ok = false
		}
	}
	if !ok {
		return Phase{}, false
	}
	return Phase{Name: d.Name, Entry: *d.Entry, Exit: *d.Exit, Writable: *d.Writable, Skill: *d.Skill}, true
}

func modify(p *Phase, d def, s *workspace.Snapshot, file string, bad func(string, string, ...any)) {
	if len(d.RemoveEntry)+len(d.RemoveExit) > 0 {
		if a := s.Project.ByID[d.ADR]; a == nil || a.Kind != "ADR" {
			bad(file, "阶段 %s 移除了规则 %s，必须用 adr 字段引用一条存在的 ADR", p.Name, strings.Join(append(append([]string{}, d.RemoveEntry...), d.RemoveExit...), ", "))
			return
		}
	}
	p.Entry = append(remove(p.Entry, d.RemoveEntry), d.AddEntry...)
	p.Exit = append(remove(p.Exit, d.RemoveExit), d.AddExit...)
}

func remove(list, drop []string) []string {
	var out []string
	for _, v := range list {
		keep := true
		for _, x := range drop {
			keep = keep && v != x
		}
		if keep {
			out = append(out, v)
		}
	}
	return out
}

func index(phases []Phase, name string) int {
	for i, p := range phases {
		if p.Name == name {
			return i
		}
	}
	return -1
}

// matchesRule 判断规则引用是否对应已注册的规则。
func matchesRule(sel string) bool {
	if sel == "*" {
		return true
	}
	for _, id := range engine.IDs() {
		if id == sel || (strings.HasSuffix(sel, ".*") && strings.HasPrefix(id, strings.TrimSuffix(sel, "*"))) {
			return true
		}
	}
	return false
}

// Selector 把规则引用列表转换为引擎选择器；列表为空时返回 ok=false（不运行任何规则）。
func Selector(refs []string) (engine.Selector, bool) {
	var sel engine.Selector
	for _, r := range refs {
		switch {
		case r == "*":
			return engine.Selector{}, true
		case strings.HasSuffix(r, ".*"):
			sel.Prefixes = append(sel.Prefixes, strings.TrimSuffix(r, ".*"))
		default:
			sel.IDs = append(sel.IDs, r)
		}
	}
	return sel, len(refs) > 0
}

// ErrLastPhase 表示已处于最后一个阶段。
var ErrLastPhase = errors.New("已是最后一个阶段，无法继续推进")

// Current 返回锁文件记录的当前阶段；未记录时为第一个阶段（IF-016）。
func Current(s *workspace.Snapshot, phases []Phase) (Phase, error) {
	if len(phases) == 0 {
		return Phase{}, errors.New("没有任何阶段定义")
	}
	if s.Lock.Phase == "" {
		return phases[0], nil
	}
	if i := index(phases, s.Lock.Phase); i >= 0 {
		return phases[i], nil
	}
	return Phase{}, fmt.Errorf("锁文件中的当前阶段 %q 没有定义", s.Lock.Phase)
}

// Next 检查当前阶段的退出规则与下一阶段的进入规则，返回下一阶段与检查结果；不写文件（IF-016）。
func Next(s *workspace.Snapshot, phases []Phase) (Phase, engine.Report, error) {
	cur, err := Current(s, phases)
	if err != nil {
		return Phase{}, engine.Report{}, err
	}
	if cur.Order >= len(phases) {
		return Phase{}, engine.Report{}, ErrLastPhase
	}
	next := phases[cur.Order]
	var rep engine.Report
	for _, refs := range [][]string{cur.Exit, next.Entry} {
		if sel, ok := Selector(refs); ok {
			r := engine.Run(s, sel)
			rep.Violations = append(rep.Violations, r.Violations...)
			rep.Ran = append(rep.Ran, r.Ran...)
			rep.Errors = append(rep.Errors, r.Errors...)
		}
	}
	return next, rep, nil
}
