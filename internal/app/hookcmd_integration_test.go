package app_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/adapter/claudecode"
)

// hookCommands 返回 harness init 生成的 Claude Code hook 配置中各事件的命令。
func hookCommands(t *testing.T) map[string]string {
	t.Helper()
	_, b := claudecode.Adapter{}.HookConfig("harness")
	var cfg struct {
		Hooks map[string][]struct {
			Hooks []struct{ Command string }
		}
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		t.Fatal(err)
	}
	out := map[string]string{}
	for ev, groups := range cfg.Hooks {
		out[ev] = groups[0].Hooks[0].Command
	}
	return out
}

// runHookCommand 像 Claude Code 一样用 sh 执行 hook 命令，PATH 为 path。
func runHookCommand(t *testing.T, dir, command, path string, in map[string]any) (string, int) {
	t.Helper()
	in["session_id"], in["cwd"] = "s1", dir
	stdin, _ := json.Marshal(in)
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir, cmd.Stdin = dir, bytes.NewReader(stdin)
	cmd.Env = append(os.Environ(), "PATH="+path)
	var out bytes.Buffer
	cmd.Stdout = &out
	code := 0
	if err := cmd.Run(); err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatal(err)
		}
		code = ee.ExitCode()
	}
	return out.String(), code
}

// 找不到 harness 时 hook 失败即阻断（AC-012-5，T-030）：PreToolUse 拒绝，SessionStart 注入说明，
// PostToolUse 与 Stop 不阻止；找得到时交给 harness 判定。
func TestGate_MissingBinaryFailsClosed(t *testing.T) {
	dir := gateDir(t, true, activeTasks)
	cmds := hookCommands(t)
	noHarness := "/usr/bin:/bin"

	out, code := runHookCommand(t, dir, cmds["PreToolUse"], noHarness, writeTool(dir, "internal/api/a.go", "package api\n"))
	var o struct {
		Hook struct {
			HookEventName, PermissionDecision, PermissionDecisionReason, AdditionalContext string
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(out), &o); err != nil || code != 0 || o.Hook.PermissionDecision != "deny" ||
		o.Hook.HookEventName != "PreToolUse" || !strings.Contains(o.Hook.PermissionDecisionReason, "harness 未安装") {
		t.Fatalf("找不到 harness 时 PreToolUse 应拒绝：%d %v %q", code, err, out)
	}

	out, code = runHookCommand(t, dir, cmds["SessionStart"], noHarness, map[string]any{})
	o.Hook.AdditionalContext = ""
	if err := json.Unmarshal([]byte(out), &o); err != nil || code != 0 || !strings.Contains(o.Hook.AdditionalContext, "harness 未安装") {
		t.Fatalf("找不到 harness 时 SessionStart 应注入说明：%d %v %q", code, err, out)
	}

	for _, ev := range []string{"PostToolUse", "Stop"} {
		if out, code := runHookCommand(t, dir, cmds[ev], noHarness, map[string]any{}); out != "" || code != 0 {
			t.Fatalf("找不到 harness 时 %s 不应阻止：%d %q", ev, code, out)
		}
	}

	// 找得到时由 harness 判定：范围外的写入被拒绝，规则编号来自门禁。
	withHarness := filepath.Dir(harnessBin) + ":" + noHarness
	out, _ = runHookCommand(t, dir, cmds["PreToolUse"], withHarness, writeTool(dir, "other/b.go", "package other\n"))
	if !strings.Contains(out, "gate.write") {
		t.Fatalf("找得到 harness 时应由门禁判定：%q", out)
	}
	out, _ = runHookCommand(t, dir, cmds["PreToolUse"], withHarness, writeTool(dir, "internal/api/new.go", "package api\n"))
	if strings.Contains(out, "deny") {
		t.Fatalf("范围内的写入应放行：%q", out)
	}
}
