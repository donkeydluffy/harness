# MOD-14 lang 详细设计

- Status: Draft

## 职责

语言相关知识的唯一位置：文件归属、注释提取（区分字符串字面量）、测试分类。新增语言是扩展点 EXT-01。

## 接口

### IF-003 Language
- Module: MOD-14
- Covers: FR-020, FR-031

```go
// IF-003
type Language interface {
	Name() string
	Extensions() []string
	// Comments 返回源码中的注释及行号，字符串字面量中的内容不计入。
	Comments(src []byte) []Comment
	// TestKind 判定测试文件是单元测试还是集成测试；非测试文件返回 NotTest。
	TestKind(path string, src []byte) TestKind
	// TestNames 返回测试文件中声明的测试名，用于"静态声明但报告中缺失"的检查。
	TestNames(src []byte) []string
}
```

### IF-004 Registry
- Module: MOD-14
- Covers: FR-020, FR-031

```go
// IF-004
func Register(l Language)
func ForPath(path string) (Language, bool)
func All() []Language
```

### IF-034 Pack
- Module: MOD-14
- Covers: FR-014

```go
// IF-034
type Tool struct {
	Cmd     string // 在项目根以 sh -c 运行
	Output  string // files：每行一个文件名；lines：文件:行[:列]: 消息
	Install string // 安装命令提示
}
type Pack struct {
	SourcePaths  []string
	Test         string // CI 中的测试命令
	Format, Lint Tool
}
// Packer 是语言的可选接口；声明了语言配置包的语言才能在 init 与 lint 中使用。
type Packer interface{ Pack() Pack }
func PackFor(name string) (Pack, bool)
```

默认工具见 PRD FR-014 的表格。

### IF-035 Analyzer
- Module: MOD-14
- Covers: FR-021, FR-050, FR-052, FR-053

```go
// IF-035
type Import struct{ Path string; Line int }
type Decl struct {
	ID        string // 锚点上的接口编号
	File      string
	Line      int    // 锚点行
	EntryLine int    // 函数或方法体的第一行；类型声明为 0
	Signature string // 去掉注释与函数体、空白规范化后的声明
}
// Analyzer 是语言的可选接口，Release 2 只有 Go 实现（ADR-012）。
type Analyzer interface {
	Imports(src []byte) []Import
	// ModuleDecl 返回模块声明文件（如 go.mod）中的模块路径及其行号；不是该文件时 ok 为 false。
	ModuleDecl(path string, src []byte) (mod string, line int, ok bool)
	// Anchors 返回 `// harness:IF-n` 锚点紧随的声明；锚点后不是声明时 Signature 为空。
	Anchors(path string, src []byte) []Decl
	// Resolve 把 import 路径解析为仓库相对目录；标准库与第三方包 ok 为 false。
	Resolve(root, importPath string) (dir string, ok bool)
	// Coverage 运行名称匹配 pattern 的单元测试并返回结果与已执行的行（文件 -> 行号）；pattern 为空时运行全部单元测试。
	Coverage(root, pattern string) (passed bool, tests []string, executed map[string][]int, err error)
}
func AnalyzerFor(path string) (Analyzer, bool)
```

## 内部设计

- `internal/lang/registry.go` 只包含各语言包的注册调用，每种语言一行。
- Release 1 提供 `internal/lang/golang/`：用 `go/scanner` 提取注释；`*_integration_test.go` 或含 `//go:build integration` 的文件为集成测试。
- Release 2：五种语言都实现 Packer（原 `internal/scaffold/templates/langs/<语言>/pack.json` 的数据移入各语言包）；只有 Go 实现 Analyzer：`go/parser` 解析 import 与声明，`go test -run <pattern> -coverprofile -coverpkg=./...` 采集覆盖率。
