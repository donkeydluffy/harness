package builtin

import (
	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// lint01：项目有需要 lint 的源文件时，lint 报告必须存在、源码哈希与当前一致、没有问题（AC-014-5）。
// 只读报告，不运行外部工具（ADR-011）。
type lint01 struct{}

func (lint01) ID() string          { return "lint.01" }
func (lint01) Covers() []string    { return []string{"AC-014-5"} }
func (lint01) Needs() engine.Input { return engine.InputSources | engine.InputTests }
func (lint01) Fix() string {
	return "修复格式化与静态检查的问题后运行 harness check lint 重新生成报告"
}

func (lint01) Check(s *workspace.Snapshot) []engine.Violation {
	if !needsLint(s) {
		return nil
	}
	at := func(msg string) []engine.Violation {
		return []engine.Violation{{File: testreport.LintFile, Line: 1, Msg: msg}}
	}
	switch {
	case s.LintErr != nil:
		return at("lint 报告无法读取：" + s.LintErr.Error())
	case s.Lint == nil:
		return at("没有 lint 报告；运行 harness check lint 生成")
	case s.Lint.SourceHash != s.SourceHash():
		return at("lint 报告已过期：生成报告后源码有改动；运行 harness check lint 重新生成")
	}
	var out []engine.Violation
	for _, is := range s.Lint.Issues {
		file, line := is.File, is.Line
		if file == "" {
			file, line = testreport.LintFile, 1
		}
		out = append(out, engine.Violation{File: file, Line: line, Msg: "[" + is.Tool + "] " + is.Msg})
	}
	return out
}

// needsLint 判断项目声明的语言中，是否有带语言配置包且存在源文件的语言。
func needsLint(s *workspace.Snapshot) bool {
	for _, name := range s.Config.Languages {
		if _, ok := lang.PackFor(name); !ok {
			continue
		}
		for _, f := range s.Sources {
			if f.Lang != nil && f.Lang.Name() == name {
				return true
			}
		}
	}
	return false
}
