package app_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initGit 在 git 仓库中执行 harness init --lang go，返回目录。
func initGit(t *testing.T) string {
	t.Helper()
	dir := gitRepo(t, map[string]string{"README.md": "x\n"})
	if r := runHarness(t, dir, "init", "--lang", "go"); r.Code != 0 {
		t.Fatalf("init：%s%s", r.Stdout, r.Stderr)
	}
	return dir
}

// ---- AC-013-3 git hook 与 CI 兜底 ----

func TestInit_AC_013_3_N1(t *testing.T) {
	dir := initGit(t)
	for _, h := range []string{"commit-msg", "pre-commit"} {
		p := filepath.Join(dir, ".git/hooks", h)
		fi, err := os.Stat(p)
		if err != nil || fi.Mode()&0o111 == 0 {
			t.Fatalf("%s 不存在或不可执行：%v", h, err)
		}
		if b, _ := os.ReadFile(p); !strings.Contains(string(b), "harness check") {
			t.Fatalf("%s 没有调用 harness：%s", h, b)
		}
	}
	// commit-msg hook 实际调用 harness：提交信息缺少 Task 尾注时失败。
	os.MkdirAll(filepath.Join(dir, "internal"), 0o755)
	os.WriteFile(filepath.Join(dir, "internal/a.go"), []byte("package a\n"), 0o644)
	git(t, dir, "add", "internal/a.go")
	msg := filepath.Join(t.TempDir(), "MSG")
	os.WriteFile(msg, []byte("改动\n"), 0o644)
	cmd := exec.Command(filepath.Join(dir, ".git/hooks/commit-msg"), msg)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PATH="+filepath.Dir(harnessBin)+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "commit.01") {
		t.Fatalf("commit-msg hook 应拒绝没有 Task 尾注的提交：%v %s", err, out)
	}
}

func TestCheck_AC_013_3_N2(t *testing.T) {
	dir := gitRepo(t, map[string]string{"docs/architecture/overview.md": commitArch, "docs/tasks.md": commitTasks, "internal/core/core.go": "package core\n"})
	stage(t, dir, coreChange)
	git(t, dir, "commit", "-q", "-m", "第一个\n\nTask: T-99")
	stage(t, dir, map[string]string{"internal/core/b.go": "package core\n"})
	git(t, dir, "commit", "-q", "-m", "第二个，没有尾注")
	c := checkJSON(t, dir, "commits", "HEAD~2..HEAD")
	wantViolation(t, c, "commit.01", "T-99 不存在")
	wantViolation(t, c, "commit.01", "没有 Task 尾注")
}

func TestCheck_AC_013_3_E1(t *testing.T) {
	dir := initGit(t)
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-q", "--no-verify", "-m", "绕过 hook 的提交")
	wantViolation(t, checkJSON(t, dir, "commits", "HEAD~1..HEAD"), "commit.01", "没有 Task 尾注")
}

func TestInit_AC_013_3_B1(t *testing.T) {
	dir := t.TempDir()
	r := runHarness(t, dir, "init", "--lang", "go")
	if r.Code != 0 || !strings.Contains(r.Stdout, "跳过 git hook 安装") {
		t.Fatalf("%d %s", r.Code, r.Stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
		t.Fatal("非 git 仓库中不应创建 .git")
	}
}

// ---- AC-043-1 CI 模板 ----

func TestInit_AC_043_1_N1(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(initDir(t, "go"), ".github/workflows/harness.yml"))
	if err != nil || !strings.Contains(string(b), "run: harness check\n") || !strings.Contains(string(b), "harness check commits") {
		t.Fatalf("%v\n%s", err, b)
	}
}

func TestInit_AC_043_1_N2(t *testing.T) {
	dir := t.TempDir()
	if r := runHarness(t, dir, "init", "--lang", "go", "--ci", "gitlab"); r.Code != 0 {
		t.Fatal(r.Stderr)
	}
	b, err := os.ReadFile(filepath.Join(dir, ".gitlab-ci.yml"))
	if err != nil || !strings.Contains(string(b), "- harness check\n") || !strings.Contains(string(b), "harness check commits") {
		t.Fatalf("%v\n%s", err, b)
	}
	if _, err := os.Stat(filepath.Join(dir, ".github")); !os.IsNotExist(err) {
		t.Fatal("选 gitlab 时不应生成 GitHub Actions 配置")
	}
	wantPass(t, checkJSON(t, dir, "cfg.02"))
}

func TestInit_AC_043_1_E1(t *testing.T) {
	dir := t.TempDir()
	r := runHarness(t, dir, "init", "--lang", "go", "--ci", "jenkins")
	if r.Code == 0 || !strings.Contains(r.Stderr, "github, gitlab") || len(listFiles(t, dir)) != 0 {
		t.Fatalf("%d %s", r.Code, r.Stderr)
	}
}

// ---- AC-041-2 生成文件一致性 ----

func TestCheck_AC_041_2_N1(t *testing.T) {
	wantPass(t, checkJSON(t, initDir(t, "go"), "cfg.02"))
}

func TestCheck_AC_041_2_E1(t *testing.T) {
	dir := initDir(t, "go")
	p := filepath.Join(dir, ".github/workflows/harness.yml")
	b, _ := os.ReadFile(p)
	os.WriteFile(p, []byte(strings.Replace(string(b), "run: harness check\n", "run: echo skip\n", 1)), 0o644)
	v := wantViolation(t, checkJSON(t, dir, "cfg.02"), "cfg.02", ".github/workflows/harness.yml", "harness init --lang go")
	if lines := strings.Split(string(b), "\n"); v.Line < 1 || strings.TrimSpace(lines[v.Line-1]) != "run: harness check" {
		t.Fatalf("行号 %d 应指向被改动的行", v.Line)
	}
}

func TestCheck_AC_041_2_B1(t *testing.T) {
	dir := initDir(t, "go")
	os.Remove(filepath.Join(dir, ".claude/settings.json"))
	wantViolation(t, checkJSON(t, dir, "cfg.02"), "cfg.02", "缺少生成文件 .claude/settings.json", "harness init --lang go")
}

// CI 配置在 harness check 之前按语言配置包运行测试、写出测试报告；没有通用测试命令的语言留注释（T-027）。
func TestInit_CIRunsTests(t *testing.T) {
	gh, err := os.ReadFile(filepath.Join(initDir(t, "go,python"), ".github/workflows/harness.yml"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(gh)
	step := "go test -json ./... > .harness/reports/go.jsonl"
	if i, j := strings.Index(s, step), strings.Index(s, "run: harness check\n"); i < 0 || j < i {
		t.Fatalf("GitHub Actions 应先运行 Go 测试再运行 harness check：\n%s", s)
	}
	if !strings.Contains(s, "# python：没有通用的测试命令") || strings.Contains(s, "测试（python）") {
		t.Fatalf("Python 没有通用测试命令，应只留注释：\n%s", s)
	}
	dir := t.TempDir()
	if r := runHarness(t, dir, "init", "--lang", "go", "--ci", "gitlab"); r.Code != 0 {
		t.Fatal(r.Stderr)
	}
	gl, _ := os.ReadFile(filepath.Join(dir, ".gitlab-ci.yml"))
	if i, j := strings.Index(string(gl), "    - mkdir -p .harness/reports && "+step), strings.Index(string(gl), "    - harness check\n"); i < 0 || j < i {
		t.Fatalf("GitLab CI 应先运行 Go 测试再运行 harness check：\n%s", gl)
	}
}
