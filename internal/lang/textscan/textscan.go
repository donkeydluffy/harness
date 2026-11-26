// Package textscan 是非 Go 语言包共用的词法扫描：按语法描述提取注释（跳过字符串字面量），
// 以及按路径约定判定测试类型。
package textscan

import (
	"path"
	"strings"

	"harness/internal/lang/spec"
)

// Syntax 描述一种语言的注释与字符串语法。
type Syntax struct {
	Line      []string    // 行注释前缀，如 "//"、"#"
	Block     [][2]string // 块注释起止，如 {"/*", "*/"}
	Quotes    []string    // 字符串定界符，按长度从长到短，如 `"""`、`"`、`'`
	RawQuotes []string    // 可跨行的定界符，如 JavaScript 模板字符串 "`"
}

// Comments 返回 src 中的注释及起始行号；字符串字面量中的注释符号不计入。
func (sx Syntax) Comments(src []byte) []spec.Comment {
	s := string(src)
	var out []spec.Comment
	line := 1
	for i := 0; i < len(s); {
		if s[i] == '\n' {
			line++
			i++
			continue
		}
		if q := prefix(s[i:], sx.Quotes, sx.RawQuotes); q != "" {
			multiline := len(q) > 1 || contains(sx.RawQuotes, q)
			j := i + len(q)
			for j < len(s) && !strings.HasPrefix(s[j:], q) {
				if s[j] == '\\' {
					j++
				}
				if j < len(s) && s[j] == '\n' {
					line++
					if !multiline { // 单行字符串不跨行，避免未闭合时吞掉后文
						break
					}
				}
				j++
			}
			i = j + len(q)
			continue
		}
		if p := prefix(s[i:], sx.Line, nil); p != "" {
			end := strings.IndexByte(s[i:], '\n')
			if end < 0 {
				end = len(s) - i
			}
			out = append(out, spec.Comment{Text: s[i : i+end], Line: line})
			i += end
			continue
		}
		matched := false
		for _, b := range sx.Block {
			if strings.HasPrefix(s[i:], b[0]) {
				end := strings.Index(s[i+len(b[0]):], b[1])
				stop := len(s)
				if end >= 0 {
					stop = i + len(b[0]) + end + len(b[1])
				}
				text := s[i:stop]
				out = append(out, spec.Comment{Text: text, Line: line})
				line += strings.Count(text, "\n")
				i, matched = stop, true
				break
			}
		}
		if !matched {
			i++
		}
	}
	return out
}

func prefix(s string, a, b []string) string {
	for _, list := range [][]string{a, b} {
		for _, p := range list {
			if strings.HasPrefix(s, p) {
				return p
			}
		}
	}
	return ""
}

func contains(list []string, x string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}

// IntegrationPath 判断测试文件按路径约定属于集成测试：目录名或文件名含 integration，
// 或位于 it、e2e 目录。
func IntegrationPath(p string) bool {
	for _, seg := range strings.Split(strings.ToLower(p), "/") {
		if strings.Contains(seg, "integration") || seg == "it" || seg == "e2e" {
			return true
		}
	}
	return false
}

// Kind 根据是否为测试文件与路径约定返回测试类型。
func Kind(p string, isTest, integration bool) spec.TestKind {
	switch {
	case !isTest:
		return spec.NotTest
	case integration || IntegrationPath(path.Dir(p)) || strings.Contains(strings.ToLower(path.Base(p)), "integration"):
		return spec.Integration
	default:
		return spec.Unit
	}
}
