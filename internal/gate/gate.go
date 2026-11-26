// Package gate 是与代理无关的门禁判定（MOD-11）：按判定顺序决定放行或拒绝代理的读写与命令。
package gate

import (
	"fmt"
	"path"
	"strings"

	"github.com/donkeydluffy/harness/internal/approval"
	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/phase"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Kind 是门禁请求的类型。
type Kind string

const (
	Read         Kind = "read"
	Write        Kind = "write"
	Command      Kind = "command"
	SessionStart Kind = "session_start"
	PostWrite    Kind = "post_write"
	Stop         Kind = "stop"
)

// Request 是门禁请求（IF-017）。
type Request struct {
	Kind      Kind
	Root      string   // 项目根（绝对路径），由适配器从代理执行工具时的目录向上查找
	Cwd       string   // 代理执行工具时的目录（绝对路径）；命令中的相对路径从这里开始解析，为空时从项目根开始
	Paths     []string // 已解析为仓库相对路径；仓库外的路径以 "../" 开头
	Content   string   // write 时写入后的完整内容（无法得到时为新增片段）
	Command   string   // command 时的原始命令
	SessionID string
}

// Decision 是门禁判定（IF-017）。
type Decision struct {
	Allow    bool
	Rule     string // 拒绝时的门禁规则编号，如 gate.write
	Reason   string
	Feedback string // 反馈给代理的文字
}

func allow() Decision { return Decision{Allow: true} }

func deny(rule, format string, a ...any) Decision {
	return Decision{Rule: rule, Reason: fmt.Sprintf(format, a...)}
}

// Decide 按判定顺序给出结论（IF-017）：gate.protect → gate.sensitive → gate.secret → gate.command
// （命令中的写入目标再按前述顺序与 gate.artifact、gate.write 判定）→ gate.artifact → gate.write。
//
// harness:IF-017
func Decide(s *workspace.Snapshot, req Request) Decision {
	switch req.Kind {
	case Read:
		for _, p := range req.Paths {
			if sensitive(s, p) {
				return denySensitive(p)
			}
		}
		return allow()
	case Write:
		for _, p := range req.Paths {
			if d := decideWrite(s, p, req.Content); !d.Allow {
				return d
			}
		}
		return allow()
	case Command:
		return decideCommand(s, req.Command, req.Cwd)
	case PostWrite:
		return postWrite(s, req)
	case Stop:
		return stop(s, req)
	case SessionStart:
		return sessionStart(s)
	default:
		return allow()
	}
}

// protectedFiles 只能由人或 harness 命令修改。
var protectedFiles = []string{"harness.lock", workspace.ConfigFile, ".claude/settings.json"}

// protectedDirs 下是门禁自身的状态（会话计数、命令前快照）。
var protectedDirs = []string{".harness/session/", ".harness/snap/"}

// decideWrite 判定对一个路径的写入；content 为写入后的内容，未知时为空串。
func decideWrite(s *workspace.Snapshot, p, content string) Decision {
	switch {
	case strings.HasPrefix(p, "../") || p == "..":
		return allow() // 仓库外的文件不受本项目流程约束
	case contains(protectedFiles, p) || hasPrefix(p, protectedDirs):
		return deny("gate.protect", "%s 只能由人或 harness 命令修改（锁文件、项目配置、hook 配置、门禁状态）", p)
	case workspace.Approvable(p) && content != "" && approvesStatus(s, p, content):
		return deny("gate.protect", "代理不能把 %s 的状态改为 %s；批准须由人在终端执行 harness approve", p, workspace.ApprovedStatus(p))
	case sensitive(s, p):
		return denySensitive(p)
	}
	if content != "" {
		if id, desc := findSecret(addedText(s.Root, p, content)); id != "" {
			return denySecret("写入 "+p+" 的内容", id, desc)
		}
	}
	switch {
	case strings.HasPrefix(p, "../") || p == "..":
		return allow()
	case strings.HasPrefix(p, ".harness/"):
		return allow() // 测试报告等运行产物
	case strings.HasPrefix(p, "docs/"):
		return decideArtifact(s, p)
	default:
		return decideSource(s, p)
	}
}

// approvesStatus 判断写入后文件头状态变为批准状态，而写入前不是。
func approvesStatus(s *workspace.Snapshot, p, content string) bool {
	lines := strings.Split(content, "\n")
	i := workspace.HeaderStatusLine(lines)
	want := workspace.ApprovedStatus(p)
	if i < 0 || strings.TrimSpace(strings.TrimPrefix(lines[i], "- Status:")) != want {
		return false
	}
	d := s.Doc(p)
	return d == nil || workspace.HeaderStatus(d) != want
}

// decideArtifact：docs/ 下的写入须在当前阶段的可写路径内（gate.artifact）。
// 锁文件为第 2 版时 docs/ 任何时候可写（AC-048-2），批准状态由批准链体现。
func decideArtifact(s *workspace.Snapshot, p string) Decision {
	if s.Lock.ByUnit() {
		return allow()
	}
	phases, err := phase.Load(s)
	if err != nil {
		return deny("gate.artifact", "阶段定义有误，无法判定 %s 是否可写：%v", p, err)
	}
	cur, err := phase.Current(s, phases)
	if err != nil {
		return deny("gate.artifact", "无法确定当前阶段：%v", err)
	}
	if workspace.MatchAny(cur.Writable, p) {
		return allow()
	}
	return deny("gate.artifact", "当前阶段 %s 不允许写入 %s；可写路径：%s", cur.Name, p, strings.Join(cur.Writable, ", "))
}

// decideSource：源码写入须有 InProgress 任务且路径在其范围内（gate.write）；锁文件为第 2 版时
// 该任务的批准链还须全部有效（AC-047-4）。
func decideSource(s *workspace.Snapshot, p string) Decision {
	var active []string
	var broken *Decision
	for _, t := range tasks(s) {
		if t.Fields["Status"] != "InProgress" {
			continue
		}
		active = append(active, t.ID)
		if !workspace.MatchAny(t.List("Scope"), p) {
			continue
		}
		if !s.Lock.ByUnit() {
			return allow()
		}
		bad := brokenChain(s, t.ID)
		if len(bad) == 0 {
			return allow()
		}
		if broken == nil {
			d := deny("gate.write", "%s 的批准链无效，不能修改其范围内的源码 %s：%s；由人在终端执行 harness approve 重新批准后再继续",
				t.ID, p, strings.Join(bad, "；"))
			broken = &d
		}
	}
	if broken != nil {
		return *broken
	}
	if len(active) == 0 {
		return deny("gate.write", "没有 InProgress 任务，不能修改源码 %s；先在 docs/tasks.md 中把要做的任务改为 InProgress", p)
	}
	var scopes []string
	for _, t := range tasks(s) {
		if t.Fields["Status"] == "InProgress" {
			scopes = append(scopes, t.ID+"："+strings.Join(t.List("Scope"), ", "))
		}
	}
	return deny("gate.write", "%s 不在 InProgress 任务的文件范围内（%s）", p, strings.Join(scopes, "；"))
}

func tasks(s *workspace.Snapshot) []*model.Entity {
	var out []*model.Entity
	if d := s.Doc("docs/tasks.md"); d != nil {
		for _, e := range d.Entities {
			if e.Kind == "T" {
				out = append(out, e)
			}
		}
	}
	return out
}

func contains(list []string, x string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

func hasPrefix(p string, prefixes []string) bool {
	for _, pre := range prefixes {
		if strings.HasPrefix(p, pre) {
			return true
		}
	}
	return false
}

// Clean 规范化仓库相对路径。
func Clean(p string) string { return path.Clean(strings.TrimPrefix(p, "./")) }

// AfterAllowed 返回命令后比对时允许保留的改动：整条命令只调用一次 harness（如 harness lock、harness init）时，
// 允许它改写锁文件、项目配置与 hook 配置；其余路径按 Allowed 判定。
func AfterAllowed(s *workspace.Snapshot, cmd string) func(string) bool {
	harness := onlyHarness(cmd)
	removed := regenRemovals(s, cmd)
	return func(p string) bool {
		return (harness && (contains(protectedFiles, p) || contains(workspace.RegeneratedFiles, p))) || removed[p] || Allowed(s, p)
	}
}

// Allowed 判断代理对路径 p 的写入是否放行；命令后比对（IF-018 After）用它决定哪些改动保留。
func Allowed(s *workspace.Snapshot, p string) bool { return decideWrite(s, p, "").Allow }

// brokenChain 返回任务批准链中未批准与失效的单位及原因。
func brokenChain(s *workspace.Snapshot, id string) []string {
	var out []string
	for _, st := range approval.Chain(s, id) {
		switch st.State {
		case approval.Unapproved:
			out = append(out, st.Unit.ID+" 未批准")
		case approval.Stale:
			out = append(out, st.Unit.ID+" 批准已失效（"+strings.Join(st.Cause, ", ")+" 已变化）")
		}
	}
	return out
}
