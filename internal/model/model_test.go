package model

import (
	"os"
	"path/filepath"
	"testing"
)

const sample = `# PRD: demo

- Status: Draft
- Current-Release: 1

## 2. 统一语言

| ID | 术语 | 代码名 | 定义 | 禁用别名 |
|---|---|---|---|---|
| TERM-001 | 产物 | Artifact | 文档 | 交付物、文件 |

## 3. 功能需求

### FR-001 初始化
- Status: Active
- Release: 1
- Priority: P0

描述。

- AC-001-1: 生成骨架。
  - AC-001-1/N1: 空目录
  - AC-001-1/E1: 只读目录

` + "```go\n### FR-999 代码块中的标题不是实体\n```\n" + `
### FR-3 格式错误
- Status: Active

## 8. 变更日志

| ID | 日期 | 动作 | 对象 | 说明 |
|---|---|---|---|---|
| CHG-001 | 2026-11-25 | 新增 | FR-001, FR-002 | 初稿 |
`

func TestParseMarkdown_IF_001(t *testing.T) {
	d, diags := ParseMarkdown("docs/PRD.md", []byte(sample))
	if len(diags) != 0 {
		t.Fatalf("diags: %v", diags)
	}
	if d.Header["Current-Release"] != "1" || d.Header["Status"] != "Draft" {
		t.Fatalf("header: %v", d.Header)
	}
	if len(d.Entities) != 2 {
		t.Fatalf("entities: %d", len(d.Entities))
	}
	fr := d.Entities[0]
	if fr.ID != "FR-001" || fr.Kind != "FR" || fr.Title != "初始化" || fr.Line != 14 {
		t.Fatalf("entity: %+v", fr)
	}
	if fr.Fields["Priority"] != "P0" || len(fr.Order) != 3 {
		t.Fatalf("fields: %v", fr.Fields)
	}
	if len(fr.Items) != 1 || len(fr.Items[0].Children) != 2 || fr.Items[0].Children[1].ID != "AC-001-1/E1" {
		t.Fatalf("items: %+v", fr.Items)
	}
	if len(fr.Code) != 1 {
		t.Fatalf("code blocks: %d", len(fr.Code))
	}
	if d.Entities[1].ID != "FR-3" {
		t.Fatalf("loose id: %s", d.Entities[1].ID)
	}
	if len(d.Tables) != 2 || d.Tables[0].Rows[0].Line != 10 {
		t.Fatalf("tables: %+v", d.Tables)
	}
}

func TestParseMarkdown_ADRHeader_IF_001(t *testing.T) {
	src := "# ADR-001 用 Go\n\n- Status: Accepted\n- Relates: NFR-001\n\n## 背景\n\n文字\n"
	d, _ := ParseMarkdown("docs/adr/ADR-001.md", []byte(src))
	if len(d.Entities) != 1 || d.Entities[0].Fields["Relates"] != "NFR-001" || d.Header["Status"] != "Accepted" {
		t.Fatalf("adr: %+v header %v", d.Entities, d.Header)
	}
}

func TestParseMarkdown_ModuleDetailTitleNotEntity_IF_001(t *testing.T) {
	src := "# MOD-03 model 详细设计\n\n- Status: Approved\n\n### IF-001 Parse\n- Module: MOD-03\n"
	d, _ := ParseMarkdown("m.md", []byte(src))
	if len(d.Entities) != 1 || d.Entities[0].ID != "IF-001" {
		t.Fatalf("entities: %+v", d.Entities)
	}
}

func TestBuild_IF_002(t *testing.T) {
	a, _ := ParseMarkdown("docs/PRD.md", []byte(sample))
	b, _ := ParseMarkdown("docs/prd/x.md", []byte("### FR-001 重复\n- Status: Active\n"))
	p, diags := Build([]*Doc{a, b})
	if len(diags) != 1 {
		t.Fatalf("want 1 duplicate diag, got %v", diags)
	}
	if c := p.Cases["AC-001-1/E1"]; c == nil || c.Category != 'E' || c.Req != "FR-001" || c.AC != "AC-001-1" {
		t.Fatalf("case: %+v", c)
	}
	if len(p.Terms) != 1 || len(p.Terms[0].Banned) != 2 {
		t.Fatalf("terms: %+v", p.Terms)
	}
	if len(p.Changes) != 1 || len(p.Changes[0].Targets) != 2 {
		t.Fatalf("changes: %+v", p.Changes)
	}
}

// 本仓库自己的产物必须能无诊断解析。
func TestParseRepoArtifacts_IF_001(t *testing.T) {
	files, _ := filepath.Glob("../../docs/*.md")
	more, _ := filepath.Glob("../../docs/*/*.md")
	more2, _ := filepath.Glob("../../docs/*/*/*.md")
	files = append(append(files, more...), more2...)
	var docs []*Doc
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		d, diags := ParseMarkdown(f, src)
		if len(diags) > 0 {
			t.Errorf("%s: %v", f, diags)
		}
		docs = append(docs, d)
	}
	p, diags := Build(docs)
	if len(diags) > 0 {
		t.Errorf("build: %v", diags)
	}
	if len(p.ByKind["FR"]) != 46 || len(p.ByKind["MOD"]) != 14 || len(p.ByKind["IF"]) != 26 || len(p.ByKind["T"]) != 27 {
		t.Errorf("counts FR=%d MOD=%d IF=%d T=%d", len(p.ByKind["FR"]), len(p.ByKind["MOD"]), len(p.ByKind["IF"]), len(p.ByKind["T"]))
	}
}
