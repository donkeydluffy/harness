# MOD-02 app 详细设计

- Status: Draft

## 接口

### IF-022 Check
- Module: MOD-02
- Covers: FR-007, FR-017

```go
// IF-022
type CheckOptions struct{ Rules []string; Commits string; JSON bool }
func Check(root string, opt CheckOptions) (engine.Report, error)
func Trace(root, format string) ([]byte, error) // md / json 覆盖矩阵
```

### IF-023 Lock
- Module: MOD-02
- Covers: FR-004

```go
// IF-023
func Lock(root string) error // 产物文档规则（不需要源码、测试报告、git 的规则）通过后重算哈希写入锁文件；保留批准记录与阶段
```

### IF-024 Approve
- Module: MOD-02
- Covers: FR-037, FR-045

```go
// IF-024
type Terminal interface{ IsInteractive() bool; Confirm(prompt string) bool; Identity() string }
func Approve(root, artifact string, t Terminal) error
func PhaseNext(root string, t Terminal) error
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
