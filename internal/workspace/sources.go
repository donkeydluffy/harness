package workspace

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/testreport"
)

// SourceFile 是按 sourcePaths 收集的源码文件，内容惰性读取。
type SourceFile struct {
	Path string // 相对项目根目录，使用 /
	Lang lang.Language
	Kind lang.TestKind
	root string
	src  []byte
}

// Content 读取文件内容并缓存。
func (f *SourceFile) Content() ([]byte, error) {
	if f.src != nil {
		return f.src, nil
	}
	b, err := os.ReadFile(filepath.Join(f.root, filepath.FromSlash(f.Path)))
	f.src = b
	return b, err
}

// 未配置 sourcePaths 时遍历整个项目，跳过这些目录。
var skipDirs = map[string]bool{".git": true, ".harness": true, "node_modules": true, "vendor": true, "testdata": true}

func loadSources(root string, globs []string) ([]*SourceFile, error) {
	var out []*SourceFile
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel != "." && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if len(globs) > 0 && !MatchAny(globs, rel) {
			return nil
		}
		l, ok := lang.ForPath(rel)
		if !ok {
			return nil
		}
		f := &SourceFile{Path: rel, Lang: l, root: root}
		src, err := f.Content()
		if err != nil {
			return err
		}
		f.Kind = l.TestKind(rel, src)
		out = append(out, f)
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, err
}

func loadReports(root string, globs []string) (files []string, tests []testreport.TestCase, err error) {
	for _, g := range globs {
		matches, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(g)))
		sort.Strings(matches)
		for _, m := range matches {
			tcs, err := testreport.ParseFile(m)
			if err != nil {
				return nil, nil, err
			}
			rel, _ := filepath.Rel(root, m)
			files = append(files, filepath.ToSlash(rel))
			tests = append(tests, tcs...)
		}
	}
	return files, tests, nil
}

// MatchAny 判断相对路径是否匹配任一 glob。
func MatchAny(globs []string, rel string) bool {
	for _, g := range globs {
		if Match(g, rel) {
			return true
		}
	}
	return false
}

// Match 判断相对路径是否匹配 glob：** 匹配任意层目录，* 与 ? 不跨越 /。
func Match(pattern, rel string) bool {
	return matchParts(strings.Split(pattern, "/"), strings.Split(rel, "/"))
}

func matchParts(pat, name []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			if len(pat) == 1 {
				return true
			}
			for i := 0; i <= len(name); i++ {
				if matchParts(pat[1:], name[i:]) {
					return true
				}
			}
			return false
		}
		if len(name) == 0 {
			return false
		}
		if ok, _ := filepath.Match(pat[0], name[0]); !ok {
			return false
		}
		pat, name = pat[1:], name[1:]
	}
	return len(name) == 0
}
