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
