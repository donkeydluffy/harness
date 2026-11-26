// Package python 是 Python 语言包。
package python

import (
	"path"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang/spec"
	"github.com/donkeydluffy/harness/internal/lang/textscan"
)

// Lang 实现 spec.Language。
type Lang struct{}

var syntax = textscan.Syntax{Line: []string{"#"}, Quotes: []string{`"""`, "'''", `"`, "'"}}

func (Lang) Name() string                       { return "python" }
func (Lang) Extensions() []string               { return []string{".py"} }
func (Lang) Comments(src []byte) []spec.Comment { return syntax.Comments(src) }

// TestKind：test_*.py 与 *_test.py 为测试文件；位于 integration 目录或文件名含 integration 的为集成测试。
func (Lang) TestKind(p string, _ []byte) spec.TestKind {
	b := path.Base(p)
	return textscan.Kind(p, strings.HasPrefix(b, "test_") || strings.HasSuffix(b, "_test.py"), false)
}

var testDef = regexp.MustCompile(`(?m)^\s*(?:async\s+)?def\s+(test\w*)\s*\(`)

// TestNames 返回 test 开头的函数与方法名（pytest、unittest 约定）。
func (Lang) TestNames(src []byte) []string {
	var out []string
	for _, m := range testDef.FindAllStringSubmatch(string(src), -1) {
		out = append(out, m[1])
	}
	return out
}

// Pack 返回语言配置包（IF-034）。
func (Lang) Pack() spec.Pack {
	return spec.Pack{
		SourcePaths: []string{"src/**", "tests/**"},
		Format:      spec.Tool{Cmd: "ruff format --check .", Output: "files", Install: "pip install ruff"},
		Lint:        spec.Tool{Cmd: "ruff check --output-format concise .", Output: "lines", Install: "pip install ruff"},
	}
}
