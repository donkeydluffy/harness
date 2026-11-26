package builtin

import (
	"fmt"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// test03：测试报告中不能有失败的测试。test.02 只看引用用例编号的测试，单元测试等其他测试失败时由本规则报告，
// 全量检查（任务完成定义、会话结束检查）因此不会在测试失败时通过。
type test03 struct{}

func (test03) Needs() engine.Input { return engine.InputTests }
func (test03) ID() string          { return "test.03" }
func (test03) Covers() []string    { return []string{"AC-039-2"} }
func (test03) Fix() string {
	return "修复失败的测试，重新运行测试并把报告写到 harness.json 的 testReports 路径"
}

func (test03) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, tc := range s.Tests {
		if tc.Status != testreport.Failed {
			continue
		}
		file := tc.File
		if file == "" {
			file = "."
		}
		where := ""
		if tc.Suite != "" {
			where = "（" + tc.Suite + "）"
		}
		out = append(out, engine.Violation{File: file, Line: 1, Msg: fmt.Sprintf("测试 %s%s 失败", tc.Name, where)})
	}
	return out
}
