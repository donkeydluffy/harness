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
