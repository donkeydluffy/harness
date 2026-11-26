package gate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"harness/internal/phase"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// artifactRules 按写入的产物文件选择检查的规则前缀（FLOW-04）。
func artifactRules(p string) []string {
	switch {
	case p == workspace.PRDIndex || strings.HasPrefix(p, "docs/prd/"):
		return []string{"prd", "term"}
	case strings.HasPrefix(p, "docs/architecture/"):
		return []string{"arch", "trace", "term"}
	case strings.HasPrefix(p, "docs/adr/"):
		return []string{"adr"}
	case p == "docs/tasks.md":
		return []string{"task"}
	case p == "docs/debt.md":
		return []string{"debt"}
	}
	return nil
}

// postWrite 在代理写入产物文件后运行对应的文档规则，结果作为反馈（AC-012-3）。
// 依赖测试报告、源码或版本库的规则不在写后运行，留给会话结束检查。
func postWrite(s *workspace.Snapshot, req Request) Decision {
	var prefixes []string
	for _, p := range req.Paths {
		prefixes = append(prefixes, artifactRules(p)...)
	}
	if len(prefixes) == 0 {
		return allow()
	}
	var sel engine.Selector
	for _, id := range engine.IDs() {
		r, _ := engine.Lookup(id)
		if engine.Needs(r) != 0 {
			continue
		}
		for _, pre := range prefixes {
			if strings.HasPrefix(id, pre+".") {
				sel.IDs = append(sel.IDs, id)
				break
			}
		}
	}
	rep := engine.Run(s, sel)
	what := "harness check " + strings.Join(prefixes, " ")
	if len(rep.Violations) == 0 {
		return Decision{Allow: true, Feedback: what + "：通过"}
	}
	return Decision{Allow: true, Feedback: fmt.Sprintf("%s：%d 条违规\n%s", what, len(rep.Violations), formatViolations(rep.Violations))}
}

func formatViolations(vs []engine.Violation) string {
	var b strings.Builder
	for _, v := range vs {
		fmt.Fprintf(&b, "%s:%d [%s] %s\n", v.File, v.Line, v.Rule, v.Msg)
	}
	return strings.TrimRight(b.String(), "\n")
}

type sessionState struct {
	Blocks int `json:"blocks"`
}

func sessionFile(root, id string) string {
	return filepath.Join(root, ".harness", "session", filepath.Base(id)+".json")
}

// stop 在会话结束前运行全量检查（AC-012-4）。未通过时阻止结束；同一会话连续阻止 stopMaxBlocks 次后放行，
// 未通过项写入进度文件的"未完成"段，下次会话开始时优先注入。检查通过时计数清零、清除"未完成"段。
func stop(s *workspace.Snapshot, req Request) Decision {
	file := sessionFile(s.Root, req.SessionID)
	var st sessionState
	if b, err := os.ReadFile(file); err == nil {
		json.Unmarshal(b, &st)
	}
	rep := engine.Run(s, engine.FullCheck())
	save := func() error {
		os.MkdirAll(filepath.Dir(file), 0o755)
		b, _ := json.Marshal(st)
		return os.WriteFile(file, b, 0o644)
	}
	if len(rep.Violations) == 0 && len(rep.Errors) == 0 {
		st.Blocks = 0
		if err := save(); err != nil {
			return deny("gate.error", "无法写入会话状态：%v", err)
		}
		if err := writeUnfinished(s.Root, ""); err != nil {
			return deny("gate.error", "无法更新 %s：%v", progressFile, err)
		}
		return allow()
	}
	list := formatViolations(rep.Violations)
	for _, e := range rep.Errors {
		list += "\n" + e.Error()
	}
	list = strings.TrimLeft(list, "\n")
	if st.Blocks >= s.Config.StopMaxBlocks {
		st.Blocks = 0
		if err := save(); err != nil {
			return deny("gate.error", "无法写入会话状态：%v", err)
		}
		text := fmt.Sprintf("%s 会话 %s 结束时未通过的检查（已连续阻止 %d 次后放行；相关任务不能标为 Done）：\n\n%s\n",
			time.Now().Format("2006-01-02"), req.SessionID, s.Config.StopMaxBlocks, list)
		if err := writeUnfinished(s.Root, text); err != nil {
			return deny("gate.error", "无法更新 %s：%v", progressFile, err)
		}
		return Decision{Allow: true, Feedback: "未通过项已写入 " + progressFile + " 的\"未完成\"段"}
	}
	st.Blocks++
	if err := save(); err != nil {
		return deny("gate.error", "无法写入会话状态：%v", err)
	}
	return deny("gate.stop", "全量检查有 %d 条未通过（第 %d 次阻止，连续 %d 次后放行）：\n%s",
		len(rep.Violations)+len(rep.Errors), st.Blocks, s.Config.StopMaxBlocks, list)
}

const (
	progressFile    = "docs/progress.md"
	unfinishedTitle = "## 未完成"
)

// writeUnfinished 用 text 替换进度文件的"未完成"段；text 为空时删除该段。
func writeUnfinished(root, text string) error {
	p := filepath.Join(root, filepath.FromSlash(progressFile))
	b, err := os.ReadFile(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	rest, _ := splitUnfinished(string(b))
	if text == "" && len(b) == 0 {
		return nil
	}
	out := strings.TrimRight(rest, "\n")
	if text != "" {
		out += "\n\n" + unfinishedTitle + "\n\n" + text
	}
	out = strings.TrimLeft(out, "\n")
	if !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	if out == string(b) {
		return nil
	}
	os.MkdirAll(filepath.Dir(p), 0o755)
	return os.WriteFile(p, []byte(out), 0o644)
}

// splitUnfinished 把进度文件分为"未完成"段以外的内容与该段正文。
func splitUnfinished(s string) (rest, section string) {
	lines := strings.Split(s, "\n")
	var keep, sec []string
	in := false
	for _, l := range lines {
		switch {
		case strings.TrimSpace(l) == unfinishedTitle:
			in = true
			continue
		case in && strings.HasPrefix(l, "## "):
			in = false
		}
		if in {
			sec = append(sec, l)
		} else {
			keep = append(keep, l)
		}
	}
	return strings.Join(keep, "\n"), strings.TrimSpace(strings.Join(sec, "\n"))
}

var phaseTitles = map[string]string{"requirements": "需求", "architecture": "架构", "planning": "规划", "implementation": "实现"}

// sessionStart 向代理注入项目状态（AC-024-1）与冒烟测试结果（AC-024-3）；上次会话的未完成项最先注入。
func sessionStart(s *workspace.Snapshot) Decision {
	var parts []string
	if d := s.Doc(progressFile); d != nil {
		if _, sec := splitUnfinished(strings.Join(d.Lines, "\n")); sec != "" {
			parts = append(parts, "上次会话遗留：\n"+sec)
		}
	}
	name := "未知"
	if phases, err := phase.Load(s); err == nil {
		if cur, err := phase.Current(s, phases); err == nil {
			name = cur.Name
			if t, ok := phaseTitles[name]; ok {
				name = t
			}
		}
	}
	var active []string
	for _, t := range tasks(s) {
		if t.Fields["Status"] == "InProgress" {
			active = append(active, t.ID)
		}
	}
	line := "阶段：" + name + "，无进行中任务"
	if len(active) > 0 {
		line = "阶段：" + name + "，进行中任务：" + strings.Join(active, "、")
	}
	parts = append(parts, line)
	rep := engine.Run(s, engine.FullCheck())
	if n := len(rep.Violations) + len(rep.Errors); n > 0 {
		parts = append(parts, fmt.Sprintf("未通过的检查：%d 条（运行 harness check 查看）", n))
	}
	var oqs []string
	for _, oq := range s.Project.ByKind["OQ"] {
		if oq.Fields["Status"] == "Open" {
			oqs = append(oqs, oq.ID)
		}
	}
	if len(oqs) > 0 {
		parts = append(parts, "Open 待澄清项："+strings.Join(oqs, "、"))
	}
	if smoke := runSmoke(s.Root, s.Config.Smoke); smoke != "" {
		parts = append(parts, smoke)
	}
	return Decision{Allow: true, Feedback: strings.Join(parts, "\n")}
}

// smokeTimeout 是冒烟命令的最长运行时间。
const smokeTimeout = 5 * time.Minute

// runSmoke 运行项目声明的冒烟命令；未声明时返回空串。失败时返回输出的最后 50 行。
func runSmoke(root, cmd string) string {
	if strings.TrimSpace(cmd) == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), smokeTimeout)
	defer cancel()
	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	c.Dir = root
	out, err := c.CombinedOutput()
	if err == nil {
		return "冒烟通过"
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) > 50 {
		lines = lines[len(lines)-50:]
	}
	return fmt.Sprintf("冒烟失败（%v），输出最后 %d 行：\n%s", err, len(lines), strings.Join(lines, "\n"))
}
