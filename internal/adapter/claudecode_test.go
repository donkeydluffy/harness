package adapter

import (
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/gate"
)

func TestClaudeCode_IF_019(t *testing.T) {
	a, ok := Get("claudecode")
	if !ok || a.Name() != "claudecode" {
		t.Fatal("应内置 claudecode 适配器")
	}
	dir := t.TempDir()
	req, err := a.Decode("PreToolUse", []byte(`{"session_id":"s1","cwd":"`+dir+`","tool_name":"Bash","tool_input":{"command":"ls"}}`))
	if err != nil || req.Kind != gate.Command || req.Command != "ls" || req.SessionID != "s1" {
		t.Fatalf("解码 Bash：%v %+v", err, req)
	}
	out, code := a.Encode("PreToolUse", gate.Decision{Rule: "gate.write", Reason: "不在范围内"})
	if code != 0 || !strings.Contains(string(out), `"deny"`) || !strings.Contains(string(out), "不在范围内") {
		t.Fatalf("编码拒绝：%d %s", code, out)
	}
	if p, c := a.HookConfig("harness"); p != ".claude/settings.json" || !strings.Contains(string(c), "harness gate claudecode") {
		t.Fatalf("hook 配置：%s %s", p, c)
	}
}
