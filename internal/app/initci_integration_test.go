package app_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initOK(t *testing.T, dir string, args ...string) {
	t.Helper()
	if r := runHarness(t, dir, append([]string{"init"}, args...)...); r.Code != 0 {
		t.Fatalf("harness init 失败：%d %s%s", r.Code, r.Stdout, r.Stderr)
	}
}

// before 判断 a 在 s 中出现在 b 之前。
func before(s, a, b string) bool {
	i, j := strings.Index(s, a), strings.Index(s, b)
	return i >= 0 && j >= 0 && i < j
}

// ---- AC-014-1 语言配置包 ----

func TestInit_AC_014_1_N1(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "go,python", "--ci-version", "0.2.0")
	var cfg struct{ Tools map[string]map[string]string }
	json.Unmarshal([]byte(readFile(t, dir, "harness.json")), &cfg)
	if cfg.Tools["go"]["format"] != "gofmt -l ." || cfg.Tools["go"]["lint"] != "go vet ./..." ||
		!strings.HasPrefix(cfg.Tools["python"]["format"], "ruff format") || !strings.HasPrefix(cfg.Tools["python"]["lint"], "ruff check") {
		t.Fatalf("项目配置应包含两种语言的格式化与静态检查命令：%+v", cfg.Tools)
	}
}

func TestInit_AC_014_1_E1(t *testing.T) {
	dir := t.TempDir()
	r := runHarness(t, dir, "init", "--lang", "rust", "--ci-version", "0.2.0")
	if r.Code == 0 || !strings.Contains(r.Stderr, "go") || !strings.Contains(r.Stderr, "python") {
		t.Fatalf("不支持的语言应失败并列出支持的语言：%d %s", r.Code, r.Stderr)
	}
}

// ---- AC-014-6 工具配置文件 ----

func TestInit_AC_014_6_N1(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "python", "--ci-version", "0.2.0")
	if !strings.Contains(readFile(t, dir, "ruff.toml"), "[lint]") {
		t.Fatal("应生成 ruff.toml")
	}
}

func TestInit_AC_014_6_B1(t *testing.T) {
	dir := writeTree(t, map[string]string{"ruff.toml": "line-length = 79\n"})
	initOK(t, dir, "--lang", "python", "--ci-version", "0.2.0")
	if got := readFile(t, dir, "ruff.toml"); got != "line-length = 79\n" {
		t.Fatalf("已有的 ruff.toml 不应改动：%q", got)
	}
}

func TestInit_AC_014_6_B2(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "python", "--ci-version", "0.2.0")
	editFile(t, dir, "ruff.toml", "line-length = 100", "line-length = 120")
	wantPass(t, checkJSON(t, dir, "cfg.02"))
}

// ---- AC-014-7 CI 运行 lint ----

func TestInit_AC_014_7_N1(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "go", "--ci-version", "0.2.0")
	if ci := readFile(t, dir, ".github/workflows/harness.yml"); !before(ci, "run: harness check lint\n", "run: harness check\n") {
		t.Fatalf("lint 步骤应位于 harness check 之前：\n%s", ci)
	}
}

func TestInit_AC_014_7_B1(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "go", "--ci", "gitlab", "--ci-version", "0.2.0")
	if ci := readFile(t, dir, ".gitlab-ci.yml"); !before(ci, "- harness check lint\n", "- harness check\n") {
		t.Fatalf("GitLab CI 应包含 lint 步骤：\n%s", ci)
	}
}

// ---- AC-055-2 harness check smoke ----

func TestCheck_AC_055_2_N1(t *testing.T) {
	dir := writeTree(t, map[string]string{"harness.json": `{"smoke": "true"}`})
	if r := runHarness(t, dir, "check", "smoke"); r.Code != 0 {
		t.Fatalf("冒烟测试退出码 0 时应通过：%d %s", r.Code, r.Stdout)
	}
}

func TestCheck_AC_055_2_E1(t *testing.T) {
	dir := writeTree(t, map[string]string{"harness.json": `{"smoke": "for i in $(seq 1 60); do echo line$i; done; exit 1"}`})
	r := runHarness(t, dir, "check", "smoke")
	if r.Code != 1 || !strings.Contains(r.Stdout, "line60") || !strings.Contains(r.Stdout, "line11\n") || strings.Contains(r.Stdout, "line10\n") {
		t.Fatalf("应失败并输出最后 50 行：%d %s", r.Code, r.Stdout)
	}
}

func TestInit_AC_055_2_B1(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "go", "--ci-version", "0.2.0")
	if ci := readFile(t, dir, ".github/workflows/harness.yml"); !before(ci, "run: harness check smoke\n", "run: harness check\n") {
		t.Fatalf("冒烟测试步骤应位于 harness check 之前：\n%s", ci)
	}
}

// ---- AC-051-1 CI 下载固定版本并校验 ----

func TestInit_AC_051_1_N1(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "python", "--ci-version", "0.2.0")
	ci := readFile(t, dir, ".github/workflows/harness.yml")
	for _, want := range []string{"releases/download/v0.2.0/harness_0.2.0_linux_amd64.tar.gz", "releases/download/v0.2.0/checksums.txt", "sha256sum -c"} {
		if !strings.Contains(ci, want) {
			t.Errorf("应包含 %q：\n%s", want, ci)
		}
	}
	if strings.Contains(ci, "setup-go") || strings.Contains(ci, "go install") {
		t.Errorf("非 Go 项目的 CI 不需要 Go 工具链：\n%s", ci)
	}
}

func TestInit_AC_051_1_N2(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "python", "--ci", "gitlab", "--ci-version", "0.2.0")
	ci := readFile(t, dir, ".gitlab-ci.yml")
	if !strings.Contains(ci, "harness_0.2.0_linux_amd64.tar.gz") || !strings.Contains(ci, "sha256sum -c") || strings.Contains(ci, "golang") {
		t.Fatalf("GitLab CI 应固定为 0.2.0 并校验：\n%s", ci)
	}
}

// installScript 取出 GitHub Actions 配置中安装 harness 步骤的脚本。
func installScript(t *testing.T, ci string) string {
	t.Helper()
	i := strings.Index(ci, "- name: 安装 harness")
	j := strings.Index(ci[i:], "      - name: 测试")
	if i < 0 || j < 0 {
		t.Fatalf("找不到安装步骤：\n%s", ci)
	}
	var lines []string
	for _, l := range strings.Split(ci[i:i+j], "\n")[2:] {
		lines = append(lines, strings.TrimSpace(l))
	}
	return strings.Join(lines, "\n")
}

func TestInit_AC_051_1_E1(t *testing.T) {
	if _, err := exec.LookPath("sha256sum"); err != nil {
		t.Skip("没有 sha256sum")
	}
	dir := t.TempDir()
	initOK(t, dir, "--lang", "go", "--ci-version", "0.2.0")
	script := installScript(t, readFile(t, dir, ".github/workflows/harness.yml"))
	// 假 curl：checksums.txt 记的是另一份内容的哈希，下载的归档与之不符。
	bin, work := t.TempDir(), t.TempDir()
	fake := "#!/bin/sh\nfor a; do u=$a; done\nf=$(basename \"$u\")\n" +
		"if [ \"$f\" = checksums.txt ]; then printf 'good' | sha256sum | sed 's/-$/harness_0.2.0_linux_amd64.tar.gz/' > \"$f\"; else printf 'bad' > \"$f\"; fi\n"
	os.WriteFile(filepath.Join(bin, "curl"), []byte(fake), 0o755)
	cmd := exec.Command("bash", "-eo", "pipefail", "-c", script+"\ntouch checks-ran")
	cmd.Dir, cmd.Env = work, append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("校验不符时安装步骤应失败：\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(work, "checks-ran")); err == nil {
		t.Fatal("安装失败后不应继续")
	}
	if _, err := os.Stat(filepath.Join(work, "harness")); err == nil {
		t.Fatal("校验失败时不应解出 harness")
	}
}

// ---- AC-051-2 没有发布版本号时拒绝 ----

func TestInit_AC_051_2_N1(t *testing.T) {
	dir := t.TempDir()
	initOK(t, dir, "--lang", "go", "--ci-version", "0.2.0")
	if !strings.Contains(readFile(t, dir, ".github/workflows/harness.yml"), "安装 harness 0.2.0") {
		t.Fatal("CI 应固定为 0.2.0")
	}
}

func TestInit_AC_051_2_E1(t *testing.T) {
	dir := t.TempDir()
	r := runHarness(t, dir, "init", "--lang", "go")
	if r.Code == 0 || !strings.Contains(r.Stderr, "--ci-version") {
		t.Fatalf("从源码构建的 harness 应拒绝并提示 --ci-version：%d %s", r.Code, r.Stderr)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Fatalf("不应写入任何文件：%v", entries)
	}
}

func TestInit_AC_051_2_B1(t *testing.T) {
	dir := writeTree(t, map[string]string{"go.mod": "module github.com/donkeydluffy/harness\n"})
	initOK(t, dir, "--lang", "go")
	if ci := readFile(t, dir, ".github/workflows/harness.yml"); !strings.Contains(ci, "run: go install ./cmd/harness") || strings.Contains(ci, "releases/download") {
		t.Fatalf("harness 自身仓库应从源码构建：\n%s", ci)
	}
}

// ---- AC-051-3 cfg.02 按固定版本比对 ----

// versioned 构建注入版本号 v 的 harness。
func versioned(t *testing.T, v string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "harness")
	cmd := exec.Command("go", "build", "-ldflags", "-X github.com/donkeydluffy/harness/internal/version.injected="+v, "-o", bin, "./cmd/harness")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("构建 %s：%v\n%s", v, err, out)
	}
	return bin
}

func cfg02With(t *testing.T, bin, dir string) (string, int) {
	t.Helper()
	cmd := exec.Command(bin, "check", "cfg.02")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	return string(out), 0
}

func pinnedRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	initOK(t, dir, "--lang", "go", "--ci-version", "0.2.0")
	return dir
}

func TestCheck_AC_051_3_N1(t *testing.T) {
	if out, code := cfg02With(t, versioned(t, "0.2.0"), pinnedRepo(t)); code != 0 {
		t.Fatalf("本地与 CI 都是 0.2.0 时应通过：%d %s", code, out)
	}
}

func TestCheck_AC_051_3_E1(t *testing.T) {
	out, code := cfg02With(t, versioned(t, "0.3.0"), pinnedRepo(t))
	if code != 1 || !strings.Contains(out, "CI 固定 0.2.0，本地 0.3.0") || !strings.Contains(out, "harness init") {
		t.Fatalf("应报告两个版本并提示重新生成：%d %s", code, out)
	}
}

func TestCheck_AC_051_3_B1(t *testing.T) {
	wantPass(t, checkJSON(t, pinnedRepo(t), "cfg.02"))
}
