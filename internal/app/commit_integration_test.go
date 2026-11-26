package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const commitArch = "# 架构总览\n\n- Status: Draft\n\n### MOD-01 core\n- Status: Active\n- Paths: internal/core/**\n\n### MOD-02 plugin\n- Status: Active\n- Paths: internal/plugin/**\n\n### EXT-01 新增插件\n- Status: Active\n- Closed: MOD-01\n"

const commitTasks = "# 任务列表\n\n- Status: Draft\n\n### T-003 核心\n- Status: InProgress\n- Scope: internal/core/**\n\n### T-004 插件\n- Status: Todo\n- Scope: internal/plugin/**\n- Extends: EXT-01\n\n### T-005 维护\n- Status: Todo\n- Scope: internal/core/**, internal/plugin/**\n"

// commitRepo 建立带架构与任务列表的仓库，再提交 files，提交信息为 msg；返回目录。
func commitRepo(t *testing.T, files map[string]string, msg string) string {
	t.Helper()
	dir := gitRepo(t, map[string]string{
		"docs/architecture/overview.md": commitArch,
		"docs/tasks.md":                 commitTasks,
		"internal/core/core.go":         "package core\n",
	})
	stage(t, dir, files)
	git(t, dir, "commit", "-q", "-m", msg)
	return dir
}

func checkLast(t *testing.T, dir string) checked {
	t.Helper()
	return checkJSON(t, dir, "commits", "HEAD~1..HEAD")
}

func lines(n int) string { return strings.Repeat("x\n", n) }

var coreChange = map[string]string{"internal/core/core.go": "package core\n\nvar X = 1\n"}

// ---- AC-013-1 Task 尾注 ----

func TestCheck_AC_013_1_N1(t *testing.T) {
	wantPass(t, checkLast(t, commitRepo(t, coreChange, "改动\n\nTask: T-3")))
}

func TestCheck_AC_013_1_E1(t *testing.T) {
	// commit-msg hook 以 --msg-file 调用：检查暂存区与即将形成的提交信息。
	dir := gitRepo(t, map[string]string{"docs/tasks.md": commitTasks})
	stage(t, dir, coreChange)
	msg := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	os.WriteFile(msg, []byte("改动\n\n# 以 # 开头的行是 git 注释\n"), 0o644)
	wantViolation(t, checkJSON(t, dir, "commit", "--msg-file", msg), "commit.01", "暂存区", "没有 Task 尾注")
	// 写上尾注后通过。
	os.WriteFile(msg, []byte("改动\n\nTask: T-003\n"), 0o644)
	wantPass(t, checkJSON(t, dir, "commit", "--msg-file", msg))
}

func TestCheck_AC_013_1_E2(t *testing.T) {
	c := checkLast(t, commitRepo(t, coreChange, "改动\n\nTask: T-003\nTask: T-005"))
	wantViolation(t, c, "commit.01", "2 个 Task 尾注", "T-003, T-005")
}

func TestCheck_AC_013_1_E3(t *testing.T) {
	wantViolation(t, checkLast(t, commitRepo(t, coreChange, "改动\n\nTask: T-99")), "commit.01", "T-99 不存在")
}

// ---- AC-013-5 需求提交使用 Change 尾注 ----

const changePRD = "# PRD: x\n\n- Status: Draft\n\n## 变更日志\n\n| 编号 | 日期 | 动作 | 对象 | 说明 |\n|---|---|---|---|---|\n| CHG-018 | 2026-11-26 | 修改 | FR-013 | 说明 |\n"

// changeCommit 提交 PRD（含变更日志）、锁文件与 extra 中的文件，提交信息为 msg。
func changeCommit(t *testing.T, extra map[string]string, msg string) checked {
	t.Helper()
	files := map[string]string{"docs/PRD.md": changePRD, "harness.lock": "{}\n"}
	for k, v := range extra {
		files[k] = v
	}
	return checkLast(t, commitRepo(t, files, msg))
}

func TestCheck_AC_013_5_N1(t *testing.T) {
	wantPass(t, changeCommit(t, nil, "澄清需求\n\nChange: CHG-018"))
}

func TestCheck_AC_013_5_E1(t *testing.T) {
	wantViolation(t, changeCommit(t, coreChange, "澄清需求\n\nChange: CHG-018"), "commit.01", "internal/core/core.go", "需要关联任务")
}

func TestCheck_AC_013_5_E2(t *testing.T) {
	wantViolation(t, changeCommit(t, nil, "澄清需求\n\nChange: CHG-099"), "commit.01", "CHG-099 不在变更日志中")
}

func TestCheck_AC_013_5_E3(t *testing.T) {
	wantViolation(t, changeCommit(t, nil, "澄清需求\n\nTask: T-003\nChange: CHG-018"), "commit.01", "同时有 Task 与 Change 尾注")
}

func TestCheck_AC_013_5_B1(t *testing.T) {
	wantViolation(t, changeCommit(t, nil, "澄清需求\n\nChange: CHG-018\nChange: CHG-018"), "commit.01", "2 个 Change 尾注")
}

// ---- AC-013-2 改动在任务范围内 ----

func TestCheck_AC_013_2_N1(t *testing.T) {
	wantPass(t, checkJSON(t, commitRepo(t, coreChange, "改动\n\nTask: T-003"), "commit", "--commits", "HEAD~1..HEAD"))
}

func TestCheck_AC_013_2_B1(t *testing.T) {
	files := map[string]string{"internal/core/core.go": "package core\n\nvar X = 1\n", "docs/tasks.md": commitTasks + "\n补充说明。\n"}
	wantPass(t, checkLast(t, commitRepo(t, files, "改动\n\nTask: T-003")))
}

func TestCheck_AC_013_2_B2(t *testing.T) {
	files := map[string]string{"internal/core/core.go": "package core\n\nvar X = 1\n", "harness.json": "{\"commitMaxLines\": 400}\n"}
	wantPass(t, checkLast(t, commitRepo(t, files, "改动\n\nTask: T-003")))
}

func TestCheck_AC_013_2_E1(t *testing.T) {
	files := map[string]string{"internal/core/core.go": "package core\n\nvar X = 1\n", "internal/plugin/p.go": "package plugin\n"}
	v := wantViolation(t, checkLast(t, commitRepo(t, files, "改动\n\nTask: T-003")), "commit.02", "internal/plugin/p.go", "T-003")
	if v.File != "internal/plugin/p.go" {
		t.Fatalf("位置：%s", v.File)
	}
}

// ---- AC-013-4 提交行数上限 ----

func TestCheck_AC_013_4_N1(t *testing.T) {
	wantPass(t, checkLast(t, commitRepo(t, map[string]string{"internal/core/big.go": lines(400)}, "大\n\nTask: T-003")))
}

func TestCheck_AC_013_4_E1(t *testing.T) {
	c := checkLast(t, commitRepo(t, map[string]string{"internal/core/big.go": lines(401)}, "大\n\nTask: T-003"))
	wantViolation(t, c, "commit.03", "401 行", "上限 400")
}

func TestCheck_AC_013_4_B1(t *testing.T) {
	files := map[string]string{
		"internal/core/big.go":  lines(100),
		"docs/PRD.md":           lines(500),
		".claude/settings.json": lines(400),
	}
	wantPass(t, checkLast(t, commitRepo(t, files, "大\n\nTask: T-003")))
}

// ---- AC-034-1 扩展任务不改封闭模块 ----

func TestCheck_AC_034_1_N1(t *testing.T) {
	wantPass(t, checkLast(t, commitRepo(t, map[string]string{"internal/plugin/p.go": "package plugin\n"}, "插件\n\nTask: T-004")))
}

func TestCheck_AC_034_1_E1(t *testing.T) {
	files := map[string]string{"internal/plugin/p.go": "package plugin\n", "internal/core/core.go": "package core\n\nvar X = 1\n"}
	// T-004 的范围也放宽到核心，确认失败来自开闭校验而不是范围检查。
	dir := commitRepo(t, files, "插件\n\nTask: T-004")
	os.WriteFile(filepath.Join(dir, "docs/tasks.md"), []byte(strings.Replace(commitTasks, "- Scope: internal/plugin/**\n- Extends", "- Scope: internal/plugin/**, internal/core/**\n- Extends", 1)), 0o644)
	wantViolation(t, checkJSON(t, dir, "commit.04", "--commits", "HEAD~1..HEAD"), "commit.04", "T-004", "封闭模块 MOD-01", "internal/core/core.go")
}

func TestCheck_AC_034_1_B1(t *testing.T) {
	wantPass(t, checkLast(t, commitRepo(t, coreChange, "维护\n\nTask: T-005")))
}
