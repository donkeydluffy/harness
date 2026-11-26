package app_test

import "testing"

const doneTasks = "# 任务\n\n- Status: Draft\n\n### T-001 a\n- Status: Done\n- Covers: FR-001\n- Design: MOD-01\n- Scope: src/**\n- Depends:\n"

func smokeRepo(t *testing.T, cfg string) string {
	t.Helper()
	return writeTree(t, map[string]string{"harness.json": cfg, "docs/tasks.md": doneTasks})
}

// ---- AC-055-1 启动与冒烟测试命令声明 ----

func TestCheck_AC_055_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, smokeRepo(t, `{"start": "go run ./cmd/app", "smoke": "go run ./cmd/app --version"}`), "smoke.01"))
}

func TestCheck_AC_055_1_E1(t *testing.T) {
	wantViolation(t, checkJSON(t, smokeRepo(t, `{"start": "go run ./cmd/app"}`), "smoke.01"), "smoke.01", "T-001", "冒烟测试命令")
}

func TestCheck_AC_055_1_B1(t *testing.T) {
	wantPass(t, checkJSON(t, smokeRepo(t, `{"start": "none: 库项目，没有可启动的应用"}`), "smoke.01"))
}

func TestCheck_AC_055_1_E2(t *testing.T) {
	wantViolation(t, checkJSON(t, smokeRepo(t, `{"start": "none"}`), "smoke.01"), "smoke.01", "没有写明原因")
}
