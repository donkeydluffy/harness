# MOD-02 app 详细设计

- Status: Approved

## 接口

### IF-022 Check
- Module: MOD-02
- Covers: FR-007, FR-017

```go
// IF-022
type CheckOptions struct {
	Rules       []string // 规则编号或前缀；空表示全部
	Commits     string   // 提交范围；为空时检查暂存区
	MessageFile string   // commit-msg hook 传入的提交信息文件
	JSON        bool
	Commands    []string // cli 的全部子命令名（AC-030-2）
}
// Check 注册内置规则与项目自定义规则（IF-036），先用样本验证自定义规则（IF-013），再运行引擎。
// 返回的 error 只表示内部错误，与"有违规"区分。
func Check(root string, opt CheckOptions) (engine.Report, error)

// TraceRow 是覆盖矩阵的一行：一条纳入范围的需求及覆盖它的有效设计元素。
type TraceRow struct {
	Requirement string   `json:"requirement"`
	Title       string   `json:"title"`
	Elements    []string `json:"elements"`
}
var TraceFormats = []string{"md", "json"}
func Trace(root, format string) ([]byte, error)
```

### IF-023 Lock
- Module: MOD-02
- Covers: FR-004, FR-047, FR-048

```go
// IF-023
// ViolationsError 表示命令因检查未通过而没有执行（Lock、Approve 返回），cli 据此输出违规并以退出码 1 结束。
type ViolationsError struct{ Report engine.Report }

// Lock 在产物文档规则（不需要源码、测试报告、git 的规则）通过后重算哈希写入锁文件；保留批准记录。
// 写出的锁文件总是第 2 版：第 1 版的按文件批准经 IF-029 Migrate 迁移，读到的 phase 字段丢弃（AC-048-1）。
func Lock(root string) error
```

### IF-024 Approve
- Module: MOD-02
- Covers: FR-037, FR-047, FR-054

```go
// IF-024
type Terminal interface{ IsInteractive() bool; Confirm(prompt string) bool; Identity() string }
var ErrNotInteractive, ErrDeclined error // 不在交互式终端中；批准人在确认提示处拒绝
// Approve 列出 target（产物路径、prd 或单位编号）范围内待批准的单位与差异（IF-029），
// 运行批准前检查（MOD-15 内部设计、AC-054-5），人确认后把批准记录写入锁文件；没有待批准单位时提示并返回 nil。
// 锁文件仍有第 1 版的按文件批准时返回错误，提示先运行 harness lock；检查未通过时返回 *ViolationsError。
func Approve(root, target string, t Terminal) error
```

### IF-025 Gate
- Module: MOD-02
- Covers: FR-012, FR-024

```go
// IF-025
// Gate 加载快照并调用门禁判定；cli 负责用 adapter 解码代理输入、编码判定结果。
func Gate(root string, req gate.Request) gate.Decision
```

### IF-026 Init
- Module: MOD-02
- Covers: FR-001, FR-043

```go
// IF-026
func Init(root string, opt scaffold.InitOptions) (scaffold.Result, error)
```

### IF-038 Run
- Module: MOD-02
- Covers: FR-014, FR-021, FR-055

```go
// IF-038
// MissingTool 是工具未安装时 Run 返回的错误，cli 据此给出安装提示（AC-014-2）。
type MissingTool = toolrun.MissingTool
var ToolKinds = []string{"lint", "unit", "smoke"}
// Run 运行外部工具并写报告（kind 取 ToolKinds，调用 IF-030 ~ IF-032），
// 然后重新加载快照、运行同名区域的规则（lint.*、unit.*、smoke.*）并返回结果。
func Run(root, kind string) (engine.Report, error)
```

### IF-039 Snapshot
- Module: MOD-02
- Covers: FR-052

```go
// IF-039
// Snapshot 从源码锚点提取接口签名（IF-035），写入接口快照文件 docs/architecture/interfaces.txt（AC-052-2）。
func Snapshot(root string) error
```
