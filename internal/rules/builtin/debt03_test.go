package builtin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/workspace"
)

// 同一行用不同写法引用同一条已解决的技术债时，位置只列一次（T-032）。
func TestDebt03ListsLocationOnce(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"docs/debt.md":  "# 技术债清单\n\n### DEBT-001 x\n- Status: Resolved\n- Module: MOD-01\n\n说明。\n",
		"internal/a.go": "package a\n\n// DEBT-1 即 DEBT-001\nfunc F() {}\n",
	}
	for p, c := range files {
		os.MkdirAll(filepath.Dir(filepath.Join(dir, p)), 0o755)
		os.WriteFile(filepath.Join(dir, p), []byte(c), 0o644)
	}
	s, err := workspace.Load(dir, workspace.LoadOptions{WithSources: true})
	if err != nil {
		t.Fatal(err)
	}
	vs := debt03{}.Check(s)
	if len(vs) != 1 || strings.Count(vs[0].Msg, "internal/a.go:3") != 1 {
		t.Fatalf("应只列一次位置：%v", vs)
	}
}
