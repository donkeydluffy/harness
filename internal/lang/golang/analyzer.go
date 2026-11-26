package golang

import (
	"bufio"
	"bytes"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang/spec"
)

// Imports 返回源码中的 import 路径与行号；无法解析时返回 nil。
func (Lang) Imports(src []byte) []spec.Import {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ImportsOnly)
	if err != nil {
		return nil
	}
	var out []spec.Import
	for _, im := range f.Imports {
		p, _ := strconv.Unquote(im.Path.Value)
		out = append(out, spec.Import{Path: p, Line: fset.Position(im.Pos()).Line})
	}
	return out
}

var moduleRe = regexp.MustCompile(`^module\s+("?)(\S+?)("?)\s*(//.*)?$`)

// ModuleDecl 读取 go.mod 的 module 行。
func (Lang) ModuleDecl(p string, src []byte) (string, int, bool) {
	if path.Base(filepath.ToSlash(p)) != "go.mod" {
		return "", 0, false
	}
	sc := bufio.NewScanner(bytes.NewReader(src))
	for n := 1; sc.Scan(); n++ {
		if m := moduleRe.FindStringSubmatch(strings.TrimSpace(sc.Text())); m != nil {
			return m[2], n, true
		}
	}
	return "", 0, false
}

var anchorRe = regexp.MustCompile(`^//\s*harness:(IF-\d+)\s*$`)

// Anchors 返回锚点注释紧随的顶层声明：锚点的下一行须是声明的首行（func 或 type、var、const 关键字所在行）。
func (Lang) Anchors(p string, src []byte) []spec.Decl {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return nil
	}
	byLine := map[int]ast.Decl{}
	for _, d := range f.Decls {
		byLine[fset.Position(d.Pos()).Line] = d
	}
	var out []spec.Decl
	for _, g := range f.Comments {
		for _, c := range g.List {
			m := anchorRe.FindStringSubmatch(c.Text)
			if m == nil {
				continue
			}
			line := fset.Position(c.Pos()).Line
			dc := spec.Decl{ID: m[1], File: p, Line: line}
			if d, ok := byLine[line+1]; ok {
				start, end := fset.Position(d.Pos()).Offset, fset.Position(d.End()).Offset
				if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil {
					end = fset.Position(fn.Body.Lbrace).Offset
					dc.EntryLine = fset.Position(fn.Body.Lbrace).Line
				}
				dc.Signature = Normalize(src[start:end])
			}
			out = append(out, dc)
		}
	}
	return out
}

// Normalize 把一段 Go 声明规范化为一行：去掉注释与结构体字段标签，按词法单元重新拼接空白，去掉右括号前多余的分号与逗号。
// 只改空白、注释与换行的两段声明规范化后相同（AC-052-2/B2、AC-052-3/B1）。
func Normalize(src []byte) string {
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	var s scanner.Scanner
	s.Init(file, src, nil, 0)
	type tk struct {
		tok token.Token
		lit string
		gap bool // 源码中与前一个词法单元之间有空白
	}
	var toks []tk
	var structs []bool // 每层 { 是否为结构体
	end := 0
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		off := file.Offset(pos)
		gap := off > end
		if tok == token.SEMICOLON && lit == "\n" {
			lit = ";"
		} else {
			if !tok.IsLiteral() {
				lit = tok.String()
			}
			end = off + len(lit)
		}
		switch tok {
		case token.LBRACE:
			structs = append(structs, len(toks) > 0 && toks[len(toks)-1].tok == token.STRUCT)
		case token.RBRACE:
			if n := len(structs); n > 0 {
				structs = structs[:n-1]
			}
		case token.STRING:
			if n := len(structs); n > 0 && structs[n-1] {
				continue // 结构体字段标签
			}
		}
		closes := tok == token.RPAREN || tok == token.RBRACE || tok == token.RBRACK
		for closes && len(toks) > 0 && (toks[len(toks)-1].tok == token.SEMICOLON || toks[len(toks)-1].tok == token.COMMA) {
			toks = toks[:len(toks)-1]
		}
		if tok == token.SEMICOLON && (len(toks) == 0 || toks[len(toks)-1].tok == token.SEMICOLON || toks[len(toks)-1].tok == token.LBRACE) {
			continue
		}
		toks = append(toks, tk{tok, lit, gap})
	}
	for len(toks) > 0 && toks[len(toks)-1].tok == token.SEMICOLON {
		toks = toks[:len(toks)-1]
	}
	var b strings.Builder
	var brackets []bool // 每个 [ 是否紧跟标识符（类型参数或实例化）
	identBracket := false
	for i, t := range toks {
		indexed := t.tok == token.LBRACK && i > 0 && toks[i-1].tok == token.IDENT && !t.gap
		if i > 0 && space(toks[i-1].tok, t.tok, i-1 == 0, identBracket, indexed) {
			b.WriteByte(' ')
		}
		identBracket = false
		switch t.tok {
		case token.LBRACK:
			brackets = append(brackets, indexed)
		case token.RBRACK:
			if n := len(brackets); n > 0 {
				identBracket, brackets = brackets[n-1], brackets[:n-1]
			}
		}
		b.WriteString(t.lit)
	}
	return b.String()
}

// space 判断两个相邻词法单元之间是否写空格；first 表示 prev 是声明的第一个词法单元，
// afterIdentBracket 表示 prev 是紧跟标识符的 [ 对应的 ]，indexed 表示 cur 是紧跟标识符的 [。
func space(prev, cur token.Token, first, afterIdentBracket, indexed bool) bool {
	switch prev {
	case token.LPAREN, token.LBRACK, token.PERIOD, token.MUL, token.ELLIPSIS, token.ARROW:
		return false
	case token.RBRACK:
		if !afterIdentBracket {
			return false
		}
	case token.LBRACE:
		return cur != token.RBRACE
	}
	switch cur {
	case token.RPAREN, token.RBRACK, token.COMMA, token.SEMICOLON, token.PERIOD:
		return false
	case token.LPAREN:
		return !(prev == token.IDENT || (prev == token.FUNC && !first) || (prev == token.RBRACK && afterIdentBracket))
	case token.LBRACK:
		return !(indexed || prev == token.MAP)
	}
	return true
}

// Resolve 按 go.mod 的模块路径把 import 路径解析为仓库相对目录。
func (l Lang) Resolve(root, importPath string) (string, bool) {
	src, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", false
	}
	mod, _, ok := l.ModuleDecl("go.mod", src)
	switch {
	case !ok:
		return "", false
	case importPath == mod:
		return ".", true
	case strings.HasPrefix(importPath, mod+"/"):
		return strings.TrimPrefix(importPath, mod+"/"), true
	}
	return "", false
}

// Coverage 运行 go test -coverprofile -coverpkg=./...，返回是否全部通过、运行过的测试名与已执行的行。
// pattern 为空时运行全部单元测试：名称含 _AC_ 的集成测试被跳过。
func (l Lang) Coverage(root, pattern string) (bool, []string, map[string][]int, error) {
	prof, err := os.CreateTemp("", "harness-cover-*.out")
	if err != nil {
		return false, nil, nil, err
	}
	prof.Close()
	defer os.Remove(prof.Name())
	args := []string{"test", "-count=1", "-json", "-coverprofile=" + prof.Name(), "-coverpkg=./..."}
	if pattern != "" {
		args = append(args, "-run", pattern)
	} else {
		args = append(args, "-skip", "_AC_")
	}
	cmd := exec.Command("go", append(args, "./...")...)
	cmd.Dir = root
	out, runErr := cmd.Output()
	if _, ok := runErr.(*exec.ExitError); runErr != nil && !ok {
		return false, nil, nil, runErr
	}
	seen := map[string]bool{}
	var tests []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		var ev struct{ Action, Test string }
		if json.Unmarshal(sc.Bytes(), &ev) == nil && ev.Test != "" && (ev.Action == "pass" || ev.Action == "fail") && !seen[ev.Test] {
			seen[ev.Test] = true
			tests = append(tests, ev.Test)
		}
	}
	sort.Strings(tests)
	profile, err := os.ReadFile(prof.Name())
	if err != nil {
		return false, tests, nil, err
	}
	executed := l.executed(root, profile)
	return runErr == nil, tests, executed, nil
}

// executed 解析覆盖率文件，返回执行次数大于 0 的语句块覆盖的行（仓库相对路径 -> 升序行号）。
func (l Lang) executed(root string, profile []byte) map[string][]int {
	lines := map[string]map[int]bool{}
	for _, row := range strings.Split(string(profile), "\n")[1:] {
		// <import 路径>/<文件>:<起始行>.<列>,<结束行>.<列> <语句数> <次数>
		colon := strings.LastIndex(row, ":")
		f := strings.Fields(row[colon+1:])
		if colon < 0 || len(f) != 3 || f[2] == "0" {
			continue
		}
		dir, ok := l.Resolve(root, path.Dir(row[:colon]))
		if !ok {
			continue
		}
		file := path.Join(dir, path.Base(row[:colon]))
		var a, b int
		ranges := strings.SplitN(f[0], ",", 2)
		a, _ = strconv.Atoi(strings.SplitN(ranges[0], ".", 2)[0])
		if len(ranges) == 2 {
			b, _ = strconv.Atoi(strings.SplitN(ranges[1], ".", 2)[0])
		}
		if lines[file] == nil {
			lines[file] = map[int]bool{}
		}
		for n := a; n <= b; n++ {
			lines[file][n] = true
		}
	}
	out := map[string][]int{}
	for file, set := range lines {
		for n := range set {
			out[file] = append(out[file], n)
		}
		sort.Ints(out[file])
	}
	return out
}
