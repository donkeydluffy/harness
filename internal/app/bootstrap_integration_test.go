package app_test

import "testing"

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
