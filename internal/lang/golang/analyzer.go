package golang

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"io/fs"
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
					if len(fn.Body.List) > 0 { // 覆盖率的语句块从第一条语句开始
						dc.EntryLine = fset.Position(fn.Body.List[0].Pos()).Line
					}
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

// Coverage 运行 go test -coverprofile -coverpkg=./...，返回是否全部通过、运行过的测试名、已执行的行与含语句的行。
// pattern 不为空时只运行声明了匹配测试的包。pattern 为空时运行全部测试（含集成测试），并设置环境变量 HARNESS_COVERDIR：
// 测试以 go build -cover 构建子进程、并把子进程的 GOCOVERDIR 设为该目录时，子进程的覆盖率合并进结果（ADR-014）。
// go test -cover 会改写测试进程的 GOCOVERDIR，且不合并子进程的数据，所以另用一个变量传递目录。
func (l Lang) Coverage(root, pattern string) (bool, []string, map[string][]int, map[string][]int, error) {
	prof, err := os.CreateTemp("", "harness-cover-*.out")
	if err != nil {
		return false, nil, nil, nil, err
	}
	prof.Close()
	defer os.Remove(prof.Name())
	args := []string{"test", "-count=1", "-json", "-coverprofile=" + prof.Name(), "-coverpkg=./..."}
	pkgs := []string{"./..."}
	if pattern != "" {
		args = append(args, "-run", pattern)
		if pkgs = testPackages(root, pattern); len(pkgs) == 0 {
			return false, nil, map[string][]int{}, map[string][]int{}, nil
		}
	}
	cmd := exec.Command("go", append(args, pkgs...)...)
	cmd.Dir = root
	var covdir string
	if pattern == "" {
		if covdir, err = os.MkdirTemp("", "harness-covdata-"); err != nil {
			return false, nil, nil, nil, err
		}
		defer os.RemoveAll(covdir)
		cmd.Env = append(os.Environ(), "HARNESS_COVERDIR="+covdir)
	}
	out, runErr := cmd.Output()
	if _, ok := runErr.(*exec.ExitError); runErr != nil && !ok {
		return false, nil, nil, nil, runErr
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
		return false, tests, nil, nil, err
	}
	if covdir != "" {
		sub, err := subprocessProfile(root, covdir)
		if err != nil {
			return false, tests, nil, nil, err
		}
		profile = append(append(profile, '\n'), sub...)
	}
	executed, statements := l.lines(root, profile)
	return runErr == nil, tests, executed, statements, nil
}

// lines 解析覆盖率文件，返回执行次数大于 0 的语句块覆盖的行与语句数大于 0 的语句块覆盖的行（仓库相对路径 -> 升序行号）。
// 空函数体是语句数为 0 的块，执行过时计入 executed，供判断入口行。
func (l Lang) lines(root string, profile []byte) (map[string][]int, map[string][]int) {
	hit, stmt := map[string]map[int]bool{}, map[string]map[int]bool{}
	for _, row := range strings.Split(string(profile), "\n")[1:] {
		// <import 路径>/<文件>:<起始行>.<列>,<结束行>.<列> <语句数> <次数>
		colon := strings.LastIndex(row, ":")
		f := strings.Fields(row[colon+1:])
		if colon < 0 || len(f) != 3 {
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
			end := strings.SplitN(ranges[1], ".", 2)
			b, _ = strconv.Atoi(end[0])
			if len(end) == 2 && end[1] == "1" && b > a {
				b-- // 块在该行第 1 列之前结束（如右括号所在行），该行不属于块
			}
		}
		mark(hit, file, a, b, f[2] != "0")
		mark(stmt, file, a, b, f[1] != "0")
	}
	return sorted(hit), sorted(stmt)
}

func mark(m map[string]map[int]bool, file string, a, b int, ok bool) {
	if !ok {
		return
	}
	if m[file] == nil {
		m[file] = map[int]bool{}
	}
	for n := a; n <= b; n++ {
		m[file][n] = true
	}
}

func sorted(m map[string]map[int]bool) map[string][]int {
	out := map[string][]int{}
	for file, set := range m {
		for n := range set {
			out[file] = append(out[file], n)
		}
		sort.Ints(out[file])
	}
	return out
}

// testPackages 返回声明了名称匹配 pattern 的测试的包（"./" 开头的相对目录，升序），只运行这些包以缩短耗时。
// pattern 不是合法正则时返回 ./...，交给 go test 报错。
func testPackages(root, pattern string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return []string{"./..."}
	}
	dirs := map[string]bool{}
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if p != root && (strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata" || name == "vendor") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(name, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		for _, t := range (Lang{}).TestNames(src) {
			if re.MatchString(t) {
				rel, _ := filepath.Rel(root, filepath.Dir(p))
				dirs["./"+filepath.ToSlash(rel)] = true
				break
			}
		}
		return nil
	})
	out := make([]string, 0, len(dirs))
	for d := range dirs {
		out = append(out, d)
	}
	sort.Strings(out)
	return out
}

// subprocessProfile 把 HARNESS_COVERDIR 中子进程写入的覆盖率数据转换为覆盖率文件格式；目录为空时返回 nil。
func subprocessProfile(root, covdir string) ([]byte, error) {
	if entries, _ := os.ReadDir(covdir); len(entries) == 0 {
		return nil, nil
	}
	out := filepath.Join(covdir, "profile.txt")
	cmd := exec.Command("go", "tool", "covdata", "textfmt", "-i="+covdir, "-o="+out)
	cmd.Dir = root
	if b, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("go tool covdata textfmt：%v：%s", err, b)
	}
	return os.ReadFile(out)
}
