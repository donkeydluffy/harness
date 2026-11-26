// Package typescript 是 TypeScript/JavaScript 语言包。
package typescript

import (
	"path"
	"regexp"
	"strings"

	"harness/internal/lang/spec"
	"harness/internal/lang/textscan"
)

// Lang 实现 spec.Language。
type Lang struct{}

var syntax = textscan.Syntax{Line: []string{"//"}, Block: [][2]string{{"/*", "*/"}}, Quotes: []string{`"`, "'"}, RawQuotes: []string{"`"}}

func (Lang) Name() string { return "typescript" }
func (Lang) Extensions() []string {
	return []string{".ts", ".tsx", ".mts", ".cts", ".js", ".jsx", ".mjs", ".cjs"}
}
func (Lang) Comments(src []byte) []spec.Comment { return syntax.Comments(src) }

var testFile = regexp.MustCompile(`\.(test|spec)\.[cm]?[jt]sx?$`)

// TestKind：*.test.ts、*.spec.js 等为测试文件；文件名含 integration（如 a.integration.test.ts）
// 或位于 integration、e2e 目录的为集成测试。
func (Lang) TestKind(p string, _ []byte) spec.TestKind {
	return textscan.Kind(p, testFile.MatchString(path.Base(p)), false)
}

var testCall = regexp.MustCompile(`\b(?:test|it)(?:\.only|\.skip)?\s*\(\s*(?:"((?:\\.|[^"\\])*)"|'((?:\\.|[^'\\])*)'|` + "`([^`]*)`)")

// TestNames 返回 test("…") 与 it("…") 的名称。
func (Lang) TestNames(src []byte) []string {
	var out []string
	for _, m := range testCall.FindAllStringSubmatch(string(src), -1) {
		out = append(out, strings.TrimSpace(m[1]+m[2]+m[3]))
	}
	return out
}
