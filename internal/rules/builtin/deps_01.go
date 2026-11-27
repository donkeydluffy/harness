package builtin

import (
	"fmt"
	"path"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// deps01：非测试 Go 文件中跨模块的 import，目标模块须在本模块的 Depends 中（AC-053-1）。
// 标准库、第三方包与测试文件中的 import 不检查。
type deps01 struct{}

func (deps01) ID() string          { return "deps.01" }
func (deps01) Covers() []string    { return []string{"AC-053-1"} }
func (deps01) Needs() engine.Input { return engine.InputSources }
func (deps01) Fix() string {
	return "去掉这处 import，或在架构设计中把目标模块加入本模块的 Depends（设计改动须由人重新批准）"
}

func (deps01) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, im := range crossImports(s) {
		if !contains(im.from.List("Depends"), im.to.ID) {
			out = append(out, engine.Violation{File: im.file, Line: im.line,
				Msg: fmt.Sprintf("%s import 了 %s 的 %s，但 %s 不在 %s 的 Depends 中", im.from.ID, im.to.ID, im.path, im.to.ID, im.from.ID)})
		}
	}
	return out
}

// crossImport 是非测试源文件中指向另一个模块的 import。
type crossImport struct {
	file, path, dir string // 所在文件、import 路径、目标目录（仓库相对）
	line            int
	from, to        *model.Entity
}

// crossImports 返回全部跨模块的 import；目标在 sourcePaths 之外或不属于任何模块时跳过。
func crossImports(s *workspace.Snapshot) []crossImport {
	mods := modules(s)
	var out []crossImport
	for _, f := range s.Sources {
		a, ok := lang.AnalyzerFor(f.Path)
		from := owner(mods, f.Path)
		if !ok || f.Kind != lang.NotTest || from == nil {
			continue
		}
		src, err := f.Content()
		if err != nil {
			continue
		}
		for _, im := range a.Imports(src) {
			dir, ok := a.Resolve(s.Root, im.Path)
			if !ok {
				continue // 标准库与第三方包
			}
			if to := dirOwner(s, mods, dir); to != nil && to != from {
				out = append(out, crossImport{file: f.Path, path: im.Path, dir: dir, line: im.Line, from: from, to: to})
			}
		}
	}
	return out
}

// modules 返回有效的模块。
func modules(s *workspace.Snapshot) []*model.Entity {
	var out []*model.Entity
	for _, m := range archEntities(s, "MOD") {
		if effective(m) {
			out = append(out, m)
		}
	}
	return out
}

// owner 返回路径所属的模块：多个模块的 Paths 都匹配时取匹配的 glob 前缀最长的一个。
func owner(mods []*model.Entity, p string) *model.Entity {
	var best *model.Entity
	n := -1
	for _, m := range mods {
		for _, g := range m.List("Paths") {
			if workspace.Match(g, p) && len(globBase(g)) > n {
				best, n = m, len(globBase(g))
			}
		}
	}
	return best
}

// dirOwner 返回目录中非测试源文件所属的模块；目录中没有源文件时返回 nil。
func dirOwner(s *workspace.Snapshot, mods []*model.Entity, dir string) *model.Entity {
	for _, f := range s.Sources {
		if f.Kind == lang.NotTest && path.Dir(f.Path) == path.Clean(dir) {
			if m := owner(mods, f.Path); m != nil {
				return m
			}
		}
	}
	return nil
}
