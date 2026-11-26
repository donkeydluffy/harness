package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/adapter"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Expected 返回纳入版本库的生成文件的期望内容（IF-021）：claudecode 的 hook 配置必有；
// CI 配置只比对已存在的那一份（缺少 CI 配置由 ci.01 报告）。git hook 不在版本库中，不比对。
func Expected(s *workspace.Snapshot) (map[string][]byte, error) {
	out := map[string][]byte{}
	if a, ok := adapter.Get("claudecode"); ok {
		p, b := a.HookConfig("harness")
		out[p] = b
	}
	for ci, p := range ciFiles {
		if _, err := os.Stat(filepath.Join(s.Root, filepath.FromSlash(p))); err != nil {
			continue
		}
		b, err := templates.ReadFile("templates/ci/" + ci + ".yml")
		if err != nil {
			return nil, err
		}
		out[p] = render(b, InitOptions{})
	}
	return out, nil
}

// Rule 返回本模块注册的规则 cfg.02：生成文件与 Expected 一致（AC-041-2）。
func Rule() engine.Rule { return cfg02{} }

type cfg02 struct{}

func (cfg02) ID() string       { return "cfg.02" }
func (cfg02) Covers() []string { return []string{"AC-041-2"} }
func (cfg02) Fix() string {
	return "删除被改动的生成文件后运行 harness init --lang <语言> 重新生成；要定制内容时改 harness 的模板而不是生成文件"
}

func (cfg02) Check(s *workspace.Snapshot) []engine.Violation {
	// 没有 harness.json 说明项目尚未执行 harness init，没有生成文件可比对。
	if _, err := os.Stat(filepath.Join(s.Root, workspace.ConfigFile)); err != nil {
		return nil
	}
	want, err := Expected(s)
	if err != nil {
		return []engine.Violation{{File: workspace.ConfigFile, Line: 1, Msg: "无法读取内嵌模板：" + err.Error()}}
	}
	langs := strings.Join(s.Config.Languages, ",")
	if langs == "" {
		langs = "<语言>"
	}
	init := "harness init --lang " + langs
	var paths []string
	for p := range want {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var out []engine.Violation
	for _, p := range paths {
		got, err := os.ReadFile(filepath.Join(s.Root, filepath.FromSlash(p)))
		switch {
		case err != nil:
			out = append(out, engine.Violation{File: p, Line: 1, Msg: "缺少生成文件 " + p + "；运行 " + init + " 生成"})
		case !bytes.Equal(got, want[p]):
			out = append(out, engine.Violation{File: p, Line: firstDiff(got, want[p]),
				Msg: p + " 与重新生成的内容不一致；删除后运行 " + init + " 重新生成"})
		}
	}
	return out
}

// firstDiff 返回两段内容第一处不同所在的行号。
func firstDiff(a, b []byte) int {
	la, lb := strings.Split(string(a), "\n"), strings.Split(string(b), "\n")
	for i := 0; i < len(la) && i < len(lb); i++ {
		if la[i] != lb[i] {
			return i + 1
		}
	}
	if len(la) < len(lb) {
		return len(la)
	}
	return len(lb)
}
