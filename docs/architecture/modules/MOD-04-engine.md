# MOD-04 engine 详细设计

- Status: Approved

## 接口

### IF-011 Rule
- Module: MOD-04
- Covers: FR-028, FR-054

```go
// IF-011
type Rule interface {
	ID() string                 // 如 "trace.01"
	Covers() []string           // 对应的验收标准，如 "AC-007-1"
	Fix() string                // 默认修复建议
	Check(s *workspace.Snapshot) []Violation
}
type Violation struct {
	Rule, File string
	Line       int
	Msg, Fix   string
	Warn       bool // 告警：输出但不影响退出码（AC-030-4、AC-047-1/E1）
}

// Needer 是可选接口：规则声明在产物文档之外需要的输入。未实现表示只读产物文档。
type Input uint8 // InputSources | InputTests | InputGit
type Needer interface{ Needs() Input }
func Needs(r Rule) Input

func Register(r Rule) // 编号重复时 panic（构建期即暴露）
func Lookup(id string) (Rule, bool)
```

### IF-012 Run
- Module: MOD-04
- Covers: FR-030, NFR-004

```go
// IF-012
type Selector struct{ IDs []string; Prefixes []string } // 空选择器表示全部
type Report struct{ Violations []Violation; Ran []string; Errors []error }

func Run(s *workspace.Snapshot, sel Selector) Report // 输出按 文件、行号、规则编号 排序
```

### IF-013 Verify
- Module: MOD-04
- Covers: FR-028, FR-054

```go
// IF-013
// Verify 用违规样本和合规样本验证规则：违规样本必须报告，合规样本必须不报告。
type Sample struct{ Name string; Snapshot *workspace.Snapshot; WantViolation bool }
func Verify(r Rule, samples []Sample) error
```

## 内部设计

- 项目自定义规则实现同一个 Rule 接口：Release 2 的声明式规则由 MOD-17 按配置构造；外部命令规则在 Release 3（FR-046）。
- Report 的退出码只看 `Warn` 为 false 的违规。
- 违规缺少修复建议时，Run 用规则的 `Fix()` 补齐；两者都空视为规则缺陷（自身测试失败，AC-N004-1/E1）。
