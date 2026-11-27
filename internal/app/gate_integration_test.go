package app_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gatePRD 是带一段正文的 basePRD。
var gatePRD = strings.Replace(basePRD, "- Current-Release: 1\n", "- Current-Release: 1\n\n正文第一段。\n", 1)

// gateDir 返回带架构夹具的项目；tasks 为任务列表内容（空串时不写）。approved 为真时锁文件为第 2 版，
// PRD、MOD-01、MOD-02 与任务列表中的 T-001 已按单位批准，InProgress 的 T-001 可以修改 internal/api/**。archEdits 同 arch()。
func gateDir(t *testing.T, approved bool, tasks string, archEdits ...string) string {
	t.Helper()
	files := arch(t, archEdits...)
	files["docs/PRD.md"] = gatePRD
	files["internal/api/a.go"] = "package api\n"
	files["other/b.go"] = "package other\n"
	if !approved {
		if tasks != "" {
			files["docs/tasks.md"] = tasks
		}
		return writeTree(t, files)
	}
	dir := writeTree(t, files)
	if r := runHarness(t, dir, "lock"); r.Code != 0 {
		t.Fatalf("harness lock 失败：%s%s", r.Stdout, r.Stderr)
	}
	mustApprove(t, dir, "prd", "MOD-02", "MOD-01")
	if tasks != "" {
		writeFile(t, dir, "docs/tasks.md", tasks)
	}
	if strings.Contains(tasks, "### T-001") {
		mustApprove(t, dir, "T-001")
	}
	return dir
}

var activeTasks = taskDoc(task("T-001", "InProgress", "FR-001", "MOD-01", "internal/api/**", "", ""))

type hookOut struct {
	Deny   bool
	Reason string
	Raw    string
	Code   int
}

// hook 以 Claude Code hook 的方式调用 harness gate claudecode <event>。
func hook(t *testing.T, dir, event string, in map[string]any) hookOut {
	t.Helper()
	in["session_id"], in["cwd"] = "s1", dir
	b, _ := json.Marshal(in)
	return hookRaw(t, dir, event, b)
}

func hookRaw(t *testing.T, dir, event string, stdin []byte) hookOut {
	t.Helper()
	cmd := exec.Command(harnessBin, "gate", "claudecode", event)
	cmd.Dir, cmd.Stdin = dir, bytes.NewReader(stdin)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	var o struct {
		Decision string
		Reason   string
		Hook     struct {
			PermissionDecision       string
			PermissionDecisionReason string
			AdditionalContext        string
		} `json:"hookSpecificOutput"`
	}
	if out.Len() > 0 {
		if err := json.Unmarshal(out.Bytes(), &o); err != nil {
			t.Fatalf("输出不是 JSON：%s", out.String())
		}
	}
	h := hookOut{Raw: out.String(), Code: code}
	switch {
	case o.Hook.PermissionDecision == "deny":
		h.Deny, h.Reason = true, o.Hook.PermissionDecisionReason
	case o.Decision == "block":
		h.Deny, h.Reason = true, o.Reason
	default:
		h.Reason = o.Hook.AdditionalContext
	}
	return h
}

func writeTool(dir, rel, content string) map[string]any {
	return map[string]any{"tool_name": "Write", "tool_input": map[string]any{"file_path": filepath.Join(dir, rel), "content": content}}
}

func editTool(dir, rel, old, new string) map[string]any {
	return map[string]any{"tool_name": "Edit", "tool_input": map[string]any{"file_path": filepath.Join(dir, rel), "old_string": old, "new_string": new}}
}

func bashTool(cmd string) map[string]any {
	return map[string]any{"tool_name": "Bash", "tool_input": map[string]any{"command": cmd}}
}

func wantAllow(t *testing.T, h hookOut) {
	t.Helper()
	if h.Deny || h.Code != 0 {
		t.Fatalf("期望放行：%d %s", h.Code, h.Raw)
	}
}

func wantDeny(t *testing.T, h hookOut, parts ...string) {
	t.Helper()
	if !h.Deny || h.Code != 0 {
		t.Fatalf("期望拒绝：%d %s", h.Code, h.Raw)
	}
	for _, p := range parts {
		if !strings.Contains(h.Reason, p) {
			t.Fatalf("拒绝原因缺少 %q：%s", p, h.Reason)
		}
	}
}

// ---- AC-012-1 没有 InProgress 任务时拒绝源码写入 ----

func TestGate_AC_012_1_N1(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantAllow(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/a.go", "package api\n\nvar X = 1\n")))
}

func TestGate_AC_012_1_E1(t *testing.T) {
	dir := gateDir(t, true, taskDoc(task("T-001", "Todo", "FR-001", "MOD-01", "internal/api/**", "", "")))
	wantDeny(t, hook(t, dir, "PreToolUse", editTool(dir, "internal/api/a.go", "package api", "package api // x")), "gate.write", "没有 InProgress 任务")
}

func TestGate_AC_012_1_E2(t *testing.T) {
	dir := gateDir(t, true, "")
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("echo x > src/a.go")), "gate.write", "src/a.go")
}

func TestGate_AC_012_1_B1(t *testing.T) {
	dir := gateDir(t, false, "")
	wantAllow(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "正文第一段。", "正文第一段，补充。")))
}

// ---- AC-012-2 写入须在任务范围内 ----

func TestGate_AC_012_2_N1(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantAllow(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/new.go", "package api\n")))
}

func TestGate_AC_012_2_E1(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", writeTool(dir, "other/b.go", "package other\n")), "gate.write", "other/b.go", "T-001：internal/api/**")
}

func TestGate_AC_012_2_B1(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/../../other/b.go", "x")), "other/b.go")
	if err := os.Symlink(filepath.Join(dir, "other"), filepath.Join(dir, "internal/api/link")); err != nil {
		t.Fatal(err)
	}
	wantDeny(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/link/c.go", "x")), "other/c.go")
}

func TestGate_AC_012_2_E2(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("sed -i '' 's/a/b/' other/b.go")), "gate.write", "other/b.go")
}

// 命令无法静态判定时放行执行，事后比对恢复范围外改动（FLOW-03）。
func TestGate_FLOW_03(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	cmd := bashTool("sh ./script.sh")
	wantAllow(t, hook(t, dir, "PreToolUse", cmd))
	os.WriteFile(filepath.Join(dir, "other/b.go"), []byte("changed\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "other/new.go"), []byte("new\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "internal/api/a.go"), []byte("package api // ok\n"), 0o644)
	h := hook(t, dir, "PostToolUse", cmd)
	if b, _ := os.ReadFile(filepath.Join(dir, "other/b.go")); string(b) != "package other\n" {
		t.Fatalf("范围外修改未恢复：%q", b)
	}
	if _, err := os.Stat(filepath.Join(dir, "other/new.go")); !os.IsNotExist(err) {
		t.Fatal("范围外新建文件未删除")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "internal/api/a.go")); string(b) != "package api // ok\n" {
		t.Fatal("范围内修改被恢复")
	}
	if !strings.Contains(h.Reason, "other/b.go（modified）已恢复") || !strings.Contains(h.Reason, "other/new.go") || strings.Contains(h.Reason, "internal/api") {
		t.Fatalf("反馈：%s", h.Raw)
	}
}

// ---- AC-012-5 适配器 ----

func TestGate_AC_012_5_N1(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	h := hook(t, dir, "PreToolUse", writeTool(dir, "other/b.go", "x"))
	if !strings.Contains(h.Raw, `"hookEventName":"PreToolUse"`) || !strings.Contains(h.Raw, `"permissionDecision":"deny"`) {
		t.Fatalf("deny 输出格式：%s", h.Raw)
	}
	if h := hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/a.go", "x")); h.Raw != "" {
		t.Fatalf("放行时不应有输出：%s", h.Raw)
	}
}

func TestGate_AC_012_5_E1(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantDeny(t, hookRaw(t, dir, "PreToolUse", []byte("{not json")), "gate.input")
	wantDeny(t, hookRaw(t, dir, "PreToolUse", []byte(`{"tool_name":"Write","tool_input":{}}`)), "gate.input", "缺少文件路径")
}

// ---- AC-037-2 代理不能批准产物 ----

func TestGate_AC_037_2_N1(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantAllow(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "正文第一段。", "正文第一段，改写。")))
}

func TestGate_AC_037_2_E1(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "- Status: Draft", "- Status: Approved")), "gate.protect", "harness approve")
}

func TestGate_AC_037_2_E2(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("cd . && harness approve prd")), "gate.protect")
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("go run ./cmd/harness phase next")), "gate.protect")
}
