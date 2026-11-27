package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// commit05：机械改动任务的提交只改 import 路径与模块声明行，且每处都是按 Rewrite 替换前缀（AC-050-3）。
// import 的先后顺序可以变化（如格式化工具重新排序）；产物文件与生成文件豁免。
type commit05 struct{}

func (commit05) Needs() engine.Input { return engine.InputGit }
func (commit05) ID() string          { return "commit.05" }
func (commit05) Covers() []string    { return []string{"AC-050-3"} }
func (commit05) Fix() string {
	return "机械改动任务的提交只做 Rewrite 声明的路径前缀替换；其他改动放到普通任务的提交中"
}

func (commit05) Check(s *workspace.Snapshot) []engine.Violation {
	sets, missing := changeSets(s)
	if missing != nil {
		return []engine.Violation{*missing}
	}
	var out []engine.Violation
	for _, cs := range sets {
		t := changeTask(s, cs)
		if !mechanical(t) {
			continue
		}
		rs, _ := rewrites(t)
		for _, f := range cs.Files {
			if exemptFile(f.Path) {
				continue
			}
			before, after := cs.Before(f.Path), cs.After(f.Path)
			switch {
			case before == nil:
				out = append(out, commitViolation(cs, f.Path, "机械改动任务 %s 的提交新增了文件 %s", t.ID, f.Path))
			case after == nil:
				out = append(out, commitViolation(cs, f.Path, "机械改动任务 %s 的提交删除了文件 %s", t.ID, f.Path))
			default:
				out = append(out, rewriteProblems(cs, t.ID, f.Path, rs, before, after)...)
			}
		}
	}
	return out
}

// rewriteProblems 比较文件改动前后的内容：import 行与模块声明行之外的行须逐行相同；
// 改动前的 import 行与模块声明行按 Rewrite 替换后，须与改动后的这些行一一对应（不计顺序）。
func rewriteProblems(cs changeSet, task, file string, rs []rewrite, before, after []byte) []engine.Violation {
	var out []engine.Violation
	bSpecial, bPlain := splitRewritable(file, before, rs)
	aSpecial, aPlain := splitRewritable(file, after, nil)
	for i := 0; i < len(aPlain) || i < len(bPlain); i++ {
		if i < len(aPlain) && i < len(bPlain) && aPlain[i].text == bPlain[i].text {
			continue
		}
		line := 1
		switch {
		case i < len(aPlain):
			line = aPlain[i].n
		case len(aPlain) > 0:
			line = aPlain[len(aPlain)-1].n
		}
		v := commitViolation(cs, file, "机械改动任务 %s 的提交改动了 %s 第 %d 行，它不是 import 路径或模块声明", task, file, line)
		v.Line = line
		return append(out, v)
	}
	want := map[string]int{}
	for _, l := range bSpecial {
		want[l.text]++
	}
	for _, l := range aSpecial {
		if want[l.text] > 0 {
			want[l.text]--
			continue
		}
		v := commitViolation(cs, file, "机械改动任务 %s 的提交中 %s 第 %d 行的改动不符合任何 Rewrite：%s", task, file, l.n, l.text)
		v.Line = l.n
		out = append(out, v)
	}
	return out
}

type srcLine struct {
	n    int
	text string
}

// splitRewritable 把文件分为 import 行与模块声明行（special）和其余行（plain）；
// rs 非空时把 special 行中的路径按 Rewrite 替换，得到期望的改动后内容。
func splitRewritable(file string, src []byte, rs []rewrite) (special, plain []srcLine) {
	paths := map[int][]string{}
	if a, ok := lang.AnalyzerFor(file); ok {
		for _, im := range a.Imports(src) {
			paths[im.Line] = append(paths[im.Line], im.Path)
		}
	}
	for _, l := range lang.All() {
		if a, ok := l.(lang.Analyzer); ok {
			if mod, n, ok := a.ModuleDecl(file, src); ok {
				paths[n] = append(paths[n], mod)
			}
		}
	}
	for i, text := range strings.Split(string(src), "\n") {
		n := i + 1
		ps, ok := paths[n]
		if !ok {
			plain = append(plain, srcLine{n, text})
			continue
		}
		text = strings.TrimSpace(text)
		for _, p := range ps {
			if q := applyRewrite(rs, p); q != p {
				text = replacePath(text, p, q)
			}
		}
		special = append(special, srcLine{n, text})
	}
	return special, plain
}

// replacePath 把行中作为整体出现的路径 p（带引号或以空白分隔）替换为 q。
func replacePath(text, p, q string) string {
	for _, quote := range []string{`"`, "`"} {
		if strings.Contains(text, quote+p+quote) {
			return strings.Replace(text, quote+p+quote, quote+q+quote, 1)
		}
	}
	f := strings.Fields(text)
	for i := range f {
		if f[i] == p {
			f[i] = q
		}
	}
	return strings.Join(f, " ")
}
