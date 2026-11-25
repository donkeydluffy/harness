# MOD-04 engine 详细设计

- Status: Draft

## 接口

### IF-011 Rule
- Module: MOD-04
- Covers: FR-028, FR-046

```go
// IF-011
type Rule interface {
	ID() string                 // 如 "trace.01"
	Covers() []string           // 对应的验收标准，如 "AC-007-1"
	Fix() string                // 默认修复建议
	Check(s *workspace.Snapshot) []Violation
}
type Violation struct{ Rule, File string; Line int; Msg, Fix string }

func Register(r Rule) // 编号重复时 panic（构建期即暴露）
func Lookup(id string) (Rule, bool)
```

### IF-012 Run
- Module: MOD-04
- Covers: FR-045, NFR-004

```go
// IF-012
type Selector struct{ IDs []string; Prefixes []string } // 空选择器表示全部
type Report struct{ Violations []Violation; Ran []string; Errors []error }

func Run(s *workspace.Snapshot, sel Selector) Report // 输出按 文件、行号、规则编号 排序
```

### IF-013 Verify
- Module: MOD-04
- Covers: FR-028, FR-046

```go
// IF-013
// Verify 用违规样本和合规样本验证规则：违规样本必须报告，合规样本必须不报告。
type Sample struct{ Name string; Snapshot *workspace.Snapshot; WantViolation bool }
func Verify(r Rule, samples []Sample) error
```

## 内部设计

- 项目自定义规则（Release 2）实现同一个 Rule 接口：声明式规则由配置构造，外部命令规则通过子进程执行。
- 违规缺少修复建议时，Run 用规则的 `Fix()` 补齐；两者都空视为规则缺陷（自身测试失败，AC-N004-1/E1）。
