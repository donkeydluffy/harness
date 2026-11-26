package builtin

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// 每条内置规则在 testdata/<规则编号>/ 下必须有 violation* 与 ok* 样本目录，并通过 Verify。
func TestRulesSamples_IF_014(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range Rules() {
		if seen[r.ID()] {
			t.Fatalf("规则编号重复：%s", r.ID())
		}
		seen[r.ID()] = true
		if len(r.Covers()) == 0 {
			t.Errorf("%s 没有声明覆盖的验收标准", r.ID())
		}
		dirs, _ := filepath.Glob(filepath.Join("testdata", r.ID(), "*"))
		var samples []engine.Sample
		for _, d := range dirs {
			if fi, err := os.Stat(d); err != nil || !fi.IsDir() {
				continue
			}
			root := d
			if fi, err := os.Stat(filepath.Join(d, "repo")); err == nil && fi.IsDir() {
				root = gitSample(t, filepath.Join(d, "repo"))
			}
			opt := workspace.LoadOptions{WithSources: true, WithTests: true}
			if b, err := os.ReadFile(filepath.Join(d, "range")); err == nil {
				opt.CommitRange = strings.TrimSpace(string(b))
			}
			s, err := workspace.Load(root, opt)
			if err != nil {
				t.Fatalf("%s: %v", d, err)
			}
			name := filepath.Base(d)
			samples = append(samples, engine.Sample{Name: name, Snapshot: s, WantViolation: !strings.HasPrefix(name, "ok")})
		}
		if err := engine.Verify(r, samples); err != nil {
			t.Error(err)
		}
	}
}

// gitSample 把 repo/01、repo/02…依次作为完整目录树提交（步骤目录中的 MESSAGE 文件为提交信息），
// repo/staged 存在时最后暂存，返回临时仓库目录。样本目录中的 range 文件指定提交范围。
func gitSample(t *testing.T, repo string) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	steps, _ := filepath.Glob(filepath.Join(repo, "*"))
	sort.Strings(steps) // 01、02…在 staged 之前
	for _, step := range steps {
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if e.Name() != ".git" {
				os.RemoveAll(filepath.Join(dir, e.Name()))
			}
		}
		if err := os.CopyFS(dir, os.DirFS(step)); err != nil {
			t.Fatal(err)
		}
		msg := "sample " + filepath.Base(step)
		if b, err := os.ReadFile(filepath.Join(dir, "MESSAGE")); err == nil {
			msg = string(b)
			os.Remove(filepath.Join(dir, "MESSAGE"))
		}
		run("add", "-A")
		if filepath.Base(step) != "staged" {
			run("commit", "-q", "--allow-empty", "-m", msg)
		}
	}
	return dir
}
