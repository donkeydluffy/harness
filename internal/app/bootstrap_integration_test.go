package app_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/adapter/claudecode"
)

const repoRoot = "../.."

// selfCheckRules 是对本仓库根目录运行的检查：除依赖测试报告（报告正由本次 go test 重写）与暂存区的规则外的全部规则。
var selfCheckRules = []string{"prd", "term", "arch.01", "arch.02", "arch.03", "arch.04", "trace", "adr", "debt",
	"task.01", "task.02", "task.03", "task.04", "task.05", "task.06", "task.07", "approve", "phase", "cfg", "ci"}

// repoFiles 读取本仓库中 dir 下的全部文件（仓库相对路径 → 内容）。
func repoFiles(t *testing.T, dirs ...string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, d := range dirs {
		filepath.WalkDir(filepath.Join(repoRoot, d), func(p string, e fs.DirEntry, err error) error {
			if err != nil || e.IsDir() {
				return err
			}
			b, err := os.ReadFile(p)
			rel, _ := filepath.Rel(repoRoot, p)
			out[filepath.ToSlash(rel)] = string(b)
			return err
		})
	}
	return out
}

// ---- AC-017-1 本仓库通过检查 ----

func TestCheck_AC_017_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, repoRoot, selfCheckRules...))
}

func TestCheck_AC_017_1_E1(t *testing.T) {
	files := repoFiles(t, "docs")
	files["docs/PRD.md"] = strings.Replace(files["docs/PRD.md"], "## 4. 非功能需求",
		"### FR-099 未被设计覆盖的需求\n- Status: Active\n- Release: 1\n- Priority: P2\n\n- AC-099-1: 新增。\n  - AC-099-1/N1: 正常\n  - AC-099-1/E1: 异常\n\n## 4. 非功能需求", 1)
	wantViolation(t, checkJSON(t, writeTree(t, files), "trace"), "trace.01", "FR-099 未被任何设计元素覆盖")
}

// ---- AC-017-2 本仓库启用 hook ----

func TestCheck_AC_017_2_N1(t *testing.T) {
	got, err := os.ReadFile(filepath.Join(repoRoot, ".claude/settings.json"))
	_, want := claudecode.Adapter{}.HookConfig("harness")
	if err != nil || string(got) != string(want) {
		t.Fatalf("本仓库的 hook 配置应为 harness init 生成的内容：%v\n%s", err, got)
	}
	for _, ev := range []string{"SessionStart", "PreToolUse", "PostToolUse", "Stop"} {
		if !strings.Contains(string(got), "harness gate claudecode "+ev) {
			t.Errorf("缺少 %s 的 hook", ev)
		}
	}
}

func TestCheck_AC_017_2_E1(t *testing.T) {
	files := repoFiles(t, ".claude", ".github")
	files["harness.json"] = repoFiles(t, "harness.json")["harness.json"]
	files[".claude/settings.json"] = strings.Replace(files[".claude/settings.json"], "harness gate claudecode Stop", "true", 1)
	wantViolation(t, checkJSON(t, writeTree(t, files), "cfg.02"), "cfg.02", ".claude/settings.json 与重新生成的内容不一致")
}

// harness init 写入的 skill 与 GitLab CI 配置是生成文件（TERM-026），提交范围与行数检查豁免。
func TestCheck_GeneratedSkillsExempt(t *testing.T) {
	files := map[string]string{
		".claude/skills/plan/SKILL.md": lines(500),
		".gitlab-ci.yml":               "harness:\n  script:\n    - harness check\n",
		"internal/core/core.go":        "package core\n\nvar X = 1\n",
	}
	c := checkLast(t, commitRepo(t, files, "改动\n\nTask: T-003"))
	wantPass(t, c)
	// 对照：不在生成目录下的同样文件计入范围与行数检查。
	files[".claude/x/SKILL.md"] = lines(500)
	c = checkLast(t, commitRepo(t, files, "改动\n\nTask: T-003"))
	wantViolation(t, c, "commit.02", ".claude/x/SKILL.md 不在 T-003 的文件范围内")
	wantViolation(t, c, "commit.03", "改动 502 行")
}
