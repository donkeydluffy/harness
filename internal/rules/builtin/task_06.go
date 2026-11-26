package builtin

import (
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// task06：扩展任务（Extends 引用扩展点）的文件范围不能与扩展点封闭模块的路径重叠。
type task06 struct{}

func (task06) ID() string       { return "task.06" }
func (task06) Covers() []string { return []string{"AC-034-2"} }
func (task06) Fix() string {
	return "扩展任务只改扩展点的 Paths；确需修改封闭模块时先通过 ADR 调整扩展点设计"
}

func (task06) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, t := range tasks(s) {
		for _, x := range t.List("Extends") {
			ext := s.Project.ByID[x]
			if ext == nil {
				continue // task.01 报告
			}
			for _, cm := range ext.List("Closed") {
				mod := s.Project.ByID[cm]
				if mod == nil {
					continue
				}
				for _, p := range t.List("Scope") {
					for _, g := range mod.List("Paths") {
						if globsOverlap(p, g) {
							out = append(out, at(t, "%s 扩展 %s，文件范围 %s 与封闭模块 %s 的路径 %s 重叠", t.ID, x, p, cm, g))
						}
					}
				}
			}
		}
	}
	return out
}
