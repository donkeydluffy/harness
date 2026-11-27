package builtin

import (
	"fmt"
	"path"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/workspace"

	"github.com/donkeydluffy/harness/internal/rules/engine"
)

// deps04：跨模块 import 的目标包须在被引用模块的 Public 内：包目录中有匹配 Public 的非测试源文件（AC-035-2）。
// 模块内部与测试文件中的 import 不检查。
type deps04 struct{}

func (deps04) ID() string          { return "deps.04" }
func (deps04) Covers() []string    { return []string{"AC-035-2"} }
func (deps04) Needs() engine.Input { return engine.InputSources }
func (deps04) Fix() string {
	return "改为 import 被引用模块的公开包，或在架构设计中把该包加入模块的 Public（设计改动须由人重新批准）"
}

func (deps04) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, im := range crossImports(s) {
		if !publicDir(s, im.to.List("Public"), im.dir) {
			out = append(out, engine.Violation{File: im.file, Line: im.line,
				Msg: fmt.Sprintf("%s import 的 %s 不在 %s 的 Public（%s）内", im.from.ID, im.path, im.to.ID, im.to.Fields["Public"])})
		}
	}
	return out
}

// publicDir 判断目录中是否有匹配 Public 的非测试源文件。
func publicDir(s *workspace.Snapshot, public []string, dir string) bool {
	for _, f := range s.Sources {
		if f.Kind == lang.NotTest && path.Dir(f.Path) == path.Clean(dir) && workspace.MatchAny(public, f.Path) {
			return true
		}
	}
	return false
}
