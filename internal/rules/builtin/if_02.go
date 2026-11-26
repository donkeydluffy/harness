package builtin

import (
	"sort"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// if02：接口快照与当前源码重新提取的结果一致（AC-052-2）。
type if02 struct{}

func (if02) ID() string          { return "if.02" }
func (if02) Covers() []string    { return []string{"AC-052-2"} }
func (if02) Needs() engine.Input { return engine.InputSources }
func (if02) Fix() string {
	return "运行 harness snapshot 重新生成 " + workspace.InterfacesFile + " 并提交"
}

func (if02) Check(s *workspace.Snapshot) []engine.Violation {
	want := s.ExtractInterfaces()
	got, ok := s.Interfaces()
	if !ok {
		if len(want) == 0 {
			return nil
		}
		return []engine.Violation{{File: workspace.InterfacesFile, Line: 1, Msg: "源码中有接口锚点，但没有接口快照文件"}}
	}
	ids := map[string]bool{}
	for id := range want {
		ids[id] = true
	}
	for id := range got {
		ids[id] = true
	}
	var sorted []string
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Strings(sorted)
	var out []engine.Violation
	for _, id := range sorted {
		code, snap := want[id], got[id]
		msg := ""
		switch {
		case code == snap:
			continue
		case snap == "":
			msg = id + " 在源码中有锚点，快照中没有"
		case code == "":
			msg = id + " 在快照中，源码中已没有它的锚点"
		default:
			msg = id + " 的快照与源码不一致：快照为 " + snap + "，源码为 " + code
		}
		out = append(out, engine.Violation{File: workspace.InterfacesFile, Line: snapshotLine(s, id), Msg: msg})
	}
	return out
}
