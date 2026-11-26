# MOD-06 lock 详细设计

- Status: Approved

## 接口

### IF-005 Normalize
- Module: MOD-06
- Covers: FR-004, FR-037, FR-041

```go
// IF-005
// Normalize 统一换行为 LF，去掉行尾空白，合并连续空行，去掉首尾空行。
func Normalize(text string) string
// Hash 返回 "sha256:<hex>"，输入先经 Normalize。
func Hash(text string) string
```

### IF-006 Lockfile
- Module: MOD-06
- Covers: FR-004, FR-037, FR-041, FR-045

```go
// IF-006
type Lock struct {
	Version   int
	Phase     string
	Items     map[string]string
	Approvals []Approval
	Audit     map[string]json.RawMessage
}
type Approval struct{ Artifact, Hash, By string; At time.Time }

func Read(path string) (*Lock, error) // 文件不存在时返回空 Lock 而非错误
func (l *Lock) Write(path string) error // 键排序，末尾换行，两次写出逐字节一致
func (l *Lock) Approved(artifact, currentHash string) bool
```
