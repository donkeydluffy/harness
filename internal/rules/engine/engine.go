// Package engine 是规则引擎：规则接口、注册表、按选择器运行、样本自检（MOD-04）。
package engine

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/donkeydluffy/harness/internal/workspace"
)

// Rule 是一条确定性检查（IF-011）。
//
// harness:IF-011
type Rule interface {
	ID() string       // 如 "trace.01"
	Covers() []string // 对应的验收标准，如 "AC-007-1"
	Fix() string      // 默认修复建议
	Check(s *workspace.Snapshot) []Violation
}

// Input 是规则在产物文档之外需要的输入。
type Input uint8

const (
	InputSources Input = 1 << iota // 源码文件
	InputTests                     // 测试报告
	InputGit                       // 提交或暂存区
)

// Needer 是可选接口：规则声明需要的额外输入。未实现表示只读取产物文档。
// harness lock 只运行只读产物文档的规则。
type Needer interface{ Needs() Input }

// Needs 返回规则需要的额外输入。
func Needs(r Rule) Input {
	if n, ok := r.(Needer); ok {
		return n.Needs()
	}
	return 0
}

// Violation 是一条违规。
type Violation struct {
	Rule string `json:"rule"`
	File string `json:"file"`
	Line int    `json:"line"`
	Msg  string `json:"msg"`
	Fix  string `json:"fix"`
	Warn bool   `json:"warn,omitempty"` // 告警：输出但不计入违规、不影响退出码（AC-030-4、AC-047-1/E1）
}

var (
	mu       sync.RWMutex
	registry = map[string]Rule{}
)

// Register 注册规则；编号重复时 panic，使冲突在构建和测试时即暴露。
func Register(r Rule) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := registry[r.ID()]; dup {
		panic("engine: 规则编号重复 " + r.ID())
	}
	registry[r.ID()] = r
}

// Lookup 按编号查找规则。
func Lookup(id string) (Rule, bool) {
	mu.RLock()
	defer mu.RUnlock()
	r, ok := registry[id]
	return r, ok
}

// IDs 返回已注册规则的编号，按字典序。
func IDs() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(registry))
	for id := range registry {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// Selector 选择要运行的规则；IDs 与 Prefixes 都为空表示全部。
// 前缀 "prd" 匹配 "prd.01"；复数写法 "tasks" 同样匹配 "task.01"。
type Selector struct {
	IDs      []string
	Prefixes []string
}

// Report 是一次运行的结果（IF-012）。
type Report struct {
	Violations []Violation `json:"violations"`
	Warnings   []Violation `json:"warnings,omitempty"` // Warn 为 true 的违规，与 Violations 同样排序
	Ran        []string    `json:"ran"`
	Errors     []error     `json:"-"` // 规则内部错误，与违规分开报告（退出码 2）
}

func (sel Selector) match(id string) bool {
	if len(sel.IDs) == 0 && len(sel.Prefixes) == 0 {
		return true
	}
	for _, x := range sel.IDs {
		if x == id {
			return true
		}
	}
	for _, p := range sel.Prefixes {
		if strings.HasPrefix(id, p+".") || (strings.HasSuffix(p, "s") && strings.HasPrefix(id, strings.TrimSuffix(p, "s")+".")) {
			return true
		}
	}
	return false
}

// Unknown 返回选择器中没有匹配任何已注册规则的条目。
func (sel Selector) Unknown() []string {
	ids := IDs()
	var out []string
	for _, x := range sel.IDs {
		if _, ok := Lookup(x); !ok {
			out = append(out, x)
		}
	}
	for _, p := range sel.Prefixes {
		one := Selector{Prefixes: []string{p}}
		found := false
		for _, id := range ids {
			if one.match(id) {
				found = true
				break
			}
		}
		if !found {
			out = append(out, p)
		}
	}
	return out
}

// Run 运行选中的规则，输出按文件、行号、规则编号排序（IF-012）。告警放入 Warnings，不计入 Violations。
// 违规缺少修复建议时用规则的 Fix() 补齐；规则 panic 记入 Errors。
//
// harness:IF-012
func Run(s *workspace.Snapshot, sel Selector) Report {
	var rep Report
	for _, id := range IDs() {
		if !sel.match(id) {
			continue
		}
		r, _ := Lookup(id)
		rep.Ran = append(rep.Ran, id)
		vs, err := runOne(r, s)
		if err != nil {
			rep.Errors = append(rep.Errors, err)
			continue
		}
		for _, v := range vs {
			if v.Warn {
				rep.Warnings = append(rep.Warnings, v)
			} else {
				rep.Violations = append(rep.Violations, v)
			}
		}
	}
	sortViolations(rep.Violations)
	sortViolations(rep.Warnings)
	if rep.Violations == nil {
		rep.Violations = []Violation{}
	}
	return rep
}

func sortViolations(vs []Violation) {
	sort.SliceStable(vs, func(i, j int) bool {
		a, b := vs[i], vs[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Rule != b.Rule {
			return a.Rule < b.Rule
		}
		return a.Msg < b.Msg
	})
}

func runOne(r Rule, s *workspace.Snapshot) (vs []Violation, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("规则 %s 内部错误：%v", r.ID(), p)
		}
	}()
	vs = r.Check(s)
	for i := range vs {
		vs[i].Rule = r.ID()
		if vs[i].Fix == "" {
			vs[i].Fix = r.Fix()
		}
		if vs[i].Fix == "" {
			return nil, fmt.Errorf("规则 %s 的违规缺少修复建议（AC-N004-1）", r.ID())
		}
	}
	return vs, nil
}

// Sample 是规则自检样本。
type Sample struct {
	Name          string
	Snapshot      *workspace.Snapshot
	WantViolation bool
}

// Verify 用违规样本和合规样本验证规则：违规样本必须报告，合规样本必须不报告（IF-013）。
// 至少需要一个违规样本和一个合规样本，规则的 Fix() 不得为空。
//
// harness:IF-013
func Verify(r Rule, samples []Sample) error {
	if strings.TrimSpace(r.Fix()) == "" {
		return fmt.Errorf("规则 %s 没有默认修复建议", r.ID())
	}
	var bad, ok bool
	for _, sm := range samples {
		vs, err := runOne(r, sm.Snapshot)
		if err != nil {
			return err
		}
		if sm.WantViolation {
			bad = true
			if len(vs) == 0 {
				return fmt.Errorf("规则 %s 在违规样本 %s 上没有报告违规", r.ID(), sm.Name)
			}
		} else {
			ok = true
			if len(vs) > 0 {
				return fmt.Errorf("规则 %s 在合规样本 %s 上误报：%s:%d %s", r.ID(), sm.Name, vs[0].File, vs[0].Line, vs[0].Msg)
			}
		}
	}
	if !bad || !ok {
		return fmt.Errorf("规则 %s 需要至少一个违规样本和一个合规样本", r.ID())
	}
	return nil
}

// CoverageRules 是覆盖类规则：要求当前 Release 的全部用例与扩展点都有通过的测试，在 Release 完成前必然失败，
// 由 CI 中的全量检查保证（用户 2026-11-26 确认）。任务完成定义与会话结束检查中的"全量检查"不含它们。
var CoverageRules = []string{"test.02", "arch.05"}

// FullCheck 返回"全量检查"的选择器：全部已注册规则，不含 CoverageRules 与 exclude。
func FullCheck(exclude ...string) Selector {
	skip := map[string]bool{}
	for _, id := range append(append([]string{}, CoverageRules...), exclude...) {
		skip[id] = true
	}
	var sel Selector
	for _, id := range IDs() {
		if !skip[id] {
			sel.IDs = append(sel.IDs, id)
		}
	}
	return sel
}
