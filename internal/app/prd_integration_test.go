package app_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/donkeydluffy/harness/internal/rules/builtin"
)

// 集成测试通过真实的 harness 二进制运行。TestMain 构建一次，供本包全部集成测试使用。
var harnessBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "harness-it-")
	if err != nil {
		panic(err)
	}
	harnessBin = filepath.Join(dir, "harness")
	if out, err := exec.Command("go", "build", "-o", harnessBin, "github.com/donkeydluffy/harness/cmd/harness").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "构建 harness 失败：%v\n%s", err, out)
		os.Exit(2)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// writeTree 在临时目录中写入文件，返回目录。键为相对路径。
func writeTree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, content := range files {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

type result struct {
	Stdout, Stderr string
	Code           int
}

// runHarness 在 dir 中执行 harness 命令。
func runHarness(t *testing.T, dir string, args ...string) result {
	t.Helper()
	cmd := exec.Command(harnessBin, args...)
	cmd.Dir = dir
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return result{out.String(), errb.String(), code}
}

type violation struct {
	Rule, File, Msg, Fix string
	Line                 int
}

// checked 是一次 harness check --json 的结果。
type checked struct {
	vs   []violation
	code int
}

// checkJSON 运行 harness check --json 并解析结果。
func checkJSON(t *testing.T, dir string, args ...string) checked {
	t.Helper()
	r := runHarness(t, dir, append(append([]string{"check"}, args...), "--json")...)
	if r.Code == 2 {
		t.Fatalf("内部错误：%s", r.Stderr)
	}
	var rep struct{ Violations []violation }
	if err := json.Unmarshal([]byte(r.Stdout), &rep); err != nil {
		t.Fatalf("输出不是 JSON：%v\n%s", err, r.Stdout)
	}
	return checked{rep.Violations, r.Code}
}

func wantPass(t *testing.T, c checked) {
	t.Helper()
	vs, code := c.vs, c.code
	if code != 0 || len(vs) != 0 {
		t.Fatalf("期望通过，得到退出码 %d：%+v", code, vs)
	}
}

// wantViolation 断言存在指定规则的违规，且消息包含全部片段。
func wantViolation(t *testing.T, c checked, rule string, parts ...string) violation {
	t.Helper()
	vs, code := c.vs, c.code
	if code != 1 {
		t.Fatalf("期望退出码 1，得到 %d：%+v", code, vs)
	}
	for _, v := range vs {
		if v.Rule != rule {
			continue
		}
		ok := true
		for _, p := range parts {
			if !strings.Contains(v.Msg, p) {
				ok = false
			}
		}
		if ok {
			return v
		}
	}
	t.Fatalf("没有 %s 违规包含 %q：%+v", rule, parts, vs)
	return violation{}
}

// basePRD 是一份合法的最小 PRD。
const basePRD = `# PRD: demo

- Status: Draft
- Current-Release: 1

## 2. 统一语言

| ID | 术语 | 代码名 | 定义 | 禁用别名 |
|---|---|---|---|---|
| TERM-001 | 产物 | Artifact | 结构化文档 | 交付物 |

## 3. 功能需求

### FR-001 初始化
- Status: Active
- Release: 1
- Priority: P0

- AC-001-1: 生成骨架。
  - AC-001-1/N1: 空目录
  - AC-001-1/E1: 只读目录

### FR-002 澄清
- Status: Active
- Release: 1
- Priority: P1

- AC-002-1: 记录问答。
  - AC-002-1/N1: 一轮问答
  - AC-002-1/B1: 一轮 5 个问题

## 4. 非功能需求

### NFR-001 确定性
- Status: Active
- Release: 1
- Type: Constraint
- Priority: P0

- AC-N001-1: 不联网。
  - AC-N001-1/N1: 断网运行
  - AC-N001-1/E1: 引入网络包

### NFR-002 扩展场景：新增存储
- Status: Active
- Release: 2
- Type: Scenario
- Priority: P2
- Adds: 一种存储实现
- Allowed-Paths: 存储实现目录
- Closed-Modules: 接口层

- AC-N002-1: 新增一种存储只改存储实现目录。

## 6. 待澄清项

### OQ-001 是否支持 Windows
- Status: Closed
- Resolution: 支持

## 8. 变更日志

| ID | 日期 | 动作 | 对象 | 说明 |
|---|---|---|---|---|
| CHG-001 | 2026-11-25 | 新增 | FR-001, FR-002 | 初稿 |
| CHG-001 | 2026-11-25 | 新增 | NFR-001, NFR-002 | 初稿 |
`

// prd 返回对 basePRD 依次做字符串替换后的文本；每个替换必须命中。
func prd(t *testing.T, pairs ...string) string {
	t.Helper()
	s := basePRD
	for i := 0; i+1 < len(pairs); i += 2 {
		if !strings.Contains(s, pairs[i]) {
			t.Fatalf("替换目标不存在：%q", pairs[i])
		}
		s = strings.Replace(s, pairs[i], pairs[i+1], 1)
	}
	return s
}

func checkPRD(t *testing.T, files map[string]string) checked {
	t.Helper()
	return checkJSON(t, writeTree(t, files), "prd")
}

// ---- AC-003-1 编号唯一、格式合法 ----

func TestCheck_AC_003_1_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD}))
}

func TestCheck_AC_003_1_E1(t *testing.T) {
	c := checkPRD(t, map[string]string{"docs/PRD.md": prd(t, "### FR-002 澄清", "### FR-001 澄清")})
	v := wantViolation(t, c, "prd.01", "FR-001", "docs/PRD.md:14", "docs/PRD.md:23")
	if v.Line != 23 {
		t.Fatalf("行号：%d", v.Line)
	}
}

func TestCheck_AC_003_1_E2(t *testing.T) {
	for _, bad := range []string{"FR-3", "FR003"} {
		c := checkPRD(t, map[string]string{"docs/PRD.md": prd(t, "### FR-002 澄清", "### "+bad+" 澄清")})
		wantViolation(t, c, "prd.01", bad, "FR-001 形式")
	}
}

func TestCheck_AC_003_1_B1(t *testing.T) {
	src := prd(t, "### FR-002 澄清", "### FR-999 澄清", "## 4. 非功能需求", `### FR-1000 新需求
- Status: Active
- Release: 1
- Priority: P2

- AC-1000-1: 新增。
  - AC-1000-1/N1: 正常
  - AC-1000-1/E1: 异常

## 4. 非功能需求`)
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": src}))
}

// ---- AC-003-2 必填字段 ----

func TestCheck_AC_003_2_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD}))
}

func TestCheck_AC_003_2_E1(t *testing.T) {
	c := checkPRD(t, map[string]string{"docs/PRD.md": prd(t, "- Release: 1\n- Priority: P1\n", "- Release: 1\n")})
	wantViolation(t, c, "prd.02", "FR-002", "缺少 Priority")
}

func TestCheck_AC_003_2_E2(t *testing.T) {
	c := checkPRD(t, map[string]string{"docs/PRD.md": prd(t, "### FR-002 澄清\n- Status: Active", "### FR-002 澄清\n- Status: Done")})
	wantViolation(t, c, "prd.02", "FR-002", "Done", "Active、Removed")
}

func TestCheck_AC_003_2_B1(t *testing.T) {
	src := prd(t, "### FR-002 澄清\n- Status: Active", "### FR-002 澄清\n- Status: Removed",
		"- AC-002-1: 记录问答。\n  - AC-002-1/N1: 一轮问答\n  - AC-002-1/B1: 一轮 5 个问题\n", "")
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": src}))
}

func TestCheck_AC_003_2_E3(t *testing.T) {
	src := prd(t, "- AC-002-1: 记录问答。\n  - AC-002-1/N1: 一轮问答\n  - AC-002-1/B1: 一轮 5 个问题\n", "")
	c := checkPRD(t, map[string]string{"docs/PRD.md": src})
	wantViolation(t, c, "prd.02", "FR-002", "没有验收标准")
}

// ---- AC-003-4 PRD 拆分 ----

const splitFR = `# PRD 领域 %s

### %s 拆分需求
- Status: Active
- Release: 1
- Priority: P1

- AC-%s-1: 拆分。
  - AC-%s-1/N1: 正常
  - AC-%s-1/E1: 异常
`

func split(id string) string {
	n := strings.TrimPrefix(id, "FR-")
	return fmt.Sprintf(splitFR, id, id, n, n, n)
}

func TestCheck_AC_003_4_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{
		"docs/PRD.md": basePRD, "docs/prd/a.md": split("FR-010"), "docs/prd/b.md": split("FR-011"),
	}))
}

func TestCheck_AC_003_4_B1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD}))
}

func TestCheck_AC_003_4_E1(t *testing.T) {
	c := checkPRD(t, map[string]string{
		"docs/PRD.md":   prd(t, "### FR-002 澄清", "### FR-010 澄清"),
		"docs/prd/a.md": split("FR-010"),
	})
	wantViolation(t, c, "prd.01", "FR-010", "docs/PRD.md:23", "docs/prd/a.md:3")
}

func TestCheck_AC_003_4_E2(t *testing.T) {
	c := checkPRD(t, map[string]string{
		"docs/PRD.md": basePRD,
		"docs/prd/a.md": split("FR-010") + `
## 统一语言

| ID | 术语 | 代码名 | 定义 | 禁用别名 |
|---|---|---|---|---|
| TERM-002 | 门禁 | Gate | 检查 | |
`,
	})
	wantViolation(t, c, "prd.01", "统一语言只能写在 docs/PRD.md", "TERM-002")
}

// ---- NFR-004 可操作的报错 ----

func TestCheck_AC_N004_1_N1(t *testing.T) {
	dir := writeTree(t, map[string]string{"docs/PRD.md": prd(t, "- Release: 1\n- Priority: P1\n", "- Release: 1\n")})
	r := runHarness(t, dir, "check", "prd")
	want := "docs/PRD.md:23: [prd.02] FR-002 缺少 Priority 字段\n  修复："
	if r.Code != 1 || !strings.Contains(r.Stdout, want) {
		t.Fatalf("退出码 %d，输出：\n%s", r.Code, r.Stdout)
	}
}

// 任一内置规则缺少修复建议时失败：每条规则的默认修复建议非空，且违规输出都带修复建议。
func TestCheck_AC_N004_1_E1(t *testing.T) {
	for _, r := range builtin.Rules() {
		if strings.TrimSpace(r.Fix()) == "" {
			t.Errorf("规则 %s 没有默认修复建议", r.ID())
		}
	}
	c := checkPRD(t, map[string]string{"docs/PRD.md": prd(t,
		"- Current-Release: 1\n", "", "### FR-002 澄清\n- Status: Active", "### FR-02 澄清\n- Status: Done")})
	if c.code != 1 || len(c.vs) < 3 {
		t.Fatalf("期望多条违规：%+v", c.vs)
	}
	for _, v := range c.vs {
		if v.Fix == "" || v.Rule == "" || v.File == "" || v.Line == 0 {
			t.Errorf("违规字段不完整：%+v", v)
		}
	}
}

func TestCheck_AC_N004_2_N1(t *testing.T) {
	bad := writeTree(t, map[string]string{"docs/PRD.md": prd(t, "- Release: 1\n- Priority: P1\n", "- Release: 1\n")})
	c := checkJSON(t, bad, "prd")
	text := runHarness(t, bad, "check", "prd")
	if c.code != 1 || len(c.vs) != 1 || text.Code != 1 {
		t.Fatalf("退出码 %d/%d：%+v", c.code, text.Code, c.vs)
	}
	v := c.vs[0]
	line := fmt.Sprintf("%s:%d: [%s] %s\n  修复：%s\n", v.File, v.Line, v.Rule, v.Msg, v.Fix)
	if !strings.HasPrefix(text.Stdout, line) {
		t.Fatalf("文本输出与 JSON 字段不一致：\n%s\n%s", line, text.Stdout)
	}
	good := writeTree(t, map[string]string{"docs/PRD.md": basePRD})
	wantPass(t, checkJSON(t, good, "prd"))
}

func TestCheck_AC_N004_2_E1(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root 可以读取任何文件")
	}
	dir := writeTree(t, map[string]string{"docs/PRD.md": basePRD})
	if err := os.Chmod(filepath.Join(dir, "docs/PRD.md"), 0o000); err != nil {
		t.Fatal(err)
	}
	r := runHarness(t, dir, "check", "prd")
	if r.Code != 2 || !strings.Contains(r.Stderr, "内部错误") {
		t.Fatalf("期望退出码 2，得到 %d：%s", r.Code, r.Stderr)
	}
}
