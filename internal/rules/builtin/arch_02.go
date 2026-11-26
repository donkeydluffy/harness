package builtin

import (
	"sort"
	"strings"

	"harness/internal/model"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// arch02：模块声明的依赖图无环；模块依赖自身也视为环。
type arch02 struct{}

func (arch02) ID() string       { return "arch.02" }
func (arch02) Covers() []string { return []string{"AC-006-4"} }
func (arch02) Fix() string {
	return "调整模块的 Depends，打断依赖环；常见做法是把共用部分下沉为更底层的模块，或改为依赖接口所在的模块"
}

func (arch02) Check(s *workspace.Snapshot) []engine.Violation {
	mods := map[string]*model.Entity{}
	var ids []string
	for _, e := range archEntities(s, "MOD") {
		mods[e.ID] = e
		ids = append(ids, e.ID)
	}
	sort.Strings(ids)
	const (
		white = iota
		grey
		black
	)
	color := map[string]int{}
	seen := map[string]bool{}
	var out []engine.Violation
	var stack []string
	var visit func(id string)
	visit = func(id string) {
		color[id] = grey
		stack = append(stack, id)
		deps := mods[id].List("Depends")
		sort.Strings(deps)
		for _, d := range deps {
			if mods[d] == nil {
				continue // 不存在的模块由 trace.03 报告
			}
			switch color[d] {
			case white:
				visit(d)
			case grey:
				cyc := cycleFrom(stack, d)
				key := canonical(cyc)
				if !seen[key] {
					seen[key] = true
					e := mods[cyc[0]]
					out = append(out, at(e, "模块依赖存在环：%s", strings.Join(append(cyc, cyc[0]), " -> ")))
				}
			}
		}
		stack = stack[:len(stack)-1]
		color[id] = black
	}
	for _, id := range ids {
		if color[id] == white {
			visit(id)
		}
	}
	return out
}

// cycleFrom 从栈中截取以 start 开始的环，并旋转为以最小编号开头。
func cycleFrom(stack []string, start string) []string {
	i := len(stack) - 1
	for i >= 0 && stack[i] != start {
		i--
	}
	cyc := append([]string(nil), stack[i:]...)
	min := 0
	for j := range cyc {
		if cyc[j] < cyc[min] {
			min = j
		}
	}
	return append(cyc[min:], cyc[:min]...)
}

func canonical(cyc []string) string { return strings.Join(cyc, ",") }
