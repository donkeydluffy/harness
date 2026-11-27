package app_test

import (
	"strings"
	"testing"
)

// ---- AC-048-1 阶段命令已删除 ----

func TestCLI_AC_048_1_N1(t *testing.T) {
	dir := writeTree(t, map[string]string{"docs/PRD.md": basePRD})
	for _, args := range [][]string{{"phase", "list"}, {"phase", "next"}} {
		r := runHarness(t, dir, args...)
		if r.Code == 0 || !strings.Contains(r.Stderr, "阶段已取消，改由批准状态判定") {
			t.Fatalf("harness %v 应提示已删除并以非 0 退出：%d %q", args, r.Code, r.Stderr)
		}
	}
}

func TestLock_AC_048_1_B1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	editFile(t, dir, "harness.lock", `"version": 2`, `"phase": "implementation",
  "version": 2`)
	wantPass(t, checkJSON(t, dir, "prd", "arch", "approve.01"))
	lockOK(t, dir)
	if lk := readFile(t, dir, "harness.lock"); strings.Contains(lk, "phase") {
		t.Fatalf("harness lock 后不应再有 phase 字段：%s", lk)
	}
}
