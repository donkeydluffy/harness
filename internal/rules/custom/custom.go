// Package custom 把 harness.json 的 rules 构造为规则（MOD-17，FR-054）：required、enum、ref、forbid 四类声明式规则，
// 校验编号、样本与批准前检查登记。本模块不运行外部程序。
package custom

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Def 是一条自定义规则的声明（IF-036）。
type Def struct {
	ID         string   `json:"id"`
	Kind       string   `json:"kind"`
	Entity     string   `json:"entity"`
	Paths      []string `json:"paths"`
	Field      string   `json:"field"`
	Values     []string `json:"values"`
	Pattern    string   `json:"pattern"`
	Message    string   `json:"message"`
	Samples    string   `json:"samples"`
	ApproveFor []string `json:"approveFor"`
}

// Kinds 是支持的规则类型。
var Kinds = []string{"required", "enum", "ref", "forbid"}

// UnitKinds 是可登记批准前检查的批准单位类型。
var UnitKinds = []string{"requirement", "design", "adr", "task", "body"}

var idRe = regexp.MustCompile(`^([a-z][a-z0-9-]*)\.(\d{2,})$`)

// Rule 是由声明构造的规则。
type Rule struct {
	Def Def
	re  *regexp.Regexp
}

func (r *Rule) ID() string       { return r.Def.ID }
func (r *Rule) Covers() []string { return nil }
func (r *Rule) Fix() string {
	if r.Def.Message != "" {
		return "按项目规则 " + r.Def.ID + " 修改：" + r.Def.Message
	}
	return "按项目规则 " + r.Def.ID + " 修改"
}

// Check 按规则类型检查快照。
func (r *Rule) Check(s *workspace.Snapshot) []engine.Violation {
	d := r.Def
	var out []engine.Violation
	add := func(file string, line int, format string, a ...any) {
		msg := fmt.Sprintf(format, a...)
		if d.Message != "" {
			msg += "：" + d.Message
		}
		out = append(out, engine.Violation{File: file, Line: line, Msg: msg})
	}
	if d.Kind == "forbid" {
		for _, f := range files(s.Root, d.Paths) {
			b, err := os.ReadFile(filepath.Join(s.Root, f))
			if err != nil {
				continue
			}
			sc := bufio.NewScanner(bytes.NewReader(b))
			sc.Buffer(make([]byte, 1024*1024), 1024*1024)
			for n := 1; sc.Scan(); n++ {
				if r.re.MatchString(sc.Text()) {
					add(f, n, "命中禁用模式 %s", d.Pattern)
				}
			}
		}
		return out
	}
	for _, e := range s.Project.ByKind[d.Entity] {
		v := strings.TrimSpace(e.Fields[d.Field])
		switch d.Kind {
		case "required":
			if v == "" {
				add(e.File, e.Line, "%s 缺少字段 %s", e.ID, d.Field)
			}
		case "enum":
			if v != "" && !contains(d.Values, v) {
				add(e.File, e.Line, "%s 的 %s 为 %s，不在允许的取值 %s 中", e.ID, d.Field, v, strings.Join(d.Values, ", "))
			}
		case "ref":
			for _, id := range e.List(d.Field) {
				if s.Project.ByID[id] == nil {
					add(e.File, e.Line, "%s 的 %s 引用的 %s 不存在", e.ID, d.Field, id)
				}
			}
		}
	}
	return out
}

// files 返回项目中匹配 globs 的文件（不含 .git 与 .harness），按路径排序。
func files(root string, globs []string) []string {
	var out []string
	filepath.WalkDir(root, func(p string, de fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if de.IsDir() {
			if rel == ".git" || rel == ".harness" {
				return filepath.SkipDir
			}
			return nil
		}
		if workspace.MatchAny(globs, rel) {
			out = append(out, rel)
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func problem(id, format string, a ...any) engine.Violation {
	rule := id
	if rule == "" {
		rule = "custom"
	}
	return engine.Violation{Rule: rule, File: workspace.ConfigFile, Line: 1, Msg: fmt.Sprintf(format, a...),
		Fix: "按 docs/architecture/formats.md 第 6 节修改 harness.json 的 rules，确认后运行 harness lock"}
}

// Load 解析项目配置的 rules 并构造规则（IF-036）；定义有误（类型不支持、编号无前缀或与内置规则冲突、
// 批准单位类型不存在）时返回违规，该条规则不构造（AC-054-1、AC-054-4、AC-054-5）。builtin 为内置规则编号。
//
// harness:IF-036
func Load(s *workspace.Snapshot, builtin []string) ([]engine.Rule, []engine.Violation) {
	areas := map[string]bool{}
	for _, id := range builtin {
		areas[strings.SplitN(id, ".", 2)[0]] = true
	}
	var rules []engine.Rule
	var bad []engine.Violation
	seen := map[string]bool{}
	for i, raw := range s.Config.Rules {
		var d Def
		if err := json.Unmarshal(raw, &d); err != nil {
			bad = append(bad, problem("", "rules 第 %d 项无法解析：%v", i+1, err))
			continue
		}
		m := idRe.FindStringSubmatch(d.ID)
		var errs []string
		switch {
		case m == nil:
			errs = append(errs, "编号须为 <项目前缀>.<两位以上序号>，如 acme.01")
		case areas[m[1]]:
			errs = append(errs, "前缀 "+m[1]+" 与内置规则冲突，换用项目自己的前缀")
		case seen[d.ID]:
			errs = append(errs, "编号重复")
		}
		seen[d.ID] = true
		r := &Rule{Def: d}
		switch d.Kind {
		case "required", "enum", "ref":
			if d.Entity == "" || d.Field == "" {
				errs = append(errs, d.Kind+" 规则须写 entity 与 field")
			}
			if d.Kind == "enum" && len(d.Values) == 0 {
				errs = append(errs, "enum 规则须写 values")
			}
		case "forbid":
			re, err := regexp.Compile(d.Pattern)
			switch {
			case len(d.Paths) == 0 || d.Pattern == "":
				errs = append(errs, "forbid 规则须写 paths 与 pattern")
			case err != nil:
				errs = append(errs, "pattern 不是有效的正则："+err.Error())
			}
			r.re = re
		default:
			errs = append(errs, fmt.Sprintf("类型 %q 不支持，可用类型：%s", d.Kind, strings.Join(Kinds, ", ")))
		}
		for _, k := range d.ApproveFor {
			if !contains(UnitKinds, k) {
				errs = append(errs, fmt.Sprintf("approveFor 中的批准单位类型 %q 不存在，可用类型：%s", k, strings.Join(UnitKinds, ", ")))
			}
		}
		if d.Samples == "" {
			errs = append(errs, "须写 samples（含 violation/ 与 ok/ 两个样本项目的目录）")
		}
		if len(errs) > 0 {
			name := d.ID
			if name == "" {
				name = fmt.Sprintf("第 %d 项", i+1)
			}
			bad = append(bad, problem(d.ID, "自定义规则 %s 定义有误：%s", name, strings.Join(errs, "；")))
			continue
		}
		rules = append(rules, r)
	}
	return rules, bad
}

// Verify 用每条规则的样本运行 IF-013；样本缺失、违规样本未报告或合规样本被报告时返回违规（AC-054-3）。
func Verify(root string, rules []engine.Rule) []engine.Violation {
	var bad []engine.Violation
	for _, r := range rules {
		cr, ok := r.(*Rule)
		if !ok {
			continue
		}
		dir := filepath.Join(root, filepath.FromSlash(cr.Def.Samples))
		var samples []engine.Sample
		var missing []string
		for _, sub := range []string{"violation", "ok"} {
			p := filepath.Join(dir, sub)
			if st, err := os.Stat(p); err != nil || !st.IsDir() {
				missing = append(missing, cr.Def.Samples+"/"+sub)
				continue
			}
			s, err := workspace.Load(p, workspace.LoadOptions{})
			if err != nil {
				bad = append(bad, problem(cr.Def.ID, "自定义规则 %s 的样本 %s/%s 无法加载：%v", cr.Def.ID, cr.Def.Samples, sub, err))
				continue
			}
			samples = append(samples, engine.Sample{Name: sub, Snapshot: s, WantViolation: sub == "violation"})
		}
		if len(missing) > 0 {
			bad = append(bad, problem(cr.Def.ID, "自定义规则 %s 缺少样本 %s，判定规则失效", cr.Def.ID, strings.Join(missing, "、")))
			continue
		}
		if err := engine.Verify(r, samples); err != nil {
			bad = append(bad, problem(cr.Def.ID, "%v，判定规则失效", err))
		}
	}
	return bad
}

// ApproveChecks 返回登记在某类批准单位上的规则编号。
func ApproveChecks(rules []engine.Rule, kind string) []string {
	var out []string
	for _, r := range rules {
		if cr, ok := r.(*Rule); ok && contains(cr.Def.ApproveFor, kind) {
			out = append(out, cr.Def.ID)
		}
	}
	return out
}

func contains(list []string, x string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}
