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
		sent := prompt == ""
		for {
			c, err := r.ReadByte()
			if err != nil {
				close(done)
				return
			}
			b.WriteByte(c)
			if !sent && strings.Contains(b.String(), prompt) {
				io.WriteString(in, answer+"\n")
				sent = true
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
