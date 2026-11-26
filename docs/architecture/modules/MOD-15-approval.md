# MOD-15 approval 详细设计

- Status: Draft

## 职责

按批准单位（TERM-027）计算内容哈希、批准链与批准状态，生成批准记录（ADR-009）。只读快照，不写文件；写锁文件由 app 完成。

## 接口

### IF-027 Units
- Module: MOD-15
- Covers: FR-047

```go
// IF-027
type UnitKind string // requirement / term / question / design / adr / task / body

type Unit struct {
	ID       string   // 如 FR-002、MOD-02、ADR-009、T-040、docs/PRD.md#正文
	Kind     UnitKind
	File     string
	Line     int
	Hash     string   // 规范化内容哈希（IF-005），不含 Status 字段；PRD 正文不含变更日志
	Text     string   // 规范化内容，批准时记入锁文件，用于显示下次批准的差异
	Upstream []string // 批准链上游的单位编号：设计元素取 Covers、Relates，任务取 Covers、Design、Flow；AC 归到所属需求
}

// Units 返回快照中的全部批准单位，按文件、行号排序。
func Units(s *workspace.Snapshot) []Unit
```

### IF-028 Evaluate
- Module: MOD-15
- Covers: FR-047, FR-048

```go
// IF-028
type State string // approved / unapproved / stale

type Status struct {
	Unit  Unit
	State State
	Cause []string // stale 时引起失效的单位编号：内容变化时含自身，上游变化或失效时为上游编号
}

// Evaluate 返回每个单位的批准状态。批准记录存在且自身哈希、记录的上游哈希都与当前一致，
// 且每个上游本身为 approved 时为 approved；没有记录为 unapproved；其余为 stale。
func Evaluate(s *workspace.Snapshot) map[string]Status

// Chain 返回单位及其全部上游（递归）的状态，按编号排序；任务开工与源码写入时使用（AC-047-4）。
func Chain(s *workspace.Snapshot, id string) []Status

// Summary 是会话开始注入的批准状态摘要（AC-048-3）。
type Summary struct {
	Unapproved, Stale int
	Ready, InProgress []string // 可开工的任务、进行中的任务
	OpenQuestions     []string
	Skill             string   // clarify / architect / plan / implement
}
func Summarize(s *workspace.Snapshot) Summary
```

### IF-029 Pending
- Module: MOD-15
- Covers: FR-047

```go
// IF-029
type Pending struct {
	Unit   Unit
	Before string // 上次批准时的内容；新增单位为空
}

// PendingIn 返回 target 范围内新增、修改、批准失效的单位。target 为产物路径、"prd" 或单位编号；
// 无法识别时返回错误。
func PendingIn(s *workspace.Snapshot, target string) ([]Pending, error)

// Record 为确认批准的单位生成批准记录，记录各上游单位当前的哈希；不写文件。
func Record(s *workspace.Snapshot, units []Unit, by string, at time.Time) map[string]lock.UnitApproval

// Migrate 把 Release 1 的按文件批准换算为按单位批准（AC-047-6）：仍有效的文件批准换算为文件内全部单位，
// 已失效的文件中的单位不记录；返回写入审计记录的说明。
func Migrate(s *workspace.Snapshot) (map[string]lock.UnitApproval, json.RawMessage)
```

## 内部设计

- 单位的范围：需求为 FR、NFR 实体（含其验收标准与用例）；术语为统一语言表格的一行；待澄清项、设计元素（MOD、IF、FLOW、EXT）、ADR、任务为各自的实体；每个产物文件中实体之外的文字为 `<文件路径>#正文`。
- 批准前检查（IF-024 调用）：只运行待批准单位自身所在文件的结构规则、上游必须 approved（AC-047-3），以及项目登记在该单位类型上的自定义规则（AC-054-5）；不运行覆盖类规则。
- 状态由 Evaluate 一次算出并缓存于调用方，门禁路径只为当前 InProgress 任务调用 Chain，保持 NFR-003 的耗时要求。
