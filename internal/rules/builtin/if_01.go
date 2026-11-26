package builtin

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// if01：设计中的接口在代码中恰有一个锚点，且锚点位于接口所属模块的路径内（AC-052-1）。
// 只检查路径内有可分析源文件（Release 2 为 Go）的模块的接口。
type if01 struct{}

func (if01) ID() string          { return "if.01" }
func (if01) Covers() []string    { return []string{"AC-052-1"} }
func (if01) Needs() engine.Input { return engine.InputSources }
func (if01) Fix() string {
	return "在接口实现处的声明紧前一行写 // harness:IF-n，每个接口只写一处，且位于所属模块的 Paths 内"
}

func (if01) Check(s *workspace.Snapshot) []engine.Violation {
	anchors := map[string][]lang.Decl{}
	for _, d := range s.Anchors() {
		anchors[d.ID] = append(anchors[d.ID], d)
	}
	var out []engine.Violation
	design := map[string]bool{}
	for _, e := range archEntities(s, "IF") {
		design[e.ID] = true
		mod := s.Project.ByID[e.Fields["Module"]]
		if mod == nil || !effective(mod) || !analyzable(s, mod) {
			continue
		}
		ds := anchors[e.ID]
		switch {
		case len(ds) == 0:
			out = append(out, at(e, "%s 在代码中没有锚点 // harness:%s", e.ID, e.ID))
		case len(ds) > 1:
			var where []string
			for _, d := range ds {
				where = append(where, fmt.Sprintf("%s:%d", d.File, d.Line))
			}
			for _, d := range ds {
				out = append(out, engine.Violation{File: d.File, Line: d.Line, Msg: fmt.Sprintf("%s 有 %d 个锚点：%s", e.ID, len(ds), strings.Join(where, "、"))})
			}
		case !workspace.MatchAny(mod.List("Paths"), ds[0].File):
			out = append(out, engine.Violation{File: ds[0].File, Line: ds[0].Line, Msg: fmt.Sprintf("%s 属于 %s，锚点却位于其路径（%s）之外", e.ID, mod.ID, mod.Fields["Paths"])})
		}
	}
	var ids []string
	for id := range anchors {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		for _, d := range anchors[id] {
			switch {
			case !design[id]:
				out = append(out, engine.Violation{File: d.File, Line: d.Line, Msg: "锚点引用的 " + id + " 不在详细设计中"})
			case d.Signature == "":
				out = append(out, engine.Violation{File: d.File, Line: d.Line, Msg: id + " 的锚点下一行不是声明"})
			}
		}
	}
	return out
}

// analyzable 判断模块路径内是否有可提取锚点的非测试源文件。
func analyzable(s *workspace.Snapshot, mod *model.Entity) bool {
	for _, f := range s.Sources {
		if _, ok := lang.AnalyzerFor(f.Path); ok && f.Kind == lang.NotTest && workspace.MatchAny(mod.List("Paths"), f.Path) {
			return true
		}
	}
	return false
}

// snapshotLine 返回接口在快照文件中的行号；找不到时为 1。
func snapshotLine(s *workspace.Snapshot, id string) int {
	src, _ := os.ReadFile(filepath.Join(s.Root, workspace.InterfacesFile))
	for i, l := range strings.Split(string(src), "\n") {
		if l == id {
			return i + 1
		}
	}
	return 1
}
