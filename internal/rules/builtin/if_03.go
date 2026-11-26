package builtin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// if03：详细设计中接口代码块里与快照同名的声明，规范化后与快照中的签名一致（AC-052-3）。
type if03 struct{}

func (if03) ID() string       { return "if.03" }
func (if03) Covers() []string { return []string{"AC-052-3"} }
func (if03) Fix() string {
	return "让详细设计的签名与代码一致：改代码后运行 harness snapshot，或改设计（设计改动须由人重新批准）"
}

var declNameRe = regexp.MustCompile(`^(?:func (\([^)]*\) )?|(?:type|var|const) )([A-Za-z_]\w*)`)

func (if03) Check(s *workspace.Snapshot) []engine.Violation {
	sigs, ok := s.Interfaces()
	if !ok {
		return nil
	}
	var out []engine.Violation
	for _, e := range archEntities(s, "IF") {
		sig := sigs[e.ID]
		block := signatureBlock(e)
		if sig == "" || block == nil {
			continue // 没有快照由 if.01、if.02 报告，没有签名代码块由 arch.01 报告
		}
		m := declNameRe.FindStringSubmatch(sig)
		if m == nil {
			continue
		}
		text, line := designDecl(block, m[1], m[2])
		if text == "" {
			out = append(out, engine.Violation{File: e.File, Line: block.Line, Msg: fmt.Sprintf("%s 的设计签名中没有 %s 的声明；代码签名为 %s", e.ID, m[2], sig)})
			continue
		}
		design := normalizeGo(e.ID, text)
		if design != sig {
			if design == "" {
				design = "（无法按 Go 解析）"
			}
			out = append(out, engine.Violation{File: e.File, Line: line, Msg: fmt.Sprintf("%s 的设计签名与代码不一致：设计为 %s，代码为 %s", e.ID, design, sig)})
		}
	}
	return out
}

// signatureBlock 返回首行为 "// IF-n" 的代码块。
func signatureBlock(e *model.Entity) *model.CodeBlock {
	for i, c := range e.Code {
		first, _, _ := strings.Cut(strings.TrimSpace(c.Text), "\n")
		if strings.TrimSpace(first) == "// "+e.ID {
			return &e.Code[i]
		}
	}
	return nil
}

// designDecl 取出代码块中名为 name 的顶层声明（接收者为 recv）：从声明行起，到下一行不缩进的内容为止，
// 不缩进的 } 或 ) 算作声明的结尾。line 为声明在文档中的行号。
func designDecl(c *model.CodeBlock, recv, name string) (text string, line int) {
	lines := strings.Split(c.Text, "\n")
	for i, l := range lines {
		m := declNameRe.FindStringSubmatch(l)
		if m == nil || m[2] != name || m[1] != recv {
			continue
		}
		j := i + 1
		for j < len(lines) && (strings.HasPrefix(lines[j], "\t") || strings.HasPrefix(lines[j], " ")) {
			j++
		}
		if j < len(lines) && (strings.HasPrefix(lines[j], "}") || strings.HasPrefix(lines[j], ")")) {
			j++
		}
		return strings.Join(lines[i:j], "\n"), c.Line + 1 + i
	}
	return "", 0
}

// normalizeGo 用 Go 分析器规范化一段声明；无法解析时返回空串。
func normalizeGo(id, decl string) string {
	a, ok := lang.AnalyzerFor("design.go")
	if !ok {
		return ""
	}
	ds := a.Anchors("design.go", []byte("package design\n\n// harness:"+id+"\n"+decl+"\n"))
	if len(ds) != 1 {
		return ""
	}
	return ds[0].Signature
}
