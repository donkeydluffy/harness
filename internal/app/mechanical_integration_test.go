package app_test

import (
	"fmt"
	"strings"
	"testing"
)

const mechADR = "# ADR-001 模块路径改名\n\n- Status: Accepted\n- Relates: FR-001\n\n## 背景\n\n仓库迁到 GitHub。\n\n## 决策\n\n模块路径改为 github.com/donkeydluffy/harness。\n\n## 后果\n\n全部 import 随之改动。\n"

const mechRewrite = "harness -> github.com/donkeydluffy/harness"

// mechTask 返回机械改动任务 T-003；extra 追加在字段末尾。
func mechTask(status, scope, extra string) string {
	return "\n### T-003 模块改名\n- Status: " + status + "\n- Kind: mechanical\n- Covers: FR-001, FR-002\n- Design:\n- Scope: " + scope + "\n- Depends:\n" + extra
}

// checkMechTasks 在架构夹具上写入 ADR-001 与任务列表，运行任务规则与 task.09。
func checkMechTasks(t *testing.T, overview string, tasks ...string) checked {
	t.Helper()
	files := arch(t)
	if overview != "" {
		files["docs/architecture/overview.md"] = overview
	}
	files["docs/PRD.md"] = basePRD
	files["docs/adr/ADR-001-rename.md"] = mechADR
	files["docs/tasks.md"] = taskDoc(tasks...)
	return checkJSON(t, writeTree(t, files), append(append([]string{}, taskRules...), "task.09")...)
}

// ---- AC-050-1 机械改动任务的声明 ----

func TestCheck_AC_050_1_N1(t *testing.T) {
	wantPass(t, checkMechTasks(t, "", mechTask("Todo", "internal/api/**", "- Rewrite: "+mechRewrite+"\n- Relates: ADR-001\n")))
}

func TestCheck_AC_050_1_E1(t *testing.T) {
	c := checkMechTasks(t, "", mechTask("Todo", "internal/api/**", "- Relates: ADR-001\n"))
	wantViolation(t, c, "task.09", "T-003", "没有 Rewrite")
}

func TestCheck_AC_050_1_E2(t *testing.T) {
	c := checkMechTasks(t, "", mechTask("Todo", "internal/api/**", "- Rewrite: "+mechRewrite+"\n- Relates: FR-001\n"))
	wantViolation(t, c, "task.09", "T-003", "没有引用", "ADR")
}

// ---- AC-050-2 普通任务的约束不变，机械改动任务免于这些约束 ----

func TestCheck_AC_050_2_N1(t *testing.T) {
	wantPass(t, checkTasks(t, basePRD, taskDoc(t001, t002)))
}

func TestCheck_AC_050_2_E1(t *testing.T) {
	t1 := task("T-001", "Todo", "FR-001", "MOD-01", "internal/api/**, internal/gate/x.go", "", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t1, t002)), "task.03", "T-001", "internal/gate/x.go")
}

func TestCheck_AC_050_2_E2(t *testing.T) {
	t2 := task("T-002", "InProgress", "FR-002", "MOD-01, MOD-02", "internal/store/**, internal/api/x.go", "T-001", "")
	wantViolation(t, checkTasks(t, basePRD, taskDoc(t001, t2)), "task.05", "T-002", "跨 2 个模块", "没有引用调用流程")
}

func TestCheck_AC_050_2_E3(t *testing.T) {
	c := checkLast(t, commitRepo(t, map[string]string{"internal/core/big.go": lines(401)}, "大\n\nTask: T-003"))
	wantViolation(t, c, "commit.03", "401 行", "上限 400")
}

func TestCheck_AC_050_2_B1(t *testing.T) {
	var b strings.Builder
	b.WriteString("# 架构总览\n\n- Status: Draft\n")
	var scope []string
	for i := 1; i <= 14; i++ {
		fmt.Fprintf(&b, "\n### MOD-%02d m%02d\n- Status: Active\n- Paths: internal/m%02d/**\n- Public: internal/m%02d/m.go\n- Depends:\n- Interfaces:\n- Covers: FR-001\n- Core: false\n- Coverage: 0\n", i, i, i, i)
		scope = append(scope, fmt.Sprintf("internal/m%02d/**", i))
	}
	c := checkMechTasks(t, b.String(), mechTask("Todo", strings.Join(scope, ", "), "- Rewrite: "+mechRewrite+"\n- Relates: ADR-001\n"))
	wantPass(t, c)
}

// mechRepo 建立模块路径为 harness 的 Go 仓库，机械改动任务 T-003 为 InProgress；files 追加在初始提交中。
func mechRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	base := map[string]string{
		"harness.json":               `{"languages": ["go"], "sourcePaths": ["internal/**"]}`,
		"go.mod":                     "module harness\n\ngo 1.21\n",
		"docs/adr/ADR-001-rename.md": mechADR,
		"docs/tasks.md":              taskDoc(mechTask("InProgress", "**", "- Rewrite: "+mechRewrite+"\n- Relates: ADR-001\n")),
		"internal/b/b.go":            "package b\n\nfunc B() string { return \"b\" }\n",
	}
	for k, v := range files {
		base[k] = v
	}
	return gitRepo(t, base)
}

const mechMsg = "refactor: 模块路径改为 github.com/donkeydluffy/harness\n\nTask: T-003"

// importer 返回 import 了 paths 的 Go 文件，body 为函数体之外的代码。
func importer(pkg, body string, paths ...string) string {
	return "package " + pkg + "\n\nimport (\n\t\"fmt\"\n\n\t\"" + strings.Join(paths, "\"\n\t\"") + "\"\n)\n\nfunc F() { fmt.Println() }\n" + body
}

// renamed 把文件中的 "harness/ 前缀改为 "github.com/donkeydluffy/harness/。
func renamed(src string) string {
	return strings.ReplaceAll(src, `"harness/`, `"github.com/donkeydluffy/harness/`)
}

func TestCheck_AC_050_2_B2(t *testing.T) {
	before := map[string]string{}
	after := map[string]string{"go.mod": "module github.com/donkeydluffy/harness\n\ngo 1.21\n"}
	for i := 0; i < 45; i++ {
		var paths []string
		for j := 0; j < 10; j++ {
			paths = append(paths, fmt.Sprintf("harness/internal/p%02d", j))
		}
		p := fmt.Sprintf("internal/a%02d/a.go", i)
		before[p] = importer("a", "", paths...)
		after[p] = renamed(before[p])
	}
	dir := mechRepo(t, before)
	stage(t, dir, after)
	git(t, dir, "commit", "-q", "-m", mechMsg)
	c := checkLast(t, dir)
	wantPass(t, c)
}

// ---- AC-050-3 机械改动任务的提交只改 import 路径与模块声明 ----

// mechCommit 在 mechRepo 上提交 40 个文件的 import 改名与 go.mod 的 module 行，再应用 edit，返回检查结果。
func mechCommit(t *testing.T, edit func(after map[string]string)) checked {
	t.Helper()
	before := map[string]string{}
	after := map[string]string{"go.mod": "module github.com/donkeydluffy/harness\n\ngo 1.21\n"}
	for i := 0; i < 40; i++ {
		p := fmt.Sprintf("internal/a%02d/a.go", i)
		before[p] = importer("a", "", "harness/internal/b")
		after[p] = renamed(before[p])
	}
	dir := mechRepo(t, before)
	if edit != nil {
		edit(after)
	}
	stage(t, dir, after)
	git(t, dir, "commit", "-q", "-m", mechMsg)
	return checkLast(t, dir)
}

func TestCheck_AC_050_3_N1(t *testing.T) {
	wantPass(t, mechCommit(t, nil))
}

func TestCheck_AC_050_3_E1(t *testing.T) {
	c := mechCommit(t, func(after map[string]string) {
		after["internal/a07/a.go"] = strings.Replace(after["internal/a07/a.go"], "fmt.Println()", "fmt.Println(1)", 1)
	})
	v := wantViolation(t, c, "commit.05", "internal/a07/a.go", "不是 import 路径或模块声明")
	if v.File != "internal/a07/a.go" || v.Line != 9 {
		t.Fatalf("应给出文件与行号：%+v", v)
	}
}

func TestCheck_AC_050_3_E2(t *testing.T) {
	c := mechCommit(t, func(after map[string]string) {
		after["internal/a03/a.go"] = strings.Replace(after["internal/a03/a.go"], "github.com/donkeydluffy/harness/", "gitlab.com/other/harness/", 1)
	})
	wantViolation(t, c, "commit.05", "internal/a03/a.go", "不符合任何 Rewrite")
}

func TestCheck_AC_050_3_E3(t *testing.T) {
	c := mechCommit(t, func(after map[string]string) {
		after["internal/c/c.go"] = "package c\n"
	})
	wantViolation(t, c, "commit.05", "新增了文件 internal/c/c.go")
}

func TestCheck_AC_050_3_B1(t *testing.T) {
	wantPass(t, mechCommit(t, func(after map[string]string) {
		after["docs/PRD.md"] = "# PRD\n\n改名说明。\n"
	}))
}
