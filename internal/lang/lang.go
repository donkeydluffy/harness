// Package lang 是语言注册表：文件归属、注释提取、测试分类（MOD-14）。新增语言是扩展点 EXT-01。
package lang

import (
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/donkeydluffy/harness/internal/lang/spec"
)

// 类型别名：语言实现依赖 spec，调用方使用 lang。
type (
	Language = spec.Language
	Comment  = spec.Comment
	TestKind = spec.TestKind
	Tool     = spec.Tool
	Pack     = spec.Pack
	Packer   = spec.Packer
	Analyzer = spec.Analyzer
	Import   = spec.Import
	Decl     = spec.Decl
)

const (
	NotTest     = spec.NotTest
	Unit        = spec.Unit
	Integration = spec.Integration
)

var (
	mu    sync.RWMutex
	langs = map[string]Language{}
)

// Register 注册语言；名称重复时 panic（IF-004）。
//
// harness:IF-004
func Register(l Language) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := langs[l.Name()]; dup {
		panic("lang: 语言重复注册 " + l.Name())
	}
	langs[l.Name()] = l
}

// ForPath 按扩展名返回文件所属的语言（IF-004）。
func ForPath(p string) (Language, bool) {
	ext := strings.ToLower(path.Ext(p))
	for _, l := range All() {
		for _, e := range l.Extensions() {
			if e == ext {
				return l, true
			}
		}
	}
	return nil, false
}

// All 返回已注册的语言，按名称排序（IF-004）。
func All() []Language {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]Language, 0, len(langs))
	for _, l := range langs {
		out = append(out, l)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// PackFor 返回语言的配置包；语言未注册或没有配置包时 ok 为 false（IF-034）。
//
// harness:IF-034
func PackFor(name string) (Pack, bool) {
	mu.RLock()
	l, ok := langs[name]
	mu.RUnlock()
	if !ok {
		return Pack{}, false
	}
	p, ok := l.(Packer)
	if !ok {
		return Pack{}, false
	}
	return p.Pack(), true
}

// AnalyzerFor 返回文件所属语言的源码分析器；语言未注册或没有分析器时 ok 为 false（IF-035）。
func AnalyzerFor(p string) (Analyzer, bool) {
	l, ok := ForPath(p)
	if !ok {
		return nil, false
	}
	a, ok := l.(Analyzer)
	return a, ok
}
