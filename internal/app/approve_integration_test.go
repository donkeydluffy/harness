package app_test

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// inTerminal 在伪终端中运行 harness（借助系统 script 命令），看到 prompt 后输入 answer。
// 返回全部输出与退出码。没有 script 命令的平台跳过。
func inTerminal(t *testing.T, dir, prompt, answer string, args ...string) (string, int) {
	t.Helper()
	return inTerminalSeq(t, dir, prompt, []string{answer}, args...)
}

// inTerminalSeq 与 inTerminal 相同，但第 i 次出现 prompt 时回答 answers[i]。
func inTerminalSeq(t *testing.T, dir, prompt string, answers []string, args ...string) (string, int) {
	t.Helper()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin", "freebsd":
		cmd = exec.Command("script", append([]string{"-q", "/dev/null", harnessBin}, args...)...)
	case "linux":
		cmd = exec.Command("script", "-qec", harnessBin+" "+strings.Join(args, " "), "/dev/null")
	default:
		t.Skip("需要 script 命令提供伪终端")
	}
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("需要 script 命令提供伪终端")
	}
	cmd.Dir = dir
	in, _ := cmd.StdinPipe()
	out, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	done := make(chan struct{})
	go func() {
		r := bufio.NewReader(out)
		sent := 0
		if prompt == "" {
			sent = len(answers)
		}
		for {
			c, err := r.ReadByte()
			if err != nil {
				close(done)
				return
			}
			b.WriteByte(c)
			if sent < len(answers) && strings.Count(b.String(), prompt) > sent {
				io.WriteString(in, answers[sent]+"\n")
				sent++
			}
		}
	}()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("超时，输出：%s", b.String())
	}
	err := cmd.Wait()
	in.Close()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	}
	return b.String(), code
}

const confirmPrompt = "输入 y 确认："

func approvals(t *testing.T, dir string) []map[string]string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, "harness.lock"))
	if err != nil {
		return nil
	}
	var l struct{ Approvals []map[string]string }
	json.Unmarshal(b, &l)
	return l.Approvals
}

func readFile(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// approvedRepo 返回 PRD 已由人批准的仓库。
func approvedRepo(t *testing.T) string {
	t.Helper()
	dir := lockedRepo(t, basePRD)
	out, code := inTerminal(t, dir, confirmPrompt, "y", "approve", "prd")
	if code != 0 {
		t.Fatalf("批准失败：%d %s", code, out)
	}
	return dir
}

// ---- AC-037-1 只能由人在终端批准 ----

func TestApprove_AC_037_1_N1(t *testing.T) {
	dir := approvedRepo(t)
	if w := approvalWarnings(t, dir); strings.Contains(w, "docs/PRD.md") || strings.Contains(w, "FR-001 未批准") {
		t.Fatalf("PRD 应处于已批准状态：\n%s", w)
	}
	u := unitRecords(t, dir)["FR-001"]
	if u == nil || !strings.HasPrefix(u["hash"].(string), "sha256:") || u["by"] == "" || u["at"] == "" {
		t.Fatalf("锁文件应新增批准记录：%v", u)
	}
}

func TestApprove_AC_037_1_E1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	r := runHarness(t, dir, "approve", "prd") // 标准输入不是终端
	if r.Code == 0 || !strings.Contains(r.Stderr, "交互式终端") {
		t.Fatalf("非交互环境应拒绝：%d %s", r.Code, r.Stderr)
	}
	if len(approvals(t, dir)) != 0 || strings.Contains(readFile(t, dir, "docs/PRD.md"), "Status: Approved") {
		t.Fatal("拒绝后不应写入批准")
	}
}

func TestApprove_AC_037_1_B1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	prdBefore, lockBefore := readFile(t, dir, "docs/PRD.md"), readFile(t, dir, "harness.lock")
	out, code := inTerminal(t, dir, confirmPrompt, "n", "approve", "prd")
	if code == 0 || !strings.Contains(out, "未确认") {
		t.Fatalf("输入 n 应不批准：%d %s", code, out)
	}
	if readFile(t, dir, "docs/PRD.md") != prdBefore || readFile(t, dir, "harness.lock") != lockBefore {
		t.Fatal("输入 n 后 PRD 或锁文件被修改")
	}
}

func TestApprove_AC_037_1_E2(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	edit(t, dir, "- Type: Constraint\n", "")
	out, code := inTerminal(t, dir, "", "", "approve", "prd")
	if code != 1 || !strings.Contains(out, "prd.02") || !strings.Contains(out, "拒绝批准") {
		t.Fatalf("有检查失败项时应拒绝：%d %s", code, out)
	}
	if len(approvals(t, dir)) != 0 {
		t.Fatal("拒绝后不应写入批准")
	}
}

// ---- AC-056-1 批准等待正在执行的代理命令 ----

const agentWaiting = "代理命令正在执行"

// agentRunning 在 dir 中留下会话 s1 的命令前快照，表示代理命令正在执行；age 为快照的时长。
func agentRunning(t *testing.T, dir string, age time.Duration) string {
	t.Helper()
	snap := filepath.Join(dir, ".harness", "snap", "s1")
	if err := os.MkdirAll(snap, 0o755); err != nil {
		t.Fatal(err)
	}
	at := time.Now().Add(-age)
	os.Chtimes(snap, at, at)
	return snap
}

func TestApprove_AC_056_1_N1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	out, code := inTerminal(t, dir, confirmPrompt, "y", "approve", "prd")
	if code != 0 || strings.Contains(out, agentWaiting) {
		t.Fatalf("没有正在执行的代理命令时不应等待：%d %s", code, out)
	}
	if unitRecords(t, dir)["FR-001"] == nil {
		t.Fatal("应写入批准记录")
	}
}

func TestApprove_AC_056_1_E1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	snap := agentRunning(t, dir, 0)
	go func() { time.Sleep(1500 * time.Millisecond); os.RemoveAll(snap) }() // 代理命令结束，After 清除快照
	out, code := inTerminal(t, dir, confirmPrompt, "y", "approve", "prd")
	if code != 0 || !strings.Contains(out, agentWaiting) {
		t.Fatalf("代理命令执行期间应提示等待，结束后照常批准：%d %s", code, out)
	}
	if i, j := strings.Index(out, agentWaiting), strings.Index(out, "新增 FR-001"); j < i {
		t.Fatalf("应在命令结束后才列出待批准单位：%s", out)
	}
	if unitRecords(t, dir)["FR-001"] == nil {
		t.Fatal("命令结束后应写入批准记录")
	}
}

func TestApprove_AC_056_1_B1(t *testing.T) {
	dir := lockedRepo(t, basePRD)
	agentRunning(t, dir, 11*time.Minute)
	lockBefore := readFile(t, dir, "harness.lock")
	out, code := inTerminal(t, dir, confirmPrompt, "n", "approve", "prd")
	if code == 0 || !strings.Contains(out, "异常退出") || strings.Contains(out, agentWaiting) {
		t.Fatalf("残留快照应询问人，输入 n 退出：%d %s", code, out)
	}
	if readFile(t, dir, "harness.lock") != lockBefore {
		t.Fatal("输入 n 后锁文件不应改变")
	}
	out, code = inTerminalSeq(t, dir, confirmPrompt, []string{"y", "y"}, "approve", "prd")
	if code != 0 || !strings.Contains(out, "异常退出") || unitRecords(t, dir)["FR-001"] == nil {
		t.Fatalf("输入 y 后应照常批准：%d %s", code, out)
	}
}
