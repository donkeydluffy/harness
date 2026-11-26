# MOD-10 phase 详细设计

- Status: Draft

Release 2 删除本模块（FR-048）：`Current-Release` 改为 2 时与 `internal/phase/` 一起删除，模块 Status 改为 Removed。

## 接口

### IF-015 Phases
- Module: MOD-10
- Covers: FR-045

```go
// IF-015
type Phase struct {
	Name     string   // requirements / architecture / planning / implementation
	Order    int
	Entry    []string // 规则编号
	Exit     []string
	Writable []string // 产物路径 glob
	Skill    string
	Builtin  bool
}
// Load 合并内置阶段（internal/phase/builtin/*.json，顺序见 order.json）与 harness.json 的 phases。
func Load(s *workspace.Snapshot) ([]Phase, error)
// Selector 把规则引用（编号、区域.*、*）转换为引擎选择器；列表为空时 ok 为 false。
func Selector(refs []string) (sel engine.Selector, ok bool)
// Rule 返回本模块注册的规则 phase.01，由 app 启动时注册。
func Rule() engine.Rule
```

### IF-016 Current
- Module: MOD-10
- Covers: FR-038, FR-045

```go
// IF-016
func Current(s *workspace.Snapshot, phases []Phase) (Phase, error) // 读锁文件中的 phase
// Next 检查当前阶段 Exit 与下一阶段 Entry，全部通过则返回下一阶段；不写文件。
func Next(s *workspace.Snapshot, phases []Phase) (Phase, engine.Report, error)
```

## 规则

本模块注册规则 `phase.01`：阶段定义合法、规则编号存在、阶段名不重复；移除内置规则须引用 ADR。

## 内置阶段

| 阶段 | 可写产物 | 退出规则 |
|---|---|---|
| requirements | `docs/PRD.md`, `docs/prd/**`, `docs/research/**` | prd.*, term.01, approve.01 |
| architecture | 上一阶段 + `docs/architecture/**`, `docs/adr/**` | arch.*, trace.*, adr.*, approve.01 |
| planning | 上一阶段 + `docs/tasks.md`, `docs/debt.md` | task.01 ~ task.06 |
| implementation | 上一阶段 + `docs/progress.md` | 全部规则 |

源码写入不受可写产物路径约束，由 FR-012 的任务范围判定。
