# MOD-08 testreport 详细设计

- Status: Approved

## 接口

### IF-008 Parse
- Module: MOD-08
- Covers: FR-020, FR-039

```go
// IF-008
type Status string // passed / failed / skipped
type TestCase struct{ Name, Suite, File string; Status Status }

func ParseJUnit(r io.Reader) ([]TestCase, error)
func ParseGoJSON(r io.Reader) ([]TestCase, error) // go test -json 输出，子测试名含父测试前缀
func ParseFile(path string) ([]TestCase, error)  // 按扩展名分派：.xml / .jsonl / .json
```

### IF-033 Reports
- Module: MOD-08
- Covers: FR-014, FR-021

```go
// IF-033
type LintIssue struct{ File string; Line int; Tool, Msg string }
type LintReport struct {
	SourceHash string
	Issues     []LintIssue
}
type IFCoverage struct {
	Tests    []string // 测试名含该接口编号的单元测试
	Passed   bool
	EntryHit bool     // 锚点所在函数的入口行被执行
}
type UnitReport struct {
	SourceHash string
	Interfaces map[string]IFCoverage // 接口编号 -> 结果
	Modules    map[string]float64    // 模块编号 -> 行覆盖率（0-100）
	Skipped    []string              // 不支持覆盖率采集的语言
}
// ReadLint、ReadUnit 读取 .harness/reports/ 下的报告；文件不存在时返回 nil 与 nil 错误。
func ReadLint(path string) (*LintReport, error)
func ReadUnit(path string) (*UnitReport, error)
func (r *LintReport) Write(path string) error // 问题按文件、行号排序
func (r *UnitReport) Write(path string) error
```
