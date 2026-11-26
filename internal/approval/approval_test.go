package approval

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/workspace"
)

const prd = `# PRD: demo

- Status: Draft
- Current-Release: 1

背景说明。

## 统一语言

| 编号 | 术语 | 代码名 | 定义 | 禁用别名 |
|---|---|---|---|---|
| TERM-001 | 订单 | Order | 一次购买 | |

## 功能需求

### FR-001 下单
- Status: Active
- Release: 1
- Priority: P0

- AC-001-1: 下单成功。
  - AC-001-1/N1: 正常下单

### FR-002 取消
- Status: Active
- Release: 1
- Priority: P0

- AC-002-1: 取消成功。
  - AC-002-1/N1: 正常取消

## 变更日志

| 编号 | 日期 | 动作 | 对象 | 说明 |
|---|---|---|---|---|
| CHG-001 | 2026-01-01 | 新增 | FR-001 | 初稿 |
`

const overview = `# 架构总览

- Status: Draft

### MOD-01 order
- Status: Active
- Paths: internal/order/**
- Public: internal/order/order.go
- Depends:
- Interfaces:
- Covers: FR-001
- Core: false
- Coverage: 80
`

const tasks = `# 任务

- Status: Draft

### T-001 下单
- Status: Todo
- Covers: AC-001-1
- Design: MOD-01
- Scope: internal/order/**
- Depends:
`

func load(t *testing.T, dir string) *workspace.Snapshot {
	t.Helper()
	s, err := workspace.Load(dir, workspace.LoadOptions{})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func project(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for rel, c := range files {
		p := filepath.Join(dir, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func demo(t *testing.T) string {
	return project(t, map[string]string{
		"docs/PRD.md": prd, "docs/architecture/overview.md": overview, "docs/tasks.md": tasks,
		"harness.lock": `{"version": 2}`,
	})
}

func byID(units []Unit) map[string]Unit {
	m := map[string]Unit{}
	for _, u := range units {
		m[u.ID] = u
	}
	return m
}

func edit(t *testing.T, dir, rel, old, new string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), old) {
		t.Fatalf("%s 中没有 %q", rel, old)
	}
	os.WriteFile(p, []byte(strings.Replace(string(b), old, new, 1)), 0o644)
}

func TestUnits_IF_027(t *testing.T) {
	dir := demo(t)
	u := byID(Units(load(t, dir)))
	for id, k := range map[string]UnitKind{
		"FR-001": Requirement, "FR-002": Requirement, "TERM-001": Term, "MOD-01": Design, "T-001": Task,
		"docs/PRD.md#正文": Body, "docs/architecture/overview.md#正文": Body, "docs/tasks.md#正文": Body,
	} {
		if u[id].Kind != k {
			t.Errorf("%s 的类型为 %q，期望 %q", id, u[id].Kind, k)
		}
	}
	if got := u["T-001"].Upstream; strings.Join(got, ",") != "FR-001,MOD-01" {
		t.Errorf("任务的上游应把 AC 归到所属需求：%v", got)
	}
	if strings.Contains(u["docs/PRD.md#正文"].Text, "CHG-001") || strings.Contains(u["docs/PRD.md#正文"].Text, "Status") {
		t.Errorf("PRD 正文不应含变更日志与 Status：%q", u["docs/PRD.md#正文"].Text)
	}
	if strings.Contains(u["FR-001"].Text, "Status") || !strings.Contains(u["FR-001"].Text, "AC-001-1/N1") {
		t.Errorf("需求单位应含验收标准与用例、不含 Status：%q", u["FR-001"].Text)
	}
	// Status、变更日志与空白变化不改变哈希；正文变化只改变正文单位。
	before := u
	edit(t, dir, "docs/PRD.md", "- Status: Active\n- Release: 1\n- Priority: P0\n\n- AC-001-1", "- Status: Removed\n- Release: 1   \n- Priority: P0\n\n\n- AC-001-1")
	edit(t, dir, "docs/PRD.md", "| 初稿 |", "| 初稿，补充 |")
	edit(t, dir, "docs/tasks.md", "- Status: Todo", "- Status: InProgress")
	after := byID(Units(load(t, dir)))
	for _, id := range []string{"FR-001", "docs/PRD.md#正文", "T-001"} {
		if before[id].Hash != after[id].Hash {
			t.Errorf("%s 的哈希不应变化", id)
		}
	}
	edit(t, dir, "docs/PRD.md", "背景说明。", "背景说明，补充一句。")
	after = byID(Units(load(t, dir)))
	if before["docs/PRD.md#正文"].Hash == after["docs/PRD.md#正文"].Hash || before["FR-001"].Hash != after["FR-001"].Hash {
		t.Error("改正文只应改变正文单位的哈希")
	}
	if !after["FR-001"].Owns("docs/PRD.md", after["FR-001"].Line+5) || after["FR-001"].Owns("docs/PRD.md", after["FR-002"].Line) {
		t.Error("Owns 应按单位的行范围判定")
	}
}

// approveAll 把全部单位记为已批准并写回锁文件。
func approveAll(t *testing.T, dir string) {
	t.Helper()
	s := load(t, dir)
	s.Lock.Units = Record(s, Units(s), "tester", time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC))
	if err := s.Lock.Write(filepath.Join(dir, lock.File)); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluate_IF_028(t *testing.T) {
	dir := demo(t)
	if st := Evaluate(load(t, dir))["FR-001"]; st.State != Unapproved {
		t.Fatalf("没有记录时应为 unapproved：%+v", st)
	}
	approveAll(t, dir)
	for id, st := range Evaluate(load(t, dir)) {
		if st.State != Approved {
			t.Errorf("%s 应为 approved：%+v", id, st)
		}
	}
	// 上游 FR-001 变化：FR-001 自身、覆盖它的 MOD-01、引用两者的 T-001 失效；FR-002 不受影响。
	edit(t, dir, "docs/PRD.md", "下单成功。", "下单成功并发送通知。")
	st := Evaluate(load(t, dir))
	want := map[string]string{"FR-001": "FR-001", "MOD-01": "FR-001", "T-001": "FR-001,MOD-01"}
	for id, cause := range want {
		if st[id].State != Stale || strings.Join(st[id].Cause, ",") != cause {
			t.Errorf("%s：%+v，期望 stale，原因 %s", id, st[id].State, cause)
		}
	}
	if st["FR-002"].State != Approved {
		t.Errorf("FR-002 不应失效：%+v", st["FR-002"])
	}
	chain := Chain(load(t, dir), "T-001")
	var ids []string
	for _, c := range chain {
		ids = append(ids, c.Unit.ID)
	}
	if strings.Join(ids, ",") != "FR-001,MOD-01,T-001" {
		t.Errorf("批准链：%v", ids)
	}
}

func TestPendingIn_IF_029(t *testing.T) {
	dir := demo(t)
	approveAll(t, dir)
	edit(t, dir, "docs/PRD.md", "取消成功。", "取消成功并退款。")
	s := load(t, dir)
	p, err := PendingIn(s, "prd")
	if err != nil || len(p) != 1 || p[0].Unit.ID != "FR-002" || !strings.Contains(p[0].Before, "取消成功。") {
		t.Fatalf("prd 范围内应只有修改的 FR-002，且带上次批准的内容：%v %+v", err, p)
	}
	if p, _ := PendingIn(s, "docs/architecture/overview.md"); len(p) != 0 {
		t.Errorf("设计未变化，不应有待批准单位：%+v", p)
	}
	if p, _ := PendingIn(s, "FR-001"); len(p) != 0 {
		t.Errorf("FR-001 已批准：%+v", p)
	}
	if _, err := PendingIn(s, "FR-999"); err == nil {
		t.Error("不存在的单位应返回错误")
	}
	rec := Record(s, []Unit{byID(Units(s))["T-001"]}, "x", time.Now())
	if rec["T-001"].Upstream["MOD-01"] == "" || rec["T-001"].Upstream["FR-001"] == "" {
		t.Errorf("批准记录应含上游哈希：%+v", rec["T-001"])
	}
}

func TestSummarize_IF_028(t *testing.T) {
	dir := demo(t)
	if sum := Summarize(load(t, dir)); sum.Skill != "clarify" || !sum.NoApproved || sum.Unapproved == 0 {
		t.Fatalf("没有批准时应建议 clarify：%+v", sum)
	}
	approveAll(t, dir)
	if sum := Summarize(load(t, dir)); sum.Skill != "architect" || sum.NoApproved {
		t.Fatalf("FR-002 未被设计覆盖时应建议 architect：%+v", sum)
	}
	edit(t, dir, "docs/architecture/overview.md", "- Coverage: 80\n", "- Coverage: 80\n\n### MOD-02 cancel\n- Status: Active\n- Paths: internal/cancel/**\n- Public: internal/cancel/cancel.go\n- Depends:\n- Interfaces:\n- Covers: FR-002\n- Core: false\n- Coverage: 80\n")
	approveAll(t, dir)
	if sum := Summarize(load(t, dir)); sum.Skill != "plan" || strings.Join(sum.Ready, ",") != "T-001" {
		t.Fatalf("FR-002 未被任务覆盖时应建议 plan，T-001 可开工：%+v", sum)
	}
	edit(t, dir, "docs/tasks.md", "- Depends:\n", "- Depends:\n\n### T-002 取消\n- Status: Todo\n- Covers: AC-002-1\n- Design: MOD-02\n- Scope: internal/cancel/**\n- Depends:\n")
	approveAll(t, dir)
	if sum := Summarize(load(t, dir)); sum.Skill != "implement" || strings.Join(sum.Ready, ",") != "T-001,T-002" || sum.Unapproved+sum.Stale != 0 {
		t.Fatalf("全部批准时应建议 implement：%+v", sum)
	}
	edit(t, dir, "docs/PRD.md", "## 变更日志", "## 待澄清项\n\n### OQ-001 下单是否限额\n- Status: Open\n- Relates: FR-001\n\n## 变更日志")
	approveAll(t, dir)
	if sum := Summarize(load(t, dir)); strings.Join(sum.Ready, ",") != "T-002" || strings.Join(sum.OpenQuestions, ",") != "OQ-001" {
		t.Fatalf("关联 Open 待澄清项的 T-001 不可开工：%+v", sum)
	}
	edit(t, dir, "docs/PRD.md", "取消成功。", "取消成功并退款。")
	edit(t, dir, "docs/tasks.md", "### T-001 下单\n- Status: Todo", "### T-001 下单\n- Status: InProgress")
	if sum := Summarize(load(t, dir)); len(sum.Ready) != 0 || sum.Skill != "clarify" || sum.Stale == 0 || strings.Join(sum.InProgress, ",") != "T-001" {
		t.Fatalf("FR-002 修改后 T-002 不可开工，应建议 clarify：%+v", sum)
	}
}
