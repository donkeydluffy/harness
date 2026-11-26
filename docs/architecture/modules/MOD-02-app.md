# MOD-02 app 详细设计

- Status: Draft

## 接口

### IF-022 Check
- Module: MOD-02
- Covers: FR-007, FR-017

```go
// IF-022
type CheckOptions struct {
	Rules                []string
	Commits, MessageFile string // MessageFile：commit-msg hook 传入的提交信息文件
	JSON                 bool
	Commands             []string // cli 的全部子命令名（AC-030-2）
}
// Check 注册内置规则与项目自定义规则（IF-036），先用样本验证自定义规则（IF-013），再运行引擎。
func Check(root string, opt CheckOptions) (engine.Report, error)
func Trace(root, format string) ([]byte, error) // md / json 覆盖矩阵
```

### IF-023 Lock
- Module: MOD-02
- Covers: FR-004, FR-047, FR-048

```go
// IF-023
// Lock 在产物文档规则（不需要源码、测试报告、git 的规则）通过后重算哈希写入锁文件；保留批准记录，
// Release 1 格式的按文件批准经 IF-029 Migrate 迁移，删除 phase 字段。
func Lock(root string) error
```

### IF-024 Approve
- Module: MOD-02
- Covers: FR-037, FR-047, FR-054

```go
// IF-024
type Terminal interface{ IsInteractive() bool; Confirm(prompt string) bool; Identity() string }
// Approve 列出 target 范围内待批准的单位与差异（IF-029），运行批准前检查（MOD-15 内部设计、AC-054-5），
// 人确认后把批准记录写入锁文件；没有待批准单位时提示并返回 nil。
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
// Run 运行外部工具并写报告（kind 为 lint、unit、smoke，调用 IF-030 ~ IF-032），
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
