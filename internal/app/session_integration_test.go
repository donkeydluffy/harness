package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"harness/internal/adapter/claudecode"
)

// sessionDir 返回全量检查（不含覆盖类规则）通过的项目：PRD 与架构夹具，处于实现阶段。
func sessionDir(t *testing.T, extra map[string]string) string {
	t.Helper()
	files := traced(t, basePRD)
	for k, v := range extra {
		files[k] = v
	}
	if _, ok := files["harness.json"]; ok {
		// 有 harness.json 的项目视为已初始化，cfg.02 要求 hook 配置与生成结果一致。
		p, b := claudecode.Adapter{}.HookConfig("harness")
		files[p] = string(b)
	}
	dir := writeTree(t, files)
	if r := runHarness(t, dir, "lock"); r.Code != 0 {
		t.Fatalf("lock：%s%s", r.Stdout, r.Stderr)
	}
	return dir
}

func stopHook(t *testing.T, dir string) hookOut {
	t.Helper()
	return hook(t, dir, "Stop", map[string]any{})
}

func startHook(t *testing.T, dir string) hookOut {
	t.Helper()
	h := hook(t, dir, "SessionStart", map[string]any{})
	if h.Deny || h.Code != 0 {
		t.Fatalf("会话开始不应拒绝：%s", h.Raw)
	}
	return h
}

// breakPRD 引入一条 prd 规则违规（验收标准含模糊词）。
func breakPRD(t *testing.T, dir string) {
	edit(t, dir, "- AC-001-1: 生成骨架。", "- AC-001-1: 快速生成骨架。")
}

func fixPRD(t *testing.T, dir string) {
	edit(t, dir, "- AC-001-1: 快速生成骨架。", "- AC-001-1: 生成骨架。")
}

// ---- AC-012-3 写后产物检查 ----

func TestGate_AC_012_3_N1(t *testing.T) {
	dir := sessionDir(t, nil)
	h := hook(t, dir, "PostToolUse", editTool(dir, "docs/PRD.md", "x", "y"))
	if h.Deny || !strings.Contains(h.Reason, "harness check prd") || !strings.Contains(h.Reason, "通过") {
		t.Fatalf("反馈：%s", h.Raw)
	}
}

func TestGate_AC_012_3_B1(t *testing.T) {
	dir := sessionDir(t, nil)
	if h := hook(t, dir, "PostToolUse", writeTool(dir, "internal/api/a.go", "package api\n")); h.Raw != "" {
		t.Fatalf("非产物文件不应运行检查：%s", h.Raw)
	}
}

func TestGate_AC_012_3_E1(t *testing.T) {
	dir := sessionDir(t, nil)
	breakPRD(t, dir)
	h := hook(t, dir, "PostToolUse", editTool(dir, "docs/PRD.md", "x", "y"))
	if !strings.Contains(h.Reason, "[prd.") || !strings.Contains(h.Reason, "docs/PRD.md:") || !strings.Contains(h.Reason, "快速") {
		t.Fatalf("反馈应含规则编号与行号：%s", h.Raw)
	}
}

// ---- AC-012-4 会话结束检查 ----

func TestGate_AC_012_4_N1(t *testing.T) {
	dir := sessionDir(t, nil)
	wantAllow(t, stopHook(t, dir))
}

func TestGate_AC_012_4_E1(t *testing.T) {
	dir := sessionDir(t, nil)
	breakPRD(t, dir)
	wantDeny(t, stopHook(t, dir), "gate.stop", "[prd.", "第 1 次阻止")
}

func TestGate_AC_012_4_B1(t *testing.T) {
	dir := sessionDir(t, map[string]string{"docs/progress.md": "# 进度\n\n### 2026-11-26 T-001\n- 结论：进行中。\n- 遗留：无。\n"})
	breakPRD(t, dir)
	for i := 1; i <= 3; i++ {
		wantDeny(t, stopHook(t, dir), "gate.stop")
	}
	h := stopHook(t, dir)
	wantAllow(t, h)
	b, _ := os.ReadFile(filepath.Join(dir, "docs/progress.md"))
	if !strings.Contains(string(b), "## 未完成") || !strings.Contains(string(b), "[prd.") || !strings.Contains(string(b), "### 2026-11-26 T-001") {
		t.Fatalf("进度文件：\n%s", b)
	}
	start := startHook(t, dir)
	if !strings.HasPrefix(start.Reason, "上次会话遗留：") || !strings.Contains(start.Reason, "[prd.") {
		t.Fatalf("下次会话开始应先注入未完成项：%s", start.Reason)
	}
	// 计数在放行后清零：下一次结束检查重新从第 1 次阻止开始。
	wantDeny(t, stopHook(t, dir), "第 1 次阻止")
}

func TestGate_AC_012_4_B2(t *testing.T) {
	dir := sessionDir(t, nil)
	breakPRD(t, dir)
	wantDeny(t, stopHook(t, dir), "第 1 次阻止")
	wantDeny(t, stopHook(t, dir), "第 2 次阻止")
	fixPRD(t, dir)
	wantAllow(t, stopHook(t, dir))
	breakPRD(t, dir)
	wantDeny(t, stopHook(t, dir), "第 1 次阻止")
}

// ---- AC-024-1 会话开始注入 ----

func TestGate_AC_024_1_N1(t *testing.T) {
	dir := sessionDir(t, nil)
	write(t, dir, "docs/tasks.md", activeTasks)
	edit(t, dir, "## 8. 变更日志", oq003)
	breakPRD(t, dir)
	r := startHook(t, dir).Reason
	for _, want := range []string{"阶段：需求", "进行中任务：T-001", "未通过的检查：", "Open 待澄清项：OQ-003"} {
		if !strings.Contains(r, want) {
			t.Fatalf("注入内容缺少 %q：\n%s", want, r)
		}
	}
}

func TestGate_AC_024_1_B1(t *testing.T) {
	if r := startHook(t, sessionDir(t, nil)).Reason; r != "阶段：需求，无进行中任务" {
		t.Fatalf("注入内容：%q", r)
	}
}

// ---- AC-024-3 冒烟测试 ----

func TestGate_AC_024_3_N1(t *testing.T) {
	dir := sessionDir(t, map[string]string{"harness.json": `{"smoke": "true"}`})
	if r := startHook(t, dir).Reason; !strings.HasSuffix(r, "\n冒烟通过") {
		t.Fatalf("注入内容：%q", r)
	}
}

func TestGate_AC_024_3_E1(t *testing.T) {
	dir := sessionDir(t, map[string]string{"harness.json": `{"smoke": "seq 1 80; exit 3"}`})
	r := startHook(t, dir).Reason
	if !strings.Contains(r, "冒烟失败") || !strings.Contains(r, "\n31\n") || strings.Contains(r, "\n30\n") || !strings.HasSuffix(r, "\n80") {
		t.Fatalf("应注入最后 50 行：%q", r)
	}
}

func TestGate_AC_024_3_B1(t *testing.T) {
	if r := startHook(t, sessionDir(t, nil)).Reason; strings.Contains(r, "冒烟") {
		t.Fatalf("未声明冒烟命令时不应运行：%q", r)
	}
}
