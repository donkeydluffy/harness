package builtin

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// doc02：受检 Markdown 行内代码中的仓库相对路径必须存在（AC-030-1）。
type doc02 struct{}

func (doc02) ID() string       { return "doc.02" }
func (doc02) Covers() []string { return []string{"AC-030-1"} }
func (doc02) Fix() string {
	return "把路径改为现有文件，或删掉过期的引用；示意路径用 <…> 占位符或写进代码块"
}

func (doc02) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	var names map[string]bool // 仓库中全部文件名，惰性收集
	for _, l := range s.DocLines() {
		for _, c := range l.InlineCode() {
			p, ok := pathRef(c)
			if !ok || exists(s.Root, p) || exists(s.Root, path.Join(path.Dir(l.File), p)) {
				continue
			}
			if first, _, slash := strings.Cut(p, "/"); slash && !exists(s.Root, first) && !exists(s.Root, path.Join(path.Dir(l.File), first)) {
				continue // 首段不在仓库中：标准库包（go/parser）、其他项目的目录等，不是本仓库的路径
			}
			if !strings.Contains(p, "/") {
				if names == nil {
					names = fileNames(s.Root)
				}
				if names[p] {
					continue // 只写了文件名，如 SKILL.md
				}
			}
			out = append(out, engine.Violation{File: l.File, Line: l.Line, Msg: fmt.Sprintf("引用的路径 %s 不存在", p)})
		}
	}
	return out
}

var (
	pathExtRe  = regexp.MustCompile(`\.(md|go|mod|sum|json|jsonl|txt|yml|yaml|toml|sh|py|ts|tsx|js|java|kt|kts|cpp|cc|h|hpp|gradle|xml|lock|gitignore)$`)
	lineSufRe  = regexp.MustCompile(`:\d+(:\d+)?$`)
	refTokenRe = regexp.MustCompile(`\b(FR|NFR|AC|IF|MOD|FLOW|ADR|T|DEBT|EXT|OQ|CHG)-N?\d+`)
)

// pathRef 判断行内代码是否是仓库相对路径：含 `/` 或常见文件扩展名，不含空白、通配符、`<…>` 占位符，
// 也不是 URL、绝对路径、模块导入路径（首段为域名）、编号、文件名后缀（_test.go、.api.md）或含中文的文字。
// 返回去掉锚点、行号与末尾 `/` 的路径。
func pathRef(c string) (string, bool) {
	for _, r := range c {
		if r > 0x2e7f {
			return "", false
		}
	}
	if strings.ContainsAny(c, " \t*?[]{}<>|$\"'()=,;@") || strings.Contains(c, "://") || strings.Contains(c, "..") ||
		strings.HasPrefix(c, "/") || strings.HasPrefix(c, "~") || strings.HasPrefix(c, "-") || refTokenRe.MatchString(c) {
		return "", false
	}
	p, _, _ := strings.Cut(c, "#")
	p = strings.TrimSuffix(lineSufRe.ReplaceAllString(p, ""), "/")
	if p == "" || p == "." || p == ".harness" || strings.HasPrefix(p, ".harness/") {
		return "", false // .harness 下是 harness 运行时生成的文件
	}
	if first, _, ok := strings.Cut(p, "/"); ok {
		if strings.Contains(first, ".") && !strings.HasPrefix(first, ".") {
			return "", false // github.com/x/y 之类的导入路径
		}
		return p, true
	}
	if strings.HasPrefix(p, ".") || strings.HasPrefix(p, "_") {
		return "", false
	}
	return p, pathExtRe.MatchString(p)
}

func exists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

// fileNames 返回仓库中全部文件的文件名（不含 .git）。
func fileNames(root string) map[string]bool {
	out := map[string]bool{}
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if !d.IsDir() {
			out[d.Name()] = true
		}
		return nil
	})
	return out
}
