package app_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// listFiles 返回目录下全部文件的相对路径，按字母排序。
func listFiles(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			rel, _ := filepath.Rel(dir, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	sort.Strings(out)
	return out
}

func initDir(t *testing.T, langs string) string {
	t.Helper()
	dir := t.TempDir()
	if r := runHarness(t, dir, "init", "--lang", langs); r.Code != 0 {
		t.Fatalf("init：%d %s%s", r.Code, r.Stdout, r.Stderr)
	}
	return dir
}

// goInitFiles 是 harness init --lang go 生成的全部文件；内嵌模板缺失时与之不符（AC-N002-2/E1）。
var goInitFiles = []string{
	".claude/settings.json", ".claude/skills/architect/SKILL.md", ".claude/skills/change/SKILL.md",
	".claude/skills/clarify/SKILL.md", ".claude/skills/implement/SKILL.md", ".claude/skills/plan/SKILL.md",
	".claude/skills/test/SKILL.md", ".github/workflows/harness.yml", ".harness/.gitignore", "docs/PRD.md", "docs/architecture/overview.md", "docs/debt.md",
	"docs/lang/go.md", "docs/progress.md", "docs/tasks.md", "harness.json",
}

// ---- AC-001-1 init 后检查通过 ----

func TestInit_AC_001_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, initDir(t, "go")))
}

func TestInit_AC_001_1_B1(t *testing.T) {
	dir := initDir(t, "go,typescript,python,java,cpp")
	wantPass(t, checkJSON(t, dir))
	var cfg struct{ Languages []string }
	b, _ := os.ReadFile(filepath.Join(dir, "harness.json"))
	json.Unmarshal(b, &cfg)
	packs, _ := filepath.Glob(filepath.Join(dir, "docs/lang/*.md"))
	if len(cfg.Languages) != 5 || len(packs) != 5 {
		t.Fatalf("语言配置包：%v %v", cfg.Languages, packs)
	}
}

func TestInit_AC_001_1_E1(t *testing.T) {
	dir := t.TempDir()
	r := runHarness(t, dir, "init", "--lang", "go,rust")
	if r.Code == 0 || !strings.Contains(r.Stderr, "rust") || !strings.Contains(r.Stderr, "cpp, go, java, python, typescript") {
		t.Fatalf("%d %s", r.Code, r.Stderr)
	}
	if files := listFiles(t, dir); len(files) != 0 {
		t.Fatalf("不应留下文件：%v", files)
	}
}

// ---- AC-001-2 重复执行不覆盖 ----

func TestInit_AC_001_2_N1(t *testing.T) {
	dir := initDir(t, "go")
	p := filepath.Join(dir, "docs/PRD.md")
	os.WriteFile(p, []byte("# 我的 PRD\n"), 0o644)
	r := runHarness(t, dir, "init", "--lang", "go")
	if b, _ := os.ReadFile(p); string(b) != "# 我的 PRD\n" || r.Code != 0 || !strings.Contains(r.Stdout, "跳过 docs/PRD.md") {
		t.Fatalf("%d %s 内容：%s", r.Code, r.Stdout, b)
	}
}

func TestInit_AC_001_2_B1(t *testing.T) {
	dir := initDir(t, "go")
	os.Remove(filepath.Join(dir, "docs/tasks.md"))
	r := runHarness(t, dir, "init", "--lang", "go")
	if r.Code != 0 || !strings.HasPrefix(r.Stdout, "写入 docs/tasks.md\n") || !strings.Contains(r.Stdout, "写入 1 个文件，跳过 0 个") {
		t.Fatalf("应只补回被删除的文件：%s", r.Stdout)
	}
}

func TestInit_AC_001_2_E1(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root 可以写入任何目录")
	}
	dir := t.TempDir()
	os.Chmod(dir, 0o555)
	defer os.Chmod(dir, 0o755)
	r := runHarness(t, dir, "init", "--lang", "go")
	if r.Code == 0 || !strings.Contains(r.Stderr, dir) {
		t.Fatalf("%d %s", r.Code, r.Stderr)
	}
}

// ---- AC-N002-2 模板内嵌 ----

func TestInit_AC_N002_2_N1(t *testing.T) {
	// 只有 harness 二进制：复制到隔离目录，清空环境变量运行。
	bin := filepath.Join(t.TempDir(), "harness")
	b, _ := os.ReadFile(harnessBin)
	os.WriteFile(bin, b, 0o755)
	dir := t.TempDir()
	cmd := exec.Command(bin, "init", "--lang", "go")
	cmd.Dir, cmd.Env = dir, []string{"HOME=" + t.TempDir()}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	if got := listFiles(t, dir); strings.Join(got, ",") != strings.Join(goInitFiles, ",") {
		t.Fatalf("生成的文件：%v", got)
	}
}

func TestInit_AC_N002_2_E1(t *testing.T) {
	if got := listFiles(t, initDir(t, "go")); strings.Join(got, ",") != strings.Join(goInitFiles, ",") {
		t.Fatalf("内嵌模板与期望清单不符：%v", got)
	}
}

// ---- AC-N005-1 产物可读可 diff ----

func TestInit_AC_N005_1_N1(t *testing.T) {
	for _, f := range listFiles(t, initDir(t, "go,typescript,python,java,cpp")) {
		if strings.HasPrefix(f, ".harness/") {
			continue // 门禁运行时目录（被忽略、不提交），不是产物
		}
		switch filepath.Ext(f) {
		case ".md", ".yaml", ".yml", ".json": // .yml 同为 YAML；GitHub Actions 与 GitLab CI 的约定文件名用 .yml
		default:
			t.Errorf("%s 不是 Markdown、YAML 或 JSON", f)
		}
	}
}

func TestInit_AC_N005_1_E1(t *testing.T) {
	dir := initDir(t, "go")
	lock := func() []byte {
		if r := runHarness(t, dir, "lock"); r.Code != 0 {
			t.Fatalf("lock：%s", r.Stderr)
		}
		b, _ := os.ReadFile(filepath.Join(dir, "harness.lock"))
		return b
	}
	first, second := lock(), lock()
	if !bytes.Equal(first, second) {
		t.Fatal("两次生成的锁文件不一致")
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(first, &v); err != nil {
		t.Fatal(err)
	}
	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var order []string
	for _, k := range keys {
		order = append(order, `"`+k+`"`)
	}
	pos := -1
	for _, k := range order {
		i := bytes.Index(first, []byte(k))
		if i < pos {
			t.Fatalf("键未按字典序排列：%s", first)
		}
		pos = i
	}
}

// init 写入 .harness/.gitignore：门禁的快照与会话目录、以及这个文件本身都不出现在 git status 中，
// 测试报告与项目根目录的 .gitignore 不受影响（T-033）。
func TestInit_IgnoresGateRuntimeDirs(t *testing.T) {
	dir := initGit(t)
	for _, p := range []string{".harness/snap/s1/state.json", ".harness/session/s1.json", ".harness/reports/go.jsonl"} {
		write(t, dir, p, "x\n")
	}
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain", "--untracked-files=all", "--", ".harness").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(out)); got != "?? .harness/reports/go.jsonl" {
		t.Fatalf("只有测试报告应出现在 git status 中：%q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); !os.IsNotExist(err) {
		t.Fatal("不应创建项目根目录的 .gitignore")
	}
	// 已有项目重新运行 init 补上该文件。
	os.Remove(filepath.Join(dir, ".harness/.gitignore"))
	if r := runHarness(t, dir, "init", "--lang", "go"); r.Code != 0 || !strings.Contains(r.Stdout, "写入 .harness/.gitignore") {
		t.Fatalf("%d %s", r.Code, r.Stdout)
	}
}
