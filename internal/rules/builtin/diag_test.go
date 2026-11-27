package builtin

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// 每条解析诊断恰好被一条规则报告（T-032）：结构规则报告各自文件的诊断，prd.07 报告用例诊断，其余由 doc.01 报告。
func TestDiagnosticsReportedOnce(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"docs/PRD.md":                   "# PRD\n\n- Status: Draft\n\n```go\nx\n",
		"docs/progress.md":              "# 进度\n\n```\nlog\n",
		"docs/architecture/overview.md": "# 架构\n\n- Status: Draft\n\n```\nx\n",
		"docs/tasks.md":                 "# 任务\n\n- Status: Draft\n\n```\nx\n",
	}
	for p, c := range files {
		os.MkdirAll(filepath.Dir(filepath.Join(dir, p)), 0o755)
		os.WriteFile(filepath.Join(dir, p), []byte(c), 0o644)
	}
	s, err := workspace.Load(dir, workspace.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Diags) < len(files) {
		t.Fatalf("夹具应在每个文件产生诊断：%v", s.Diags)
	}
	// 各文件的诊断应由哪条规则报告：按区域选取规则时（如 harness check arch）归属不能错。
	owner := map[string]string{
		"docs/PRD.md": "doc.01", "docs/progress.md": "doc.01",
		"docs/architecture/overview.md": "arch.01", "docs/tasks.md": "task.01",
	}
	by := map[string][]string{}
	for _, r := range Rules() {
		for _, v := range r.Check(s) {
			for _, d := range s.Diags {
				if v.File == d.File && v.Line == d.Line && v.Msg == d.Msg {
					by[d.File+":"+d.Msg] = append(by[d.File+":"+d.Msg], r.ID())
				}
			}
		}
	}
	for _, d := range s.Diags {
		if got := by[d.File+":"+d.Msg]; len(got) != 1 || got[0] != owner[d.File] {
			t.Errorf("%s:%d %s 应只由 %s 报告，实际：%v", d.File, d.Line, d.Msg, owner[d.File], got)
		}
	}
}

// 全量检查（任务完成定义、会话结束检查）包含 test.03 与 doc.01，测试失败或解析出错时不能通过（AC-039-2）。
// 全量检查是全部已注册规则减去 engine.CoverageRules，因此只需两者是内置规则且不属于覆盖类规则。
func TestFullCheckIncludesFailedTests(t *testing.T) {
	var ids []string
	for _, r := range Rules() {
		ids = append(ids, r.ID())
	}
	for _, id := range []string{"test.03", "doc.01"} {
		if !slices.Contains(ids, id) || slices.Contains(engine.CoverageRules, id) {
			t.Errorf("%s 应是内置规则且计入全量检查", id)
		}
	}
}
