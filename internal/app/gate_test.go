package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/donkeydluffy/harness/internal/gate"
)

func TestGate_IF_025(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "harness.lock"), []byte(`{"version": 2}`), 0o644)
	if d := Gate(dir, gate.Request{Kind: gate.Write, Paths: []string{"docs/PRD.md"}}); !d.Allow {
		t.Errorf("docs/ 应可写：%+v", d)
	}
	if d := Gate(dir, gate.Request{Kind: gate.Write, Paths: []string{"harness.lock"}}); d.Allow || d.Rule != "gate.protect" {
		t.Errorf("锁文件应受保护：%+v", d)
	}
	os.WriteFile(filepath.Join(dir, "harness.lock"), []byte(`{`), 0o644)
	if d := Gate(dir, gate.Request{Kind: gate.Write, Paths: []string{"docs/PRD.md"}}); d.Allow || d.Rule != "gate.error" {
		t.Errorf("无法加载项目时应拒绝：%+v", d)
	}
}
