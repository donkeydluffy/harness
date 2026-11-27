package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// networkPackages 是发起网络请求所需的标准库包。
var networkPackages = map[string]bool{
	"net": true, "net/http": true, "net/rpc": true, "net/smtp": true, "crypto/tls": true,
	"net/http/httptrace": true, "net/http/httputil": true,
}

// networkDeps 返回包（不含测试）依赖的网络包，按 go list 输出顺序。
func networkDeps(t *testing.T, pkg string) []string {
	t.Helper()
	cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}}", pkg)
	cmd.Dir = "../.."
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list %s：%v", pkg, err)
	}
	var bad []string
	for _, p := range strings.Fields(string(out)) {
		if networkPackages[p] {
			bad = append(bad, p)
		}
	}
	return bad
}

// ---- AC-N001-1 不发起网络请求 ----

func TestCheck_AC_N001_1_E1(t *testing.T) {
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = "../.."
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range strings.Fields(string(out)) {
		if strings.HasPrefix(pkg, "github.com/donkeydluffy/harness/internal/audit") {
			continue // harness audit deps 允许联网（NFR-001）
		}
		if bad := networkDeps(t, pkg); len(bad) > 0 {
			t.Errorf("%s 引入了网络包：%v", pkg, bad)
		}
	}
	// 对照：引入 net/http 的包被识别出来。
	if bad := networkDeps(t, "./internal/app/testdata/determinism/netpkg"); len(bad) == 0 {
		t.Fatal("没有识别出引入 net/http 的包")
	}
}

func TestCheck_AC_N001_1_N1(t *testing.T) {
	files := traced(t, prd(t, "- AC-001-1: 生成骨架。", "- AC-001-1: 快速生成骨架。"))
	files["docs/tasks.md"] = activeTasks
	dir := writeTree(t, files)
	offline := []string{"HTTP_PROXY=http://127.0.0.1:9", "HTTPS_PROXY=http://127.0.0.1:9", "ALL_PROXY=http://127.0.0.1:9", "NO_PROXY="}
	run := func(env []string, args ...string) string {
		cmd := exec.Command(harnessBin, args...)
		cmd.Dir, cmd.Env = dir, append(os.Environ(), env...)
		out, _ := cmd.CombinedOutput()
		return string(out)
	}
	for _, args := range [][]string{{"check", "--json"}, {"check"}, {"trace", "--format", "json"}} {
		online, off := run(nil, args...), run(offline, args...)
		if online != off || online != run(nil, args...) {
			t.Errorf("harness %s 的结果不一致：\n%s\n----\n%s", strings.Join(args, " "), online, off)
		}
	}
	if !strings.Contains(run(nil, "check"), "prd.") {
		t.Fatal("夹具应含违规，用来比较非空输出")
	}
	// harness lock 两次生成的锁文件逐字节一致。
	edit(t, dir, "- AC-001-1: 快速生成骨架。", "- AC-001-1: 生成骨架。")
	write(t, dir, "docs/tasks.md", taskDoc(task("T-001", "Todo", "FR-001, FR-002", "MOD-01", "internal/api/**", "", "")))
	locks := make([]string, 2)
	for i, env := range [][]string{nil, offline} {
		if out := run(env, "lock"); !strings.Contains(out, "已更新") {
			t.Fatalf("lock：%s", out)
		}
		b, _ := os.ReadFile(filepath.Join(dir, "harness.lock"))
		locks[i] = string(b)
	}
	if locks[0] != locks[1] {
		t.Fatal("两次生成的锁文件不一致")
	}
}
