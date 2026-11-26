package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"harness/internal/phase"
	"harness/internal/rules/builtin"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

var registerOnce sync.Once

func registerRules() {
	registerOnce.Do(func() {
		for _, r := range builtin.Rules() {
			engine.Register(r)
		}
		engine.Register(phase.Rule())
	})
}

// frontmatter 返回 SKILL.md 开头 --- 之间的 key: value。
func frontmatter(src string) map[string]string {
	out := map[string]string{}
	if !strings.HasPrefix(src, "---\n") {
		return out
	}
	body, _, ok := strings.Cut(src[4:], "\n---\n")
	if !ok {
		return out
	}
	for _, l := range strings.Split(body, "\n") {
		if k, v, ok := strings.Cut(l, ":"); ok {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return out
}

// skillProblems 检查每个 skill 声明 name、description 与结束前运行的 harness check 子命令，
// 且子命令中的每个选择器至少选中一条已注册规则（AC-016-1）。
func skillProblems(fsys fs.FS) []string {
	registerRules()
	var out []string
	matches, _ := fs.Glob(fsys, "*/SKILL.md")
	if len(matches) == 0 {
		return []string{"没有任何 skill"}
	}
	for _, m := range matches {
		b, _ := fs.ReadFile(fsys, m)
		fm := frontmatter(string(b))
		if fm["name"] != path.Dir(m) || fm["description"] == "" {
			out = append(out, m+"：frontmatter 缺少 name（应为目录名）或 description")
		}
		check, ok := strings.CutPrefix(fm["check"], "harness check")
		if !ok {
			out = append(out, m+"：frontmatter 缺少 check: harness check …")
			continue
		}
		for _, sel := range strings.Fields(check) {
			if !isRule(sel) {
				out = append(out, fmt.Sprintf("%s：check 中的 %q 没有选中任何规则", m, sel))
			}
		}
	}
	return out
}

func isRule(sel string) bool {
	for _, id := range engine.IDs() {
		if id == sel || strings.HasPrefix(id, sel+".") {
			return true
		}
	}
	return false
}

func skillsFS(t *testing.T) fs.FS {
	t.Helper()
	sub, err := fs.Sub(templates, "templates/skills")
	if err != nil {
		t.Fatal(err)
	}
	return sub
}

func skill(t *testing.T, name string) string {
	t.Helper()
	b, err := fs.ReadFile(skillsFS(t), name+"/SKILL.md")
	if err != nil {
		t.Fatalf("缺少 %s skill：%v", name, err)
	}
	return string(b)
}

// ---- AC-016-1 skill 声明结束前的检查 ----

func TestSkills_AC_016_1_N1(t *testing.T) {
	if p := skillProblems(skillsFS(t)); len(p) > 0 {
		t.Fatal(strings.Join(p, "\n"))
	}
	// 每个内置阶段引用的 skill 都存在。
	phases, err := phase.Load(&workspace.Snapshot{Config: workspace.DefaultConfig()})
	if err != nil {
		t.Fatal(err)
	}
	for _, ph := range phases {
		skill(t, ph.Skill)
	}
}

func TestSkills_AC_016_1_E1(t *testing.T) {
	fsys := fstest.MapFS{
		"plan/SKILL.md":  {Data: []byte("---\nname: plan\ndescription: x\n---\n正文\n")},
		"build/SKILL.md": {Data: []byte("---\nname: build\ndescription: x\ncheck: harness check nosuch\n---\n")},
	}
	p := strings.Join(skillProblems(fsys), "\n")
	if !strings.Contains(p, "plan/SKILL.md：frontmatter 缺少 check") || !strings.Contains(p, `"nosuch" 没有选中任何规则`) {
		t.Fatalf("问题清单：%s", p)
	}
}

// ---- AC-016-2 测试 skill 的命名约定 ----

var exampleRow = regexp.MustCompile("(?m)^\\| (\\w+) \\| `([^`]+)` \\| `([^`]+)` \\|$")

func TestSkills_AC_016_2_N1(t *testing.T) {
	src := skill(t, "test")
	if !strings.Contains(src, "`AC-001-1/N1` 写作 `AC_001_1_N1`") {
		t.Fatal("测试 skill 缺少按用例编号命名测试的约定")
	}
	rows := exampleRow.FindAllStringSubmatch(src, -1)
	var langs []string
	for _, r := range rows {
		langs = append(langs, r[1])
	}
	if strings.Join(langs, ",") != "go,typescript,python,java,cpp" {
		t.Fatalf("测试 skill 的示例应覆盖全部语言：%v", langs)
	}
}

// testFile 按语言生成声明了测试 name 的集成测试文件内容。
var testFile = map[string]string{
	"go":         "package order\n\nimport \"testing\"\n\nfunc %s(t *testing.T) {}\n",
	"typescript": "test('%s', () => {});\n",
	"python":     "def %s():\n    pass\n",
	"java":       "class OrderIT {\n  @Test\n  void %s() {}\n}\n",
	"cpp":        "TEST(Order, %s) {}\n",
}

const examplePRD = "# PRD\n\n- Status: Draft\n- Current-Release: 1\n\n## 3. 功能需求\n\n### FR-001 下单\n- Status: Active\n- Release: 1\n- Priority: P0\n\n- AC-001-1: 创建订单。\n  - AC-001-1/N1: 正常下单\n"

// coverage 在临时项目中放入按示例命名的集成测试与通过的报告，返回 test.01、test.02 的违规。
func coverage(t *testing.T, lang, file, name string) []engine.Violation {
	t.Helper()
	registerRules()
	dir := t.TempDir()
	files := map[string]string{
		"docs/PRD.md":                examplePRD,
		file:                         fmt.Sprintf(testFile[lang], name),
		".harness/reports/junit.xml": `<testsuite name="s"><testcase name="` + name + `" classname="c"/></testsuite>`,
	}
	for rel, c := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(c), 0o644)
	}
	s, err := workspace.Load(dir, workspace.LoadOptions{WithSources: true, WithTests: true})
	if err != nil {
		t.Fatal(err)
	}
	return engine.Run(s, engine.Selector{IDs: []string{"test.01", "test.02"}}).Violations
}

func TestSkills_AC_016_2_E1(t *testing.T) {
	for _, r := range exampleRow.FindAllStringSubmatch(skill(t, "test"), -1) {
		lang, file, name := r[1], r[2], r[3]
		if vs := coverage(t, lang, file, name); len(vs) > 0 {
			t.Errorf("%s 示例 %s（%s）与 harness check test 的提取规则不一致：%+v", lang, name, file, vs)
		}
	}
	// 对照：不按约定命名时同一检查失败，说明上面的比对有效。
	if vs := coverage(t, "go", "internal/order/order_integration_test.go", "TestCreate_AC0011N1"); len(vs) == 0 {
		t.Fatal("不按约定命名的测试也被计入覆盖")
	}
}

// ---- AC-011-5 规划 skill 遇到不清楚的需求时停止 ----

func TestSkills_AC_011_5_N1(t *testing.T) {
	if !strings.Contains(skill(t, "plan"), "需求或设计不清时登记待澄清项并停止规划") {
		t.Fatal("规划 skill 缺少\"需求或设计不清时登记待澄清项并停止规划\"的指令")
	}
}

// 需求澄清 skill 必须询问未来的变化（AC-033-5）。
func TestSkills_ClarifyAsksChange(t *testing.T) {
	src := skill(t, "clarify")
	if !strings.Contains(src, "未来哪些东西会变化") || !strings.Contains(src, "不预留扩展") {
		t.Fatal("需求澄清 skill 缺少\"未来哪些东西会变化\"的问题")
	}
}
