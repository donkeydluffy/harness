// Package golang 是 Go 语言包：注释提取、测试分类、测试名提取。
package golang

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"path"
	"strings"

	"harness/internal/lang/spec"
)

// Lang 实现 spec.Language。
type Lang struct{}

func (Lang) Name() string         { return "go" }
func (Lang) Extensions() []string { return []string{".go"} }

// Comments 用 go/scanner 提取注释，字符串字面量中的内容不会被识别为注释。
func (Lang) Comments(src []byte) []spec.Comment {
	fset := token.NewFileSet()
	f := fset.AddFile("", fset.Base(), len(src))
	var s scanner.Scanner
	s.Init(f, src, nil, scanner.ScanComments)
	var out []spec.Comment
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			return out
		}
		if tok == token.COMMENT {
			out = append(out, spec.Comment{Text: lit, Line: fset.Position(pos).Line})
		}
	}
}

// TestKind：*_test.go 为测试文件；文件名以 _integration_test.go 结尾或带 //go:build integration 约束的为集成测试。
func (Lang) TestKind(p string, src []byte) spec.TestKind {
	base := path.Base(p)
	if !strings.HasSuffix(base, "_test.go") {
		return spec.NotTest
	}
	if strings.HasSuffix(base, "_integration_test.go") || hasIntegrationTag(src) {
		return spec.Integration
	}
	return spec.Unit
}

func hasIntegrationTag(src []byte) bool {
	for _, l := range bytes.Split(src, []byte("\n")) {
		t := strings.TrimSpace(string(l))
		if strings.HasPrefix(t, "package ") {
			return false
		}
		if strings.HasPrefix(t, "//go:build ") && strings.Contains(" "+t[len("//go:build "):]+" ", " integration ") {
			return true
		}
	}
	return false
}

// TestNames 返回顶层测试函数名：TestXxx(t *testing.T)。
func (Lang) TestNames(src []byte) []string {
	f, err := parser.ParseFile(token.NewFileSet(), "", src, parser.SkipObjectResolution)
	if err != nil {
		return nil
	}
	var out []string
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") || fn.Name.Name == "TestMain" {
			continue
		}
		if fn.Type.Params != nil && len(fn.Type.Params.List) == 1 {
			out = append(out, fn.Name.Name)
		}
	}
	return out
}
