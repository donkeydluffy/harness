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

## 内部设计

- `internal/lang/registry.go` 只包含各语言包的注册调用，每种语言一行。
- Release 1 提供 `internal/lang/golang/`：用 `go/scanner` 提取注释；`*_integration_test.go` 或含 `//go:build integration` 的文件为集成测试。
