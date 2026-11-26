# MOD-16 toolrun 详细设计

- Status: Approved

## 职责

运行外部工具（格式化与静态检查、按接口的单元测试与覆盖率、冒烟测试），把结果写成报告。全量检查不调用本模块，只读取报告（ADR-011）。

## 接口

### IF-030 Lint
- Module: MOD-16
- Covers: FR-014

```go
// IF-030
// MissingTool 是工具未安装时的错误；Hint 为安装命令，项目覆盖的命令没有内置提示时给出命令名（AC-014-2）。
type MissingTool struct{ Language, Command, Hint string }

// Lint 对项目声明且存在源文件的每种语言，运行格式化检查与静态检查（项目配置的 tools 优先，否则用语言配置包的默认值），
// 把问题与源码哈希写入 .harness/reports/lint.json（AC-014-3、AC-014-4）。
func Lint(s *workspace.Snapshot) (*testreport.LintReport, error)
```

### IF-031 Unit
- Module: MOD-16
- Covers: FR-021

```go
// IF-031
// Unit 对每个带锚点的 Active 接口单独运行测试名含其编号的单元测试，记录锚点所在函数入口行是否执行；
// 再全量运行一次单元测试统计各模块行覆盖率；结果与源码哈希写入 .harness/reports/unit.json（AC-021-2、AC-021-3）。
// 项目没有支持覆盖率的语言时返回 Skipped 不为空的报告。
func Unit(s *workspace.Snapshot) (*testreport.UnitReport, error)
```

### IF-032 Smoke
- Module: MOD-16
- Covers: FR-055

```go
// IF-032
// Smoke 运行项目配置的冒烟测试命令，返回退出码与输出的最后 50 行（AC-055-2）。
func Smoke(s *workspace.Snapshot) (exit int, tail []string, err error)
```

## 内部设计

- 工具输出解析由语言配置包声明（IF-034 的 `Tool.Output`）：`files` 每行一个文件名（如 `gofmt -l`），`lines` 为 `文件:行[:列]: 消息`。项目覆盖的命令按 `lines` 解析；命令退出码非 0 而没有解析出问题时，记一条不带文件的问题，消息为输出的最后 20 行。
- 命令通过 `sh -c` 在项目根运行；命令的第一个词在 PATH 中找不到时返回 MissingTool，不写报告。
- 源码哈希由 IF-009 的 `SourceHash` 计算，报告与当前源码的哈希不同即过期。
