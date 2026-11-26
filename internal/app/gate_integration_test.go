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

const gatePRD = "# PRD\n\n- Status: Draft\n- Current-Release: 1\n\n正文第一段。\n"

// gateDir 返回处于 phase 阶段（空串为需求阶段）的项目；tasks 为任务列表内容（空串时不写）。
func gateDir(t *testing.T, phase, tasks string) string {
	t.Helper()
	files := map[string]string{"docs/PRD.md": gatePRD, "internal/api/a.go": "package api\n", "other/b.go": "package other\n"}
	if tasks != "" {
		files["docs/tasks.md"] = tasks
	}
	if phase != "" {
		files["harness.lock"] = `{"phase": "` + phase + `"}` + "\n"
	}
	return writeTree(t, files)
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
	dir := gateDir(t, "implementation", activeTasks)
	wantAllow(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/a.go", "package api\n\nvar X = 1\n")))
}

func TestGate_AC_012_1_E1(t *testing.T) {
	dir := gateDir(t, "implementation", taskDoc(task("T-001", "Todo", "FR-001", "MOD-01", "internal/api/**", "", "")))
	wantDeny(t, hook(t, dir, "PreToolUse", editTool(dir, "internal/api/a.go", "package api", "package api // x")), "gate.write", "没有 InProgress 任务")
}

func TestGate_AC_012_1_E2(t *testing.T) {
	dir := gateDir(t, "implementation", "")
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("echo x > src/a.go")), "gate.write", "src/a.go")
}

func TestGate_AC_012_1_B1(t *testing.T) {
	dir := gateDir(t, "", "")
	wantAllow(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "正文第一段。", "正文第一段，补充。")))
}

// ---- AC-012-2 写入须在任务范围内 ----

func TestGate_AC_012_2_N1(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	wantAllow(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/new.go", "package api\n")))
}

func TestGate_AC_012_2_E1(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", writeTool(dir, "other/b.go", "package other\n")), "gate.write", "other/b.go", "T-001：internal/api/**")
}

func TestGate_AC_012_2_B1(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/../../other/b.go", "x")), "other/b.go")
	if err := os.Symlink(filepath.Join(dir, "other"), filepath.Join(dir, "internal/api/link")); err != nil {
		t.Fatal(err)
	}
	wantDeny(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/link/c.go", "x")), "other/c.go")
}

func TestGate_AC_012_2_E2(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("sed -i '' 's/a/b/' other/b.go")), "gate.write", "other/b.go")
}

// 命令无法静态判定时放行执行，事后比对恢复范围外改动（FLOW-03）。
func TestGate_FLOW_03(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
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
	dir := gateDir(t, "implementation", activeTasks)
	h := hook(t, dir, "PreToolUse", writeTool(dir, "other/b.go", "x"))
	if !strings.Contains(h.Raw, `"hookEventName":"PreToolUse"`) || !strings.Contains(h.Raw, `"permissionDecision":"deny"`) {
		t.Fatalf("deny 输出格式：%s", h.Raw)
	}
	if h := hook(t, dir, "PreToolUse", writeTool(dir, "internal/api/a.go", "x")); h.Raw != "" {
		t.Fatalf("放行时不应有输出：%s", h.Raw)
	}
}

func TestGate_AC_012_5_E1(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	wantDeny(t, hookRaw(t, dir, "PreToolUse", []byte("{not json")), "gate.input")
	wantDeny(t, hookRaw(t, dir, "PreToolUse", []byte(`{"tool_name":"Write","tool_input":{}}`)), "gate.input", "缺少文件路径")
}

// ---- AC-037-2 代理不能批准产物 ----

func TestGate_AC_037_2_N1(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	wantAllow(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "正文第一段。", "正文第一段，改写。")))
}

func TestGate_AC_037_2_E1(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "- Status: Draft", "- Status: Approved")), "gate.protect", "harness approve")
}

func TestGate_AC_037_2_E2(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("cd . && harness approve prd")), "gate.protect")
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("go run ./cmd/harness phase next")), "gate.protect")
}

// ---- AC-045-3 代理不能修改当前阶段 ----

func TestGate_AC_045_3_E3(t *testing.T) {
	dir := gateDir(t, "requirements", "")
	wantDeny(t, hook(t, dir, "PreToolUse", editTool(dir, "harness.lock", "requirements", "implementation")), "gate.protect", "harness.lock")
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool(`echo '{"phase":"implementation"}' > harness.lock`)), "gate.protect")
}

// ---- AC-045-4 产物写入按阶段判定 ----

func TestGate_AC_045_4_N1(t *testing.T) {
	dir := gateDir(t, "planning", "")
	wantAllow(t, hook(t, dir, "PreToolUse", writeTool(dir, "docs/tasks.md", "# 任务列表\n")))
}

func TestGate_AC_045_4_E1(t *testing.T) {
	dir := gateDir(t, "", "")
	wantDeny(t, hook(t, dir, "PreToolUse", writeTool(dir, "docs/tasks.md", "# 任务列表\n")), "gate.artifact", "requirements", "docs/PRD.md, docs/prd/**")
}

func TestGate_AC_045_4_B1(t *testing.T) {
	dir := gateDir(t, "implementation", "")
	wantAllow(t, hook(t, dir, "PreToolUse", editTool(dir, "docs/PRD.md", "正文第一段。", "正文第一段，需求变更。")))
}

// 只含变量赋值的命令不使门禁崩溃；赋值里命令替换的写入照常判定（T-028）。
func TestGate_AssignmentOnlyCommand(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	h := hook(t, dir, "PreToolUse", bashTool("d=$(mktemp -d)"))
	if h.Code != 0 {
		t.Fatalf("门禁不应崩溃：%d %s", h.Code, h.Raw)
	}
	wantAllow(t, h)
	wantAllow(t, hook(t, dir, "PreToolUse", bashTool("X=1 Y=2")))
	wantDeny(t, hook(t, dir, "PreToolUse", bashTool("d=$(touch other/b.go)")), "gate.write", "other/b.go")
}

// 代理可按 cfg.02 的修复办法删除可重新生成的生成文件，再单独运行 harness init 写回（T-031）。
func TestGate_RegenerateGeneratedFile(t *testing.T) {
	dir := gateDir(t, "implementation", activeTasks)
	write(t, dir, ".github/workflows/harness.yml", "旧内容\n")
	write(t, dir, ".claude/settings.json", "{}\n")

	// 只删除生成文件：命令前放行，命令后保留删除。
	rm := bashTool("rm -f .github/workflows/harness.yml")
	wantAllow(t, hook(t, dir, "PreToolUse", rm))
	os.Remove(filepath.Join(dir, ".github/workflows/harness.yml"))
	os.Remove(filepath.Join(dir, ".claude/settings.json"))
	hook(t, dir, "PostToolUse", rm)
	if _, err := os.Stat(filepath.Join(dir, ".github/workflows/harness.yml")); !os.IsNotExist(err) {
		t.Fatal("CI 配置的删除被恢复")
	}

	// 单独一次 harness init 写回：命令后保留写回的内容。
	initCmd := bashTool("harness init --lang go")
	wantAllow(t, hook(t, dir, "PreToolUse", initCmd))
	write(t, dir, ".github/workflows/harness.yml", "新内容\n")
	write(t, dir, "other/b.go", "被 init 改动\n")
	hook(t, dir, "PostToolUse", initCmd)
	if b, _ := os.ReadFile(filepath.Join(dir, ".github/workflows/harness.yml")); string(b) != "新内容\n" {
		t.Fatalf("harness init 写回的 CI 配置被恢复：%q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "other/b.go")); string(b) != "package other\n" {
		t.Fatalf("范围外的其他改动应恢复：%q", b)
	}

	// 其余形状仍按原规则判定。
	for _, c := range []string{
		"rm .claude/settings.json",                    // hook 配置只能由人修改（AC-025-4）
		"rm harness.json",                             // 项目配置删除后会丢失
		"rm harness.lock",                             // 批准记录
		"rm .claude/settings.json && echo x",          // 复合命令
		"rm -r .claude/settings.json",                 // 其他选项
		"rm .github/workflows/harness.yml other/b.go", // 混入范围外文件
	} {
		if h := hook(t, dir, "PreToolUse", bashTool(c)); !h.Deny {
			t.Errorf("%q 应拒绝：%s", c, h.Raw)
		}
	}
	// 通配符无法静态判定：命令前放行，命令后恢复被删除的生成文件（FLOW-03）。
	glob := bashTool("rm .claude/*.json")
	wantAllow(t, hook(t, dir, "PreToolUse", glob))
	os.Remove(filepath.Join(dir, ".claude/settings.json"))
	hook(t, dir, "PostToolUse", glob)
	if _, err := os.Stat(filepath.Join(dir, ".claude/settings.json")); err != nil {
		t.Fatal("通配符删除的生成文件应恢复")
	}
}
