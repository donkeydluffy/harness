package app_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// 性能阈值（NFR-003）。
const (
	gateP95Limit   = 200 * time.Millisecond
	fullCheckLimit = 10 * time.Second
	prdCheckLimit  = 2 * time.Second
)

// perfRepo 生成 5000 个源文件、200 个任务的样例仓库。
func perfRepo(t *testing.T) string {
	t.Helper()
	files := traced(t, basePRD)
	var tasks []string
	for i := 1; i <= 200; i++ {
		status := "Todo"
		if i == 1 {
			status = "InProgress"
		}
		tasks = append(tasks, task(fmt.Sprintf("T-%03d", i), status, "FR-001", "MOD-01", fmt.Sprintf("internal/m%03d/**", i%50), "", ""))
	}
	files["docs/tasks.md"] = taskDoc(tasks...)
	dir := writeTree(t, files)
	body := strings.Repeat("\nfunc helper%d(x int) int {\n\t// 普通注释\n\treturn x + %d\n}\n", 6)
	for m := 0; m < 50; m++ {
		d := filepath.Join(dir, fmt.Sprintf("internal/m%03d", m))
		os.MkdirAll(d, 0o755)
		for f := 0; f < 100; f++ {
			src := fmt.Sprintf("package m%03d\n\n// TODO(T-001): 示例\n", m) + fmt.Sprintf(body, f, f, f+1, f, f+2, f, f+3, f, f+4, f, f+5, f)
			name := fmt.Sprintf("f%03d.go", f)
			if f%10 == 0 {
				name = fmt.Sprintf("f%03d_test.go", f)
				src = fmt.Sprintf("package m%03d\n\nimport \"testing\"\n\nfunc TestF%d(t *testing.T) {}\n", m, f)
			}
			if err := os.WriteFile(filepath.Join(d, name), []byte(src), 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return dir
}

// p95 返回第 95 百分位耗时。
func p95(ds []time.Duration) time.Duration {
	s := append([]time.Duration(nil), ds...)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	return s[(len(s)*95+99)/100-1]
}

// overLimit 返回超出阈值时的说明，未超出时为空串。
func overLimit(what string, got, limit time.Duration) string {
	if got > limit {
		return fmt.Sprintf("%s 耗时 %v，超过 %v", what, got.Round(time.Millisecond), limit)
	}
	return ""
}

func timed(f func()) time.Duration {
	start := time.Now()
	f()
	return time.Since(start)
}

// ---- AC-N003-1 写操作门禁耗时 ----

func TestPerf_AC_N003_1_N1(t *testing.T) {
	dir := perfRepo(t)
	var ds []time.Duration
	for i := 0; i < 40; i++ {
		ds = append(ds, timed(func() {
			wantAllow(t, hook(t, dir, "PreToolUse", writeTool(dir, "internal/m001/new.go", "package m001\n")))
		}))
	}
	if msg := overLimit("PreToolUse 门禁 p95", p95(ds), gateP95Limit); msg != "" {
		t.Fatal(msg)
	}
	t.Logf("PreToolUse 门禁 p95 = %v", p95(ds))
}

func TestPerf_AC_N003_1_E1(t *testing.T) {
	ds := []time.Duration{}
	for i := 0; i < 19; i++ {
		ds = append(ds, 50*time.Millisecond)
	}
	ds = append(ds, 250*time.Millisecond) // 20 次中 1 次慢：p95 取第 19 个，仍在阈值内
	if overLimit("p95", p95(ds), gateP95Limit) != "" {
		t.Fatal("p95 计算错误")
	}
	ds = append(ds, 250*time.Millisecond, 250*time.Millisecond)
	if msg := overLimit("PreToolUse 门禁 p95", p95(ds), gateP95Limit); !strings.Contains(msg, "超过 200ms") {
		t.Fatalf("超过阈值时应判为失败：%q", msg)
	}
}

// ---- AC-N003-2 全量检查耗时 ----

func TestPerf_AC_N003_2_N1(t *testing.T) {
	dir := perfRepo(t)
	// 最后一个源文件放一条没有引用编号的待办注释，确认全部源文件都被扫描。
	write(t, dir, "internal/m049/f099.go", "package m049\n\n// TODO 以后再说\n")
	var r result
	d := timed(func() { r = runHarness(t, dir, "check") })
	if r.Code == 2 || !strings.Contains(r.Stdout, "internal/m049/f099.go:3: [debt.02]") {
		t.Fatalf("全量检查没有扫描到最后一个源文件：%d %s", r.Code, r.Stderr)
	}
	if msg := overLimit("5000 个源文件的全量检查", d, fullCheckLimit); msg != "" {
		t.Fatal(msg)
	}
	t.Logf("全量检查 %v", d)
}

func TestPerf_AC_N003_2_B1(t *testing.T) {
	var b strings.Builder
	b.WriteString("# PRD\n\n- Status: Draft\n- Current-Release: 1\n\n## 2. 统一语言\n\n| ID | 术语 | 代码名 | 定义 | 禁用别名 |\n|---|---|---|---|---|\n| TERM-001 | 产物 | Artifact | 结构化文档 | |\n\n## 3. 功能需求\n")
	for i := 1; b.Len() > 0 && strings.Count(b.String(), "\n") < 50000; i++ {
		fmt.Fprintf(&b, "\n### FR-%04d 需求 %d\n- Status: Active\n- Release: 2\n- Priority: P2\n\n- AC-%04d-1: 返回第 %d 项结果。\n  - AC-%04d-1/N1: 正常输入\n  - AC-%04d-1/E1: 非法输入\n- AC-%04d-2: 记录第 %d 项日志。\n  - AC-%04d-2/N1: 正常写入\n", i, i, i, i, i, i, i, i, i)
	}
	b.WriteString("\n## 4. 非功能需求\n\n不预留扩展。\n\n### NFR-001 约束\n- Status: Active\n- Release: 2\n- Type: Constraint\n- Priority: P2\n\n- AC-N001-1: 可验证。\n  - AC-N001-1/N1: 通过\n")
	dir := writeTree(t, map[string]string{"docs/PRD.md": b.String()})
	var r result
	d := timed(func() { r = runHarness(t, dir, "check", "prd") })
	if r.Code == 2 {
		t.Fatalf("内部错误：%s", r.Stderr)
	}
	if msg := overLimit("50000 行 PRD 的 harness check prd", d, prdCheckLimit); msg != "" {
		t.Fatal(msg)
	}
	t.Logf("check prd %v（%d 行）", d, strings.Count(b.String(), "\n"))
}
