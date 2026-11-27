package toolrun

import (
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Unit 对每个带锚点的接口单独运行测试名含其编号的单元测试，记录锚点所在函数的入口行是否执行；
// 再全量运行一次单元测试统计各模块行覆盖率；结果与源码哈希写入 .harness/reports/unit.json（AC-021-2、AC-021-3）。
// 有源文件却没有 Analyzer 的语言记入 Skipped。快照须以 WithSources 加载。
//
// harness:IF-031
func Unit(s *workspace.Snapshot) (*testreport.UnitReport, error) {
	rep := &testreport.UnitReport{SourceHash: s.SourceHash(), Interfaces: map[string]testreport.IFCoverage{},
		Modules: map[string]float64{}, Skipped: []string{}}
	analyzers := map[string]lang.Analyzer{} // 语言名 -> Analyzer
	for _, name := range s.Config.Languages {
		if !hasSources(s, name) {
			continue
		}
		if a, ok := analyzerOf(s, name); ok {
			analyzers[name] = a
		} else {
			rep.Skipped = append(rep.Skipped, name)
		}
	}
	design := map[string]bool{}
	for _, e := range s.Project.ByKind["IF"] {
		design[e.ID] = true
	}
	for _, d := range s.Anchors() {
		a, ok := lang.AnalyzerFor(d.File)
		if _, seen := rep.Interfaces[d.ID]; !ok || seen || !design[d.ID] || d.Signature == "" {
			continue
		}
		passed, tests, executed, _, err := a.Coverage(s.Root, testPattern(d.ID))
		if err != nil {
			return nil, err
		}
		rep.Interfaces[d.ID] = testreport.IFCoverage{Tests: nonNil(tests), Passed: passed && len(tests) > 0,
			EntryHit: d.EntryLine > 0 && contains(executed[d.File], d.EntryLine)}
	}
	names := make([]string, 0, len(analyzers))
	for name := range analyzers {
		names = append(names, name)
	}
	sort.Strings(names)
	hit, total := map[string]int{}, map[string]int{} // 模块编号 -> 行数
	for _, name := range names {
		_, _, executed, statements, err := analyzers[name].Coverage(s.Root, "")
		if err != nil {
			return nil, err
		}
		for _, m := range s.Project.ByKind["MOD"] {
			for file, lines := range statements {
				if !workspace.MatchAny(m.List("Paths"), file) {
					continue
				}
				total[m.ID] += len(lines)
				for _, n := range lines {
					if contains(executed[file], n) {
						hit[m.ID]++
					}
				}
			}
		}
	}
	for id, n := range total {
		if n > 0 {
			rep.Modules[id] = math.Floor(float64(hit[id])*1000/float64(n)) / 10 // 保留一位小数，向下取整
		}
	}
	return rep, rep.Write(filepath.Join(s.Root, testreport.UnitFile))
}

// testPattern 返回匹配测试名含接口编号的正则：IF-002 匹配 TestSave_IF_002、TestSave_IF_002_Empty，不匹配 IF_0021。
func testPattern(id string) string {
	return strings.ReplaceAll(id, "-", "_") + "([^0-9]|$)"
}

// analyzerOf 返回该语言源文件的 Analyzer。
func analyzerOf(s *workspace.Snapshot, name string) (lang.Analyzer, bool) {
	for _, f := range s.Sources {
		if f.Lang != nil && f.Lang.Name() == name {
			if a, ok := lang.AnalyzerFor(f.Path); ok {
				return a, true
			}
		}
	}
	return nil, false
}

func contains(sorted []int, n int) bool {
	i := sort.SearchInts(sorted, n)
	return i < len(sorted) && sorted[i] == n
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
