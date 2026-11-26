package app_test

import (
	"fmt"
	"strings"
	"testing"
)

// basePRD 中的全部用例对应的测试名。
var caseTests = []string{"TestX_AC_001_1_N1", "TestX_AC_001_1_E1", "TestX_AC_002_1_N1", "TestX_AC_002_1_B1", "TestX_AC_N001_1_N1", "TestX_AC_N001_1_E1"}

func goTests(names ...string) string {
	var b strings.Builder
	b.WriteString("package a\n\nimport \"testing\"\n\n")
	for _, n := range names {
		fmt.Fprintf(&b, "func %s(t *testing.T) {}\n\n", n)
	}
	return b.String()
}

// goReport 生成 go test -json 报告；参数为测试名与结果（pass、fail、skip）交替。
func goReport(pairs ...string) string {
	var b strings.Builder
	for i := 0; i+1 < len(pairs); i += 2 {
		fmt.Fprintf(&b, `{"Action":"run","Package":"m/internal","Test":"%s"}`+"\n", pairs[i])
		fmt.Fprintf(&b, `{"Action":"%s","Package":"m/internal","Test":"%s"}`+"\n", pairs[i+1], pairs[i])
	}
	return b.String()
}

func allPass(names ...string) []string {
	var out []string
	for _, n := range names {
		out = append(out, n, "pass")
	}
	return out
}

// covered 返回全部用例都有通过的集成测试的项目文件，可再按需覆盖。
func covered() map[string]string {
	return map[string]string{
		"docs/PRD.md":                    basePRD,
		"internal/a_integration_test.go": goTests(caseTests...),
		".harness/reports/go.jsonl":      goReport(allPass(caseTests...)...),
	}
}

func checkTests(t *testing.T, files map[string]string) checked {
	t.Helper()
	return checkJSON(t, writeTree(t, files), "test")
}

// ---- AC-020-1 测试名引用用例 ----

func TestCheck_AC_020_1_N1(t *testing.T) {
	wantPass(t, checkTests(t, covered()))
	// 去掉 TestX_AC_001_1_N1 后 AC-001-1/N1 未覆盖，说明此前它被识别为引用。
	f := covered()
	f["internal/a_integration_test.go"] = goTests(caseTests[1:]...)
	f[".harness/reports/go.jsonl"] = goReport(allPass(caseTests[1:]...)...)
	wantViolation(t, checkTests(t, f), "test.02", "AC-001-1/N1")
}

func TestCheck_AC_020_1_B1(t *testing.T) {
	names := append([]string{"TestX_AC_001_1_N1_AC_001_1_E1"}, caseTests[2:]...)
	f := covered()
	f["internal/a_integration_test.go"] = goTests(names...)
	f[".harness/reports/go.jsonl"] = goReport(allPass(names...)...)
	wantPass(t, checkTests(t, f))
}

func TestCheck_AC_020_1_E1(t *testing.T) {
	names := append([]string{"TestX_AC_009_1_N1"}, caseTests...)
	f := covered()
	f["internal/a_integration_test.go"] = goTests(names...)
	f[".harness/reports/go.jsonl"] = goReport(allPass(names...)...)
	v := wantViolation(t, checkTests(t, f), "test.01", "TestX_AC_009_1_N1", "AC-009-1/N1 不存在")
	if v.File != "internal/a_integration_test.go" {
		t.Fatalf("位置：%s", v.File)
	}
}

// ---- AC-020-2 每个用例有通过的集成测试 ----

func TestCheck_AC_020_2_N1(t *testing.T) {
	wantPass(t, checkTests(t, covered()))
}

func TestCheck_AC_020_2_E1(t *testing.T) {
	f := covered()
	f[".harness/reports/go.jsonl"] = strings.Replace(f[".harness/reports/go.jsonl"],
		`"pass","Package":"m/internal","Test":"TestX_AC_001_1_E1"`, `"fail","Package":"m/internal","Test":"TestX_AC_001_1_E1"`, 1)
	wantViolation(t, checkTests(t, f), "test.02", "AC-001-1/E1", "失败", "TestX_AC_001_1_E1")
}

func TestCheck_AC_020_2_E2(t *testing.T) {
	f := covered()
	delete(f, ".harness/reports/go.jsonl")
	v := wantViolation(t, checkTests(t, f), "test.02", "没有测试报告")
	if !strings.Contains(v.Fix, "go test -json") {
		t.Fatalf("修复建议缺少生成报告的命令：%s", v.Fix)
	}
}

func TestCheck_AC_020_2_B1(t *testing.T) {
	names := append([]string{"TestY_AC_001_1_E1"}, caseTests...)
	f := covered()
	f["internal/a_integration_test.go"] = goTests(names...)
	f[".harness/reports/go.jsonl"] = goReport(append(allPass(caseTests...), "TestY_AC_001_1_E1", "fail")...)
	wantViolation(t, checkTests(t, f), "test.02", "AC-001-1/E1", "TestY_AC_001_1_E1")
}

// ---- AC-020-3 报告与源码一致，skipped 不计 ----

func TestCheck_AC_020_3_N1(t *testing.T) {
	wantPass(t, checkTests(t, covered()))
}

func TestCheck_AC_020_3_E1(t *testing.T) {
	f := covered()
	f[".harness/reports/go.jsonl"] = strings.Replace(f[".harness/reports/go.jsonl"],
		`"pass","Package":"m/internal","Test":"TestX_AC_002_1_B1"`, `"skip","Package":"m/internal","Test":"TestX_AC_002_1_B1"`, 1)
	wantViolation(t, checkTests(t, f), "test.02", "AC-002-1/B1", "skipped")
}

func TestCheck_AC_020_3_E2(t *testing.T) {
	f := covered()
	f["internal/b_integration_test.go"] = goTests("TestZ_AC_001_1_N1")
	wantViolation(t, checkTests(t, f), "test.02", "TestZ_AC_001_1_N1", "没有出现在测试报告中")
}

// ---- AC-020-4 Go 集成测试的判定 ----

func TestCheck_AC_020_4_N1(t *testing.T) {
	f := covered()
	delete(f, "internal/a_integration_test.go")
	f["internal/a_integration_test.go"] = goTests(caseTests[:3]...)
	f["internal/b_test.go"] = "//go:build integration\n\n" + goTests(caseTests[3:]...)
	wantPass(t, checkTests(t, f))
}

func TestCheck_AC_020_4_E1(t *testing.T) {
	f := covered()
	f["internal/a_integration_test.go"] = goTests(caseTests[1:]...)
	f["internal/a_test.go"] = goTests(caseTests[0])
	wantViolation(t, checkTests(t, f), "test.02", "AC-001-1/N1", "单元测试")
}
