package builtin

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// ci01：项目的 CI 配置中至少有一份运行 harness check（AC-043-2）。只检查执行过 harness init（有 harness.json）的项目。
type ci01 struct{}

func (ci01) ID() string       { return "ci.01" }
func (ci01) Covers() []string { return []string{"AC-043-2"} }
func (ci01) Fix() string {
	return "在 CI 配置中加入运行 harness check 的步骤，或运行 harness init --lang <语言> --ci github|gitlab 生成 CI 配置"
}

// ciPatterns 是识别为 CI 配置的文件。
var ciPatterns = []string{
	".github/workflows/*.yml", ".github/workflows/*.yaml", ".gitlab-ci.yml", ".circleci/config.yml",
	"azure-pipelines.yml", "bitbucket-pipelines.yml", "Jenkinsfile",
}

func (ci01) Check(s *workspace.Snapshot) []engine.Violation {
	if _, err := os.Stat(filepath.Join(s.Root, workspace.ConfigFile)); err != nil {
		return nil
	}
	var files []string
	for _, p := range ciPatterns {
		m, _ := filepath.Glob(filepath.Join(s.Root, filepath.FromSlash(p)))
		files = append(files, m...)
	}
	sort.Strings(files)
	var rels []string
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err == nil && strings.Contains(string(b), "harness check") {
			return nil
		}
		rel, _ := filepath.Rel(s.Root, f)
		rels = append(rels, filepath.ToSlash(rel))
	}
	if len(rels) == 0 {
		return []engine.Violation{{File: workspace.ConfigFile, Line: 1,
			Msg: "项目没有任何 CI 配置；运行 harness init --lang <语言> --ci github|gitlab 生成"}}
	}
	return []engine.Violation{{File: rels[0], Line: 1,
		Msg: "CI 配置都不包含 harness check 步骤：" + strings.Join(rels, ", ")}}
}
