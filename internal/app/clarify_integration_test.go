package app_test

import (
	"fmt"
	"strings"
	"testing"
)

// withClarify 在变更日志前插入澄清记录章节。
func withClarify(t *testing.T, body string) string {
	return prd(t, "## 8. 变更日志", "## 7. 澄清记录\n\n"+body+"\n## 8. 变更日志")
}

func questions(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "- Q: 问题 %d？ A: 答案 %d。\n", i, i)
	}
	return b.String()
}

// ---- AC-002-1 澄清记录按轮次、每轮不超过 5 问 ----

func TestCheck_AC_002_1_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": withClarify(t, "### 2026-11-25 第 1 轮\n"+questions(5))}))
}

func TestCheck_AC_002_1_B1(t *testing.T) {
	c := checkPRD(t, map[string]string{"docs/PRD.md": withClarify(t, "### 2026-11-25 第 1 轮\n"+questions(6))})
	wantViolation(t, c, "prd.06", "2026-11-25 第 1 轮", "6 个问题")
}

func TestCheck_AC_002_1_E1(t *testing.T) {
	c := checkPRD(t, map[string]string{"docs/PRD.md": withClarify(t, questions(1))})
	wantViolation(t, c, "prd.06", "不在任何轮次标题下")
	c = checkPRD(t, map[string]string{"docs/PRD.md": withClarify(t, "### 第 1 轮\n"+questions(1))})
	wantViolation(t, c, "prd.06", "缺少日期")
}

// ---- AC-002-2 Open 待澄清项阻止 Approved ----

const oqOpen = "### OQ-001 是否支持 Windows\n- Status: Open\n\n## 8."

func approved(t *testing.T, pairs ...string) string {
	return prd(t, append([]string{"- Status: Draft\n", "- Status: Approved\n"}, pairs...)...)
}

func TestCheck_AC_002_2_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": approved(t)}))
}

func TestCheck_AC_002_2_B1(t *testing.T) {
	open := approved(t, "### OQ-001 是否支持 Windows\n- Status: Closed\n- Resolution: 支持\n\n## 8.", oqOpen)
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": open}), "prd.05", "OQ-001")
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": approved(t)}))
}

func TestCheck_AC_002_2_E1(t *testing.T) {
	src := approved(t, "## 8. 变更日志", "### OQ-002 是否支持 Linux\n- Status: Open\n\n## 8. 变更日志")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.05", "Approved", "OQ-002")
}

func TestCheck_AC_002_2_E2(t *testing.T) {
	src := prd(t, "- Status: Closed\n- Resolution: 支持\n", "- Status: Closed\n")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.05", "OQ-001", "没有 Resolution")
}

// ---- AC-002-3 三章齐全且非空 ----

func TestCheck_AC_002_3_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD}))
}

func TestCheck_AC_002_3_E1(t *testing.T) {
	src := prd(t, "## 2. 统一语言", "## 2. 术语表")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.04", "缺少章节：统一语言")
}

func TestCheck_AC_002_3_B1(t *testing.T) {
	i := strings.Index(basePRD, "### NFR-001")
	j := strings.Index(basePRD, "## 6. 待澄清项")
	src := basePRD[:i] + basePRD[j:]
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.04", "非功能需求", "没有任何条目")
}

// ---- AC-019-1 用例编号格式 ----

func TestCheck_AC_019_1_N1(t *testing.T) {
	src := prd(t, "  - AC-001-1/E1: 只读目录", "  - AC-001-1/B1: 只读目录")
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": src}))
}

func TestCheck_AC_019_1_E1(t *testing.T) {
	src := prd(t, "  - AC-001-1/E1: 只读目录", "  - AC-001-1/E1: 只读目录\n  - AC-001-1/X1: 未知类别")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.07", "AC-001-1/X1", "类别 X 不合法")
}

func TestCheck_AC_019_1_E2(t *testing.T) {
	src := prd(t, "  - AC-001-1/E1: 只读目录\n", "  - AC-001-1/E1: 只读目录\n- AC-001-2: 第二条。\n  - AC-001-1/N1: 前缀错误\n  - AC-001-2/E1: 异常\n")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.07", "AC-001-1/N1", "与所在验收标准 AC-001-2 不一致")
}

// ---- AC-019-2 每条验收标准至少 N 加 B 或 E ----

func TestCheck_AC_019_2_N1(t *testing.T) {
	src := prd(t, "  - AC-001-1/E1: 只读目录", "  - AC-001-1/B1: 目录含隐藏文件")
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": src}))
}

func TestCheck_AC_019_2_B1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD})) // AC-001-1 只有 N1 和 E1
}

func TestCheck_AC_019_2_E1(t *testing.T) {
	src := prd(t, "  - AC-001-1/E1: 只读目录\n", "")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.07", "AC-001-1", "至少一个 N")
}

func TestCheck_AC_019_2_E2(t *testing.T) {
	src := prd(t, "  - AC-001-1/N1: 空目录\n", "  - AC-001-1/B1: 目录含隐藏文件\n")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.07", "AC-001-1", "至少一个 N")
}

func TestCheck_AC_019_2_B2(t *testing.T) {
	src := prd(t, "### FR-002 澄清\n- Status: Active\n- Release: 1\n", "### FR-002 澄清\n- Status: Active\n- Release: 2\n",
		"  - AC-002-1/N1: 一轮问答\n  - AC-002-1/B1: 一轮 5 个问题\n", "")
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": src}))
}

// ---- AC-040-1 模糊词 ----

func TestCheck_AC_040_1_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD}))
}

func TestCheck_AC_040_1_E1(t *testing.T) {
	src := prd(t, "- AC-001-1: 生成骨架。", "- AC-001-1: 生成骨架，响应要快速。")
	v := wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.09", "快速")
	if v.Line != 19 {
		t.Fatalf("行号：%d", v.Line)
	}
}

func TestCheck_AC_040_1_B1(t *testing.T) {
	src := prd(t, "- AC-001-1: 生成骨架。", "- AC-001-1: 生成骨架，输出 `快速` 一词。")
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": src}))
}

func TestCheck_AC_040_1_B2(t *testing.T) {
	src := prd(t, "- AC-001-1: 生成骨架。", "- AC-001-1: 等待生成结束，结果与模板相等。")
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": src}))
}

// ---- AC-040-2 Quantitative NFR 含数值与单位 ----

func quantitative(t *testing.T, ac string) string {
	return prd(t, "- Type: Constraint\n", "- Type: Quantitative\n", "- AC-N001-1: 不联网。", "- AC-N001-1: "+ac)
}

func TestCheck_AC_040_2_N1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": quantitative(t, "门禁耗时 < 200ms。")}))
}

func TestCheck_AC_040_2_E1(t *testing.T) {
	c := checkPRD(t, map[string]string{"docs/PRD.md": quantitative(t, "响应及时。")})
	wantViolation(t, c, "prd.10", "NFR-001", "没有数值和单位")
}

func TestCheck_AC_040_2_B1(t *testing.T) {
	wantPass(t, checkPRD(t, map[string]string{"docs/PRD.md": basePRD})) // Constraint 类不要求数值
}

func TestCheck_AC_040_2_E2(t *testing.T) {
	src := prd(t, "- Type: Constraint\n", "")
	wantViolation(t, checkPRD(t, map[string]string{"docs/PRD.md": src}), "prd.02", "NFR-001", "缺少 Type")
}
