# MOD-17 custom 详细设计

- Status: Approved

## 职责

把 `harness.json` 的 `rules` 构造为实现 IF-011 的规则，校验编号、样本与批准前检查登记（FR-054）。外部命令规则在 Release 3（FR-046），本模块不运行外部程序。

## 接口

### IF-036 Load
- Module: MOD-17
- Covers: FR-054

```go
// IF-036
type Def struct {
	ID         string   // <项目前缀>.<两位以上序号>，如 acme.01；前缀不得是内置规则的区域
	Kind       string   // required / enum / ref / forbid
	Entity     string   // 作用的实体类型（如 FR、MOD、T）；forbid 改用 Paths
	Paths      []string // forbid 作用的文件 glob
	Field      string   // required、enum、ref 检查的字段
	Values     []string // enum 允许的取值
	Pattern    string   // forbid 禁用的正则
	Message    string
	Samples    string   // 样本目录，含 violation/ 与 ok/ 两个子目录，各为一个最小项目
	ApproveFor []string // 作为批准前检查的单位类型：requirement / design / adr / task / body
}

// Load 解析项目配置的 rules 并构造规则；定义有误（类型不支持、编号无前缀或与内置规则冲突、
// 批准单位类型不存在）时返回违规，该条规则不构造（AC-054-1、AC-054-4、AC-054-5）。
func Load(s *workspace.Snapshot, builtin []string) ([]engine.Rule, []engine.Violation)

// Verify 用每条规则的样本运行 IF-013；样本缺失、违规样本未报告或合规样本被报告时返回违规（AC-054-3）。
func Verify(root string, rules []engine.Rule) []engine.Violation

// ApproveChecks 返回登记在某类批准单位上的规则编号。
func ApproveChecks(rules []engine.Rule, kind string) []string
```

## 内部设计

- 四类规则：`required` 实体缺字段；`enum` 字段值不在 Values 中；`ref` 字段引用的编号不存在；`forbid` 匹配 Paths 的文件中出现 Pattern，逐行报告。
- 样本目录与规则定义一起纳入锁文件（formats.md 第 7 节 `rule:<编号>`），修改后需 `harness lock`（AC-054-2）。
