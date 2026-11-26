package builtin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// arch01：模块、接口、调用流程、扩展点的必填字段与取值；模块路径不重叠；
// 模块的接口列表与接口的所属模块一致；同时报告架构文档的解析诊断。
type arch01 struct{}

func (arch01) ID() string { return "arch.01" }
func (arch01) Covers() []string {
	return []string{"AC-006-1", "AC-006-2", "AC-006-3", "AC-033-2"}
}
func (arch01) Fix() string {
	return "按 docs/architecture/formats.md 第 3 节补齐字段；模块路径互不重叠，接口在所属模块的 Interfaces 中列出"
}

var archIDFormat = map[string]*regexp.Regexp{
	"MOD":  regexp.MustCompile(`^MOD-\d{2,}$`),
	"EXT":  regexp.MustCompile(`^EXT-\d{2,}$`),
	"IF":   regexp.MustCompile(`^IF-\d{3,}$`),
	"FLOW": regexp.MustCompile(`^FLOW-\d{2,}$`),
}

var archRequired = map[string][]string{
	"MOD":  {"Status", "Paths", "Public", "Depends", "Interfaces", "Covers", "Core", "Coverage"},
	"IF":   {"Module", "Covers"},
	"FLOW": {"Covers", "Steps"},
	"EXT":  {"Status", "Interface", "Paths", "Closed", "Relates"},
}

// 允许为空列表的字段。
var archMayBeEmpty = map[string]bool{"Depends": true, "Interfaces": true, "Closed": true}

func (arch01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	mods := map[string]*model.Entity{}
	for _, e := range archEntities(s, "MOD") {
		mods[e.ID] = e
	}
	for kind, fields := range archRequired {
		for _, e := range archEntities(s, kind) {
			if !archIDFormat[kind].MatchString(e.ID) {
				out = append(out, at(e, "编号 %s 格式不合法", e.ID))
			}
			for _, f := range fields {
				v, ok := e.Field(f)
				if !ok || (strings.TrimSpace(v) == "" && !archMayBeEmpty[f]) {
					out = append(out, at(e, "%s 缺少字段 %s", e.ID, f))
				}
			}
		}
	}
	for _, e := range archEntities(s, "MOD") {
		if c, ok := e.Field("Core"); ok && c != "true" && c != "false" {
			out = append(out, at(e, "%s 的 Core 取值 %q 不合法，应为 true 或 false", e.ID, c))
		}
		if c, ok := e.Field("Coverage"); ok {
			if n, err := strconv.Atoi(c); err != nil || n < 0 || n > 100 {
				out = append(out, at(e, "%s 的 Coverage 取值 %q 不合法，应为 0-100 的整数", e.ID, c))
			}
		}
		for _, i := range e.List("Interfaces") {
			if f := s.Project.ByID[i]; f == nil || f.Kind != "IF" {
				out = append(out, at(e, "%s 的接口 %s 不存在", e.ID, i))
			} else if f.Fields["Module"] != e.ID {
				out = append(out, at(e, "%s 列出接口 %s，但该接口的所属模块是 %s", e.ID, i, f.Fields["Module"]))
			}
		}
	}
	out = append(out, pathOverlaps(archEntities(s, "MOD"))...)
	for _, e := range archEntities(s, "IF") {
		m := e.Fields["Module"]
		switch mod := mods[m]; {
		case m == "":
		case mod == nil:
			out = append(out, at(e, "%s 的所属模块 %s 不存在", e.ID, m))
		case !contains(mod.List("Interfaces"), e.ID):
			out = append(out, at(e, "%s 不在所属模块 %s 的 Interfaces 中", e.ID, m))
		}
		if !hasSignature(e) {
			out = append(out, at(e, "%s 缺少签名：需要一个首行为 // %s 的代码块", e.ID, e.ID))
		}
	}
	for _, e := range archEntities(s, "FLOW") {
		if _, bad := parseSteps(e.Fields["Steps"]); len(bad) > 0 {
			out = append(out, at(e, "%s 第 %v 步格式错误，应为 MOD-a -> IF-xxx", e.ID, bad))
		}
	}
	for _, e := range archEntities(s, "EXT") {
		if i := e.Fields["Interface"]; i != "" {
			if f := s.Project.ByID[i]; f == nil || f.Kind != "IF" {
				out = append(out, at(e, "%s 的承载接口 %s 不存在", e.ID, i))
			}
		}
		for _, m := range e.List("Closed") {
			if mods[m] == nil {
				out = append(out, at(e, "%s 的封闭模块 %s 不存在", e.ID, m))
			}
		}
	}
	out = append(out, ownDiags(s, "arch.01")...)
	return out
}

func pathOverlaps(mods []*model.Entity) []engine.Violation {
	var out []engine.Violation
	for i, a := range mods {
		for _, b := range mods[i+1:] {
			for _, pa := range a.List("Paths") {
				for _, pb := range b.List("Paths") {
					if globsOverlap(pa, pb) {
						out = append(out, at(b, "%s 的路径 %s 与 %s 的路径 %s 重叠", b.ID, pb, a.ID, pa))
					}
				}
			}
		}
	}
	return out
}

func hasSignature(e *model.Entity) bool {
	for _, c := range e.Code {
		first, _, _ := strings.Cut(strings.TrimSpace(c.Text), "\n")
		if strings.TrimSpace(first) == fmt.Sprintf("// %s", e.ID) {
			return true
		}
	}
	return false
}

func contains(list []string, x string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}
