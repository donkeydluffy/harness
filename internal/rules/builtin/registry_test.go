package builtin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// 每条内置规则在 testdata/<规则编号>/ 下必须有 violation* 与 ok* 样本目录，并通过 Verify。
func TestRulesSamples_IF_014(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range Rules() {
		if seen[r.ID()] {
			t.Fatalf("规则编号重复：%s", r.ID())
		}
		seen[r.ID()] = true
		if len(r.Covers()) == 0 {
			t.Errorf("%s 没有声明覆盖的验收标准", r.ID())
		}
		dirs, _ := filepath.Glob(filepath.Join("testdata", r.ID(), "*"))
		var samples []engine.Sample
		for _, d := range dirs {
			if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
				continue
			}
			s, err := workspace.Load(d, workspace.LoadOptions{})
			if err != nil {
				t.Fatalf("%s: %v", d, err)
			}
			name := filepath.Base(d)
			samples = append(samples, engine.Sample{Name: name, Snapshot: s, WantViolation: !strings.HasPrefix(name, "ok")})
		}
		if err := engine.Verify(r, samples); err != nil {
			t.Error(err)
		}
	}
}
