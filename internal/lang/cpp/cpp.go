// Package cpp 是 C/C++ 语言包。
package cpp

import (
	"path"
	"regexp"

	"harness/internal/lang/spec"
	"harness/internal/lang/textscan"
)

// Lang 实现 spec.Language。
type Lang struct{}

var syntax = textscan.Syntax{Line: []string{"//"}, Block: [][2]string{{"/*", "*/"}}, Quotes: []string{`"`, "'"}}

func (Lang) Name() string { return "cpp" }
func (Lang) Extensions() []string {
	return []string{".c", ".h", ".cc", ".cpp", ".cxx", ".hpp", ".hh", ".hxx"}
}
func (Lang) Comments(src []byte) []spec.Comment { return syntax.Comments(src) }

var testFile = regexp.MustCompile(`(^test_.*|_test|_unittest|_tests?)\.(c|cc|cpp|cxx)$`)

// TestKind：test_*.c、*_test.cc、*_unittest.cpp 等为测试文件；位于 integration 目录或文件名含 integration 的为集成测试。
func (Lang) TestKind(p string, _ []byte) spec.TestKind {
	return textscan.Kind(p, testFile.MatchString(path.Base(p)), false)
}

var testMacro = regexp.MustCompile(`(?m)^\s*(?:TEST|TEST_F|TEST_P)\s*\(\s*\w+\s*,\s*(\w+)\s*\)`)

// TestNames 返回 GoogleTest 宏 TEST(套件, 名称) 中的名称（JUnit 报告中的 testcase name）。
func (Lang) TestNames(src []byte) []string {
	var out []string
	for _, m := range testMacro.FindAllStringSubmatch(string(src), -1) {
		out = append(out, m[1])
	}
	return out
}
