# MOD-03 model 详细设计

- Status: Draft

## 职责

把 Markdown 产物解析为实体（见 `docs/architecture/formats.md`），再组装为项目模型。只做结构解析，不做规则判断；结构错误以 Diagnostic 返回，由规则 `prd.01` 等转换为违规。

## 接口

### IF-001 ParseMarkdown
- Module: MOD-03
- Covers: FR-003, FR-006, FR-019

```go
// IF-001
// ParseMarkdown 把单个 Markdown 文件解析为实体列表，同时保留表格行和代码块。
func ParseMarkdown(path string, src []byte) (*Doc, []Diagnostic)

type Doc struct {
	Path     string
	Header   map[string]string // 文件头字段，如 Status、Current-Release
	Entities []*Entity
	Tables   []Table           // 表格，每行带行号
}

type Entity struct {
	ID     string            // FR-001、MOD-04……
	Kind   string            // 由前缀推得：FR、NFR、MOD、IF……
	Title  string
	Fields map[string]string
	Items  []Item            // 列表项：AC 与其下缩进的用例
	Code   []CodeBlock
	File   string
	Line   int
	Raw    string            // 实体完整文本，用于哈希
}

type Diagnostic struct{ File string; Line int; Msg string }
```

### IF-002 Build
- Module: MOD-03
- Covers: FR-003, FR-006, FR-010, FR-011, FR-031, FR-033

```go
// IF-002
// Build 把多个 Doc 组装为项目模型，按编号建立索引；重复编号作为 Diagnostic 返回。
func Build(docs []*Doc) (*Project, []Diagnostic)

type Project struct {
	Docs     []*Doc
	ByID     map[string]*Entity
	ByKind   map[string][]*Entity
	Cases    map[string]*Case // 用例编号 -> 用例
	Terms    []Term           // 统一语言表格
	Changes  []Change         // 变更日志表格（同一 CHG 编号可多行）
}
```

## 内部设计

- 逐行扫描的状态机：标题 → 字段区 → 正文（列表项、表格、代码块）。不引入 Markdown 库。
- 代码格式（反引号、代码块）中的文字单独标记，供模糊词、禁用别名等规则跳过。
