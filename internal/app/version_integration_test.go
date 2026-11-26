package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const harnessPkg = "github.com/donkeydluffy/harness/cmd/harness"

// buildIn 在 dir 中构建 harness，返回二进制路径。
func buildIn(t *testing.T, dir string, extra ...string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "harness")
	args := append([]string{"build", "-o", bin}, extra...)
	cmd := exec.Command("go", append(args, harnessPkg)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("构建失败：%v\n%s", err, out)
	}
	return bin
}

func versionOf(t *testing.T, bin string, args ...string) result {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var r result
	out, err := cmd.Output()
	r.Stdout = string(out)
	if ee, ok := err.(*exec.ExitError); ok {
		r.Code, r.Stderr = ee.ExitCode(), string(ee.Stderr)
	} else if err != nil {
		t.Fatal(err)
	}
	return r
}

// 发布构建注入的版本号原样输出，去掉前缀 v。
func TestVersion_AC_049_1_N1(t *testing.T) {
	bin := buildIn(t, repoRoot, "-ldflags", "-X github.com/donkeydluffy/harness/internal/version.injected=v0.1.4")
	for _, arg := range []string{"version", "--version"} {
		if r := versionOf(t, bin, arg); r.Code != 0 || r.Stdout != "harness 0.1.4\n" {
			t.Errorf("harness %s：退出码 %d，输出 %q", arg, r.Code, r.Stdout)
		}
	}
}

// 在仓库中从源码构建：输出包含当前提交哈希的前 12 位。
func TestVersion_AC_049_1_N2(t *testing.T) {
	root := repoRoot
	head, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	r := runHarness(t, t.TempDir(), "version")
	if r.Code != 0 || !strings.HasPrefix(r.Stdout, "harness ") || !strings.Contains(r.Stdout, string(head[:12])) {
		t.Errorf("退出码 %d，输出 %q，应包含 %s", r.Code, r.Stdout, head[:12])
	}
}

// 构建时工作区有未提交改动：输出以 -dirty 结尾。
func TestVersion_AC_049_1_B1(t *testing.T) {
	root := repoRoot
	files, err := exec.Command("git", "-C", root, "ls-files", "-co", "--exclude-standard").Output()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	for _, rel := range strings.Fields(string(files)) {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue // 已删除但未暂存的文件
		}
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v：%v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("add", "-A")
	git("commit", "-q", "--no-verify", "-m", "snapshot")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("改动\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := versionOf(t, buildIn(t, dir), "version")
	if r.Code != 0 || !strings.HasSuffix(strings.TrimSpace(r.Stdout), "-dirty") {
		t.Errorf("退出码 %d，输出 %q，应以 -dirty 结尾", r.Code, r.Stdout)
	}
}

// 带多余参数：退出码非 0 并提示用法。
func TestVersion_AC_049_1_E1(t *testing.T) {
	r := runHarness(t, t.TempDir(), "version", "extra")
	if r.Code == 0 || !strings.Contains(r.Stderr, "用法") {
		t.Errorf("退出码 %d，错误输出 %q", r.Code, r.Stderr)
	}
}

func TestVersion_AC_049_2_N1(t *testing.T) {
	r := runHarness(t, t.TempDir(), "help")
	if r.Code != 0 || !strings.Contains(r.Stdout, "version") {
		t.Errorf("help 应列出 version：%q", r.Stdout)
	}
}

func TestVersion_AC_049_2_E1(t *testing.T) {
	r := runHarness(t, t.TempDir(), "no-such-command")
	if r.Code == 0 || !strings.Contains(r.Stderr, "version") {
		t.Errorf("未知命令的用法提示应列出 version：%q", r.Stderr)
	}
}
