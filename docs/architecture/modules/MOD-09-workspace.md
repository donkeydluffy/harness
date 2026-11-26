# MOD-09 workspace 详细设计

- Status: Approved

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
	Lint    *testreport.LintReport // WithTests 时读取，不存在为 nil
	Unit    *testreport.UnitReport
	Commands []string         // cli 的子命令名，由 app 从 CheckOptions 填入
	Diags   []model.Diagnostic
}
type LoadOptions struct{ WithTests, WithSources bool; CommitRange, Message string } // Message：即将形成的提交的信息，检查暂存区时使用

// SourceHash 返回 sourcePaths 内全部源文件（按路径排序的路径与内容）的哈希，报告据此判断是否过期。
func (s *Snapshot) SourceHash() string

func Load(root string, opt LoadOptions) (*Snapshot, error)
// 批准（FR-037）：可批准的产物、批准覆盖的文档、批准哈希（不含文件头 Status 行）。
func Approvable(rel string) bool
func (s *Snapshot) ApprovalDocs(rel string) []*model.Doc
func (s *Snapshot) ApprovalHash(rel string) string
```

### IF-010 Config
- Module: MOD-09
- Covers: FR-014, FR-041, FR-055

```go
// IF-010
type Config struct { /* 字段见 formats.md 第 6 节；Release 2 新增 tools、start，删除 phases */ }
func LoadConfig(path string) (*Config, error) // 缺省字段填默认值
```

## 内部设计

门禁路径（PreToolUse）只用 `LoadOptions{}`，不读测试报告和源码，以满足 NFR-003 的 200ms。
