# MOD-09 workspace 详细设计

- Status: Draft

## 职责

一次性加载规则和门禁需要的全部输入，形成只读快照。规则只读快照，不自行读文件，便于样本测试和性能控制。

## 接口

### IF-009 Load
- Module: MOD-09
- Covers: FR-003, FR-041

```go
// IF-009
type Snapshot struct {
	Root    string
	Config  *Config
	Project *model.Project
	Lock    *lock.Lock
	Repo    vcs.Repo          // 非 git 仓库时为 nil
	Tests   []testreport.TestCase
	Sources []SourceFile      // 按 sourcePaths 收集，惰性读取内容
	Diags   []model.Diagnostic
}
type LoadOptions struct{ WithTests, WithSources bool; CommitRange string }

func Load(root string, opt LoadOptions) (*Snapshot, error)
```

### IF-010 Config
- Module: MOD-09
- Covers: FR-041, FR-045

```go
// IF-010
type Config struct { /* 字段见 formats.md 第 6 节 */ }
func LoadConfig(path string) (*Config, error) // 缺省字段填默认值
```

## 内部设计

门禁路径（PreToolUse）只用 `LoadOptions{}`，不读测试报告和源码，以满足 NFR-003 的 200ms。
