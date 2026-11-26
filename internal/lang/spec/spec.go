// Package spec 定义语言包的类型。语言实现只依赖本包，internal/lang 通过类型别名对外提供（IF-003）。
package spec

// TestKind 是测试文件的分类。
type TestKind int

const (
	NotTest TestKind = iota
	Unit
	Integration
)

// Comment 是源码中的一条注释。
type Comment struct {
	Text string
	Line int
}

// Language 描述一种编程语言（IF-003）。
//
// harness:IF-003
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

// Tool 是语言配置包中的一个外部工具（IF-034）。
type Tool struct {
	Cmd     string // 在项目根以 sh -c 运行
	Output  string // files：每行一个文件名；lines：文件:行[:列]: 消息
	Install string // 安装命令提示
}

// Pack 是语言配置包：源码路径、CI 测试命令、格式化检查与静态检查（IF-034）。
type Pack struct {
	SourcePaths  []string
	Test         string // CI 中生成测试报告的命令；为空表示没有通用命令
	Format, Lint Tool
}

// Packer 是语言的可选接口；声明了语言配置包的语言才能在 init 与 lint 中使用。
type Packer interface{ Pack() Pack }

// Import 是源码中的一条 import（IF-035）。
type Import struct {
	Path string
	Line int
}

// Decl 是锚点紧随的声明（IF-035）。
type Decl struct {
	ID        string // 锚点上的接口编号
	File      string
	Line      int    // 锚点行
	EntryLine int    // 函数或方法体的第一行；类型声明为 0
	Signature string // 去掉注释与函数体、空白规范化后的声明
}

// Analyzer 是语言的可选接口，Release 2 只有 Go 实现（ADR-012）。
//
// harness:IF-035
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
