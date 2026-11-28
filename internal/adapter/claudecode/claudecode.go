// Package claudecode 是 Claude Code 的 hook 适配器。
package claudecode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/donkeydluffy/harness/internal/gate"
)

// Adapter 实现 adapter.Adapter。
type Adapter struct{}

func (Adapter) Name() string { return "claudecode" }

type input struct {
	SessionID string `json:"session_id"`
	Cwd       string `json:"cwd"`
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath     string `json:"file_path"`
		NotebookPath string `json:"notebook_path"`
		Content      string `json:"content"`
		OldString    string `json:"old_string"`
		NewString    string `json:"new_string"`
		ReplaceAll   bool   `json:"replace_all"`
		NewSource    string `json:"new_source"`
		Edits        []struct {
			OldString  string `json:"old_string"`
			NewString  string `json:"new_string"`
			ReplaceAll bool   `json:"replace_all"`
		} `json:"edits"`
		Command string `json:"command"`
	} `json:"tool_input"`
}

// Decode 把 hook 输入转换为门禁请求；输入无法解析时返回错误，调用方据此拒绝（AC-012-5/E1）。
func (Adapter) Decode(event string, stdin []byte) (gate.Request, error) {
	var in input
	if err := json.Unmarshal(stdin, &in); err != nil {
		return gate.Request{}, fmt.Errorf("无法解析 Claude Code hook 输入：%w", err)
	}
	req := gate.Request{SessionID: in.SessionID}
	if req.SessionID == "" {
		req.SessionID = "default"
	}
	cwd := in.Cwd
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	root := projectRoot(cwd)
	req.Root, req.Cwd = root, cwd
	ti := in.ToolInput
	switch event {
	case "SessionStart":
		req.Kind = gate.SessionStart
	case "Stop", "SubagentStop":
		req.Kind = gate.Stop
	case "PreToolUse", "PostToolUse":
		file := ti.FilePath
		if file == "" {
			file = ti.NotebookPath
		}
		switch in.ToolName {
		case "Bash":
			if ti.Command == "" {
				return req, errors.New("Bash 工具输入缺少 command")
			}
			req.Kind, req.Command = gate.Command, ti.Command
		case "Read":
			req.Kind = gate.Read
		case "Write", "Edit", "MultiEdit", "NotebookEdit":
			if file == "" {
				return req, fmt.Errorf("%s 工具输入缺少文件路径", in.ToolName)
			}
			req.Kind = gate.Write
		default:
			req.Kind = gate.Read // 其他工具不写文件
		}
		if file != "" {
			req.Paths = []string{gate.Resolve(root, cwd, file)}
			abs := file
			if !filepath.IsAbs(abs) {
				abs = filepath.Join(cwd, abs)
			}
			req.Content = newContent(in, abs)
		}
		if event == "PostToolUse" {
			req.Kind = gate.PostWrite
		}
	default:
		return req, fmt.Errorf("不支持的 Claude Code hook 事件 %q", event)
	}
	return req, nil
}

// newContent 计算写入后的完整内容；Edit 读取当前文件并应用替换，读取失败时退回新增片段。
func newContent(in input, abs string) string {
	ti := in.ToolInput
	switch in.ToolName {
	case "Write":
		return ti.Content
	case "NotebookEdit":
		return ti.NewSource
	}
	type edit struct {
		old, new string
		all      bool
	}
	edits := []edit{{ti.OldString, ti.NewString, ti.ReplaceAll}}
	if in.ToolName == "MultiEdit" {
		edits = nil
		for _, e := range ti.Edits {
			edits = append(edits, edit{e.OldString, e.NewString, e.ReplaceAll})
		}
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		var parts []string
		for _, e := range edits {
			parts = append(parts, e.new)
		}
		return strings.Join(parts, "\n")
	}
	s := string(b)
	for _, e := range edits {
		if e.all {
			s = strings.ReplaceAll(s, e.old, e.new)
		} else {
			s = strings.Replace(s, e.old, e.new, 1)
		}
	}
	return s
}

// projectRoot 是从 cwd 向上最近的含 harness.json 的目录，其次最近的 git 仓库根，都没有时为 cwd（AC-058-1）。
// 不按 docs/PRD.md 判定：仓库中的脚手架模板也含 docs/PRD.md。
func projectRoot(cwd string) string {
	for _, m := range []string{"harness.json", ".git"} {
		if d, ok := upward(cwd, m); ok {
			return d
		}
	}
	return cwd
}

// upward 从 dir 向上查找含 marker 的目录。
func upward(dir, marker string) (string, bool) {
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, marker)); err == nil {
			return d, true
		}
		if filepath.Dir(d) == d {
			return "", false
		}
	}
}

// Encode 把判定转换为 hook 输出。拒绝通过 JSON 输出表达，退出码为 0。
func (Adapter) Encode(event string, d gate.Decision) ([]byte, int) {
	var out map[string]any
	reason := d.Reason
	if d.Rule != "" {
		reason = "[" + d.Rule + "] " + reason
	}
	switch {
	case event == "PreToolUse" && !d.Allow:
		out = map[string]any{"hookSpecificOutput": map[string]any{
			"hookEventName": "PreToolUse", "permissionDecision": "deny", "permissionDecisionReason": reason}}
	case (event == "Stop" || event == "SubagentStop") && !d.Allow:
		out = map[string]any{"decision": "block", "reason": reason}
	case !d.Allow:
		out = map[string]any{"decision": "block", "reason": reason}
	case d.Feedback != "" && (event == "SessionStart" || event == "PostToolUse"):
		out = map[string]any{"hookSpecificOutput": map[string]any{"hookEventName": event, "additionalContext": d.Feedback}}
	default:
		return nil, 0
	}
	b, _ := json.Marshal(out)
	return append(b, '\n'), 0
}

// HookConfig 返回 .claude/settings.json 的 hooks 配置。harness 内部错误时以退出码 2 退出，
// Claude Code 把退出码 2 视为阻断，保证失败即拒绝。
func (Adapter) HookConfig(binary string) (string, []byte) {
	hook := func(event string) []map[string]any {
		return []map[string]any{{"type": "command", "command": hookCommand(binary, event)}}
	}
	cfg := map[string]any{"hooks": map[string]any{
		"SessionStart": []map[string]any{{"hooks": hook("SessionStart")}},
		"PreToolUse":   []map[string]any{{"matcher": "Read|Edit|Write|MultiEdit|NotebookEdit|Bash", "hooks": hook("PreToolUse")}},
		"PostToolUse":  []map[string]any{{"matcher": "Edit|Write|MultiEdit|NotebookEdit|Bash", "hooks": hook("PostToolUse")}},
		"Stop":         []map[string]any{{"hooks": hook("Stop")}},
	}}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	return ".claude/settings.json", append(b, '\n')
}

// missingBinary 是找不到 harness 时反馈给代理的说明。
const missingBinary = "harness 未安装或不在 Claude Code 的 PATH 中，门禁无法运行，写操作一律拒绝。安装 harness 并把它所在目录加入 PATH 后重启会话。"

// hookCommand 返回事件 event 的 hook 命令：先确认 binary 可执行再交给它判定。
// 找不到时失败即阻断（AC-012-5）：PreToolUse 拒绝，SessionStart 注入说明；
// PostToolUse 与 Stop 不阻止，因为工具调用已被 PreToolUse 拒绝。
func hookCommand(binary, event string) string {
	name := binary
	if f := strings.Fields(binary); len(f) > 0 {
		name = f[0]
	}
	var fallback string
	switch event {
	case "PreToolUse":
		fallback = fallbackJSON(map[string]any{"hookEventName": event, "permissionDecision": "deny", "permissionDecisionReason": missingBinary})
	case "SessionStart":
		fallback = fallbackJSON(map[string]any{"hookEventName": event, "additionalContext": missingBinary})
	}
	check := "command -v " + name + " >/dev/null 2>&1 || "
	if fallback == "" {
		check += "exit 0; "
	} else {
		check += "{ printf '%s\\n' '" + fallback + "'; exit 0; }; "
	}
	return check + "exec " + binary + " gate claudecode " + event
}

// fallbackJSON 序列化 hookSpecificOutput；内容不含单引号，可直接放进 sh 的单引号字符串。
func fallbackJSON(out map[string]any) string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.Encode(map[string]any{"hookSpecificOutput": out})
	return strings.TrimSuffix(b.String(), "\n")
}
