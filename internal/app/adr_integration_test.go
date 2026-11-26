package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func adrDoc(id, status, extra string) string {
	return strings.Replace(sprintf(adr, status, "MOD-01", extra), "ADR-001", id, 1)
}

func checkADR(t *testing.T, files map[string]string) checked {
	t.Helper()
	return checkJSON(t, writeTree(t, files), "adr")
}

// ---- AC-010-1 ADR 必填字段 ----

func TestCheck_AC_010_1_N1(t *testing.T) {
	wantPass(t, checkADR(t, map[string]string{"docs/adr/ADR-001-store.md": adrDoc("ADR-001", "Accepted", "")}))
}

func TestCheck_AC_010_1_E1(t *testing.T) {
	src := strings.Replace(adrDoc("ADR-001", "Accepted", ""), "\n## 后果\n\n文字\n", "", 1)
	wantViolation(t, checkADR(t, map[string]string{"docs/adr/ADR-001-store.md": src}), "adr.01", "ADR-001", "## 后果")
}

func TestCheck_AC_010_1_E2(t *testing.T) {
	c := checkADR(t, map[string]string{"docs/adr/ADR-001-store.md": adrDoc("ADR-001", "Done", "")})
	wantViolation(t, c, "adr.01", "Done", "Proposed、Accepted、Superseded")
}

// ---- AC-010-2 取代关系 ----

func TestCheck_AC_010_2_N1(t *testing.T) {
	wantPass(t, checkADR(t, map[string]string{
		"docs/adr/ADR-002-a.md": adrDoc("ADR-002", "Superseded", "- Superseded-By: ADR-005\n"),
		"docs/adr/ADR-005-b.md": adrDoc("ADR-005", "Accepted", ""),
	}))
}

func TestCheck_AC_010_2_E1(t *testing.T) {
	c := checkADR(t, map[string]string{"docs/adr/ADR-002-a.md": adrDoc("ADR-002", "Superseded", "")})
	wantViolation(t, c, "adr.02", "ADR-002", "没有 Superseded-By")
}

func TestCheck_AC_010_2_E2(t *testing.T) {
	c := checkADR(t, map[string]string{"docs/adr/ADR-002-a.md": adrDoc("ADR-002", "Superseded", "- Superseded-By: ADR-009\n")})
	wantViolation(t, c, "adr.02", "不存在的 ADR-009")
}

func TestCheck_AC_010_2_B1(t *testing.T) {
	c := checkADR(t, map[string]string{
		"docs/adr/ADR-002-a.md": adrDoc("ADR-002", "Superseded", "- Superseded-By: ADR-005\n"),
		"docs/adr/ADR-005-b.md": adrDoc("ADR-005", "Superseded", "- Superseded-By: ADR-002\n"),
	})
	wantViolation(t, c, "adr.02", "环", "ADR-002 -> ADR-005 -> ADR-002")
}

// ---- AC-010-3 架构结构变化须同一提交的 ADR 引用 ----

// gitRepo 把 files 写入新仓库并提交，返回目录。
func gitRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := writeTree(t, files)
	git(t, dir, "init", "-q", "-b", "main")
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// stage 写入（或覆盖）文件并加入暂存区。
func stage(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git(t, dir, "add", "-A")
}

const mod03 = "\n### MOD-03 cache\n- Status: Active\n- Paths: internal/cache/**\n- Public: internal/cache/cache.go\n- Depends:\n- Interfaces:\n- Covers: FR-001\n- Core: false\n- Coverage: 60\n"

func TestCheck_AC_010_3_N1(t *testing.T) {
	dir := gitRepo(t, arch(t))
	stage(t, dir, map[string]string{
		"docs/architecture/overview.md": strings.Replace(archOverview, "\n### EXT-01", mod03+"\n### EXT-01", 1),
		"docs/adr/ADR-001-cache.md":     strings.Replace(adrDoc("ADR-001", "Accepted", ""), "- Relates: MOD-01", "- Relates: MOD-03", 1),
	})
	wantPass(t, checkJSON(t, dir, "adr"))
}

func TestCheck_AC_010_3_E1(t *testing.T) {
	dir := gitRepo(t, arch(t))
	stage(t, dir, map[string]string{
		"docs/architecture/overview.md": strings.Replace(archOverview, "- Depends: MOD-02\n", "- Depends: MOD-02, MOD-03\n", 1),
	})
	wantViolation(t, checkJSON(t, dir, "adr"), "adr.03", "MOD-01", "允许依赖", "没有 ADR 引用")
}

func TestCheck_AC_010_3_B1(t *testing.T) {
	dir := gitRepo(t, arch(t))
	stage(t, dir, map[string]string{
		"docs/architecture/overview.md": strings.Replace(archOverview, "### MOD-01 api\n", "### MOD-01 api 网关\n", 1) + "\n补充说明文字。\n",
	})
	wantPass(t, checkJSON(t, dir, "adr"))
}
