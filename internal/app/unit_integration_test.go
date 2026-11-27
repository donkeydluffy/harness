package app_test

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

const (
	storeTest = "package store\n\nimport \"testing\"\n\nfunc TestSave_IF_002(t *testing.T) {\n\tif Save(nil) != nil {\n\t\tt.Fatal(\"save\")\n\t}\n}\n"
	apiTest   = "package api\n\nimport \"testing\"\n\nfunc TestServe_IF_001(t *testing.T) { Serve() }\n"
)

// goUnitRepo 返回带单元测试的 Go 项目（IF-001、IF-002 各有一个测试）；extra 同 anchorRepo。
func goUnitRepo(t *testing.T, extra map[string]string) string {
	t.Helper()
	files := map[string]string{"internal/store/store_test.go": storeTest, "internal/api/api_test.go": apiTest}
	for k, v := range extra {
		files[k] = v
	}
	return anchorRepo(t, files)
}

// runGoTests 在项目中运行 go test -json，写入测试报告。
func runGoTests(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command(goBin(t), "test", "-count=1", "-json", "./...")
	cmd.Dir = dir
	out, _ := cmd.Output()
	writeFile(t, dir, ".harness/reports/go.jsonl", string(out))
}

// unitOK 生成测试报告并运行 harness check unit 写入 unit 报告。
func unitOK(t *testing.T, dir string) {
	t.Helper()
	runGoTests(t, dir)
	if r := runHarness(t, dir, "check", "unit"); r.Code == 2 {
		t.Fatalf("harness check unit 内部错误：%s", r.Stderr)
	}
	if !strings.Contains(readFile(t, dir, ".harness/reports/unit.json"), "sourceHash") {
		t.Fatal("应写入 unit 报告")
	}
}

// setModuleCoverage 改写 unit 报告中模块的行覆盖率；源码未变，报告仍有效。
func setModuleCoverage(t *testing.T, dir, mod string, v float64) {
	t.Helper()
	var rep map[string]any
	if err := json.Unmarshal([]byte(readFile(t, dir, ".harness/reports/unit.json")), &rep); err != nil {
		t.Fatal(err)
	}
	rep["modules"].(map[string]any)[mod] = v
	b, _ := json.Marshal(rep)
	writeFile(t, dir, ".harness/reports/unit.json", string(b))
}

// ---- AC-021-1 每个接口有通过的单元测试 ----

func TestUnit_AC_021_1_N1(t *testing.T) {
	dir := goUnitRepo(t, nil)
	runGoTests(t, dir)
	wantPass(t, checkJSON(t, dir, "unit.01"))
}

func TestUnit_AC_021_1_E1(t *testing.T) {
	dir := goUnitRepo(t, map[string]string{"internal/store/store_test.go": strings.Replace(storeTest, "TestSave_IF_002", "TestSave", 1)})
	runGoTests(t, dir)
	wantViolation(t, checkJSON(t, dir, "unit.01"), "unit.01", "IF-002 没有测试名含 IF_002")
}

func TestUnit_AC_021_1_E2(t *testing.T) {
	dir := goUnitRepo(t, map[string]string{"internal/store/store_test.go": strings.Replace(storeTest, "Save(nil) != nil", "true", 1)})
	runGoTests(t, dir)
	wantViolation(t, checkJSON(t, dir, "unit.01"), "unit.01", "IF-002", "failed", "TestSave_IF_002")
}

// ---- AC-021-2 测试真正调用了接口 ----

func TestUnit_AC_021_2_N1(t *testing.T) {
	dir := goUnitRepo(t, nil)
	unitOK(t, dir)
	wantPass(t, checkJSON(t, dir, "unit.02"))
}

func TestUnit_AC_021_2_E1(t *testing.T) {
	dir := goUnitRepo(t, map[string]string{"internal/store/store_test.go": "package store\n\nimport \"testing\"\n\nfunc TestSave_IF_002(t *testing.T) { _ = len(\"x\") }\n"})
	unitOK(t, dir)
	v := wantViolation(t, checkJSON(t, dir, "unit.02"), "unit.02", "IF-002 的单元测试 TestSave_IF_002 没有调用 IF-002")
	if v.File != "internal/store/store.go" {
		t.Fatalf("应指向锚点：%+v", v)
	}
}

func TestUnit_AC_021_2_E2(t *testing.T) {
	dir := goUnitRepo(t, nil)
	runGoTests(t, dir)
	wantViolation(t, checkJSON(t, dir, "unit.02"), "unit.02", "没有 unit 报告", "harness check unit")
}

func TestUnit_AC_021_2_B1(t *testing.T) {
	dir := writeTree(t, map[string]string{"harness.json": `{"languages": ["python"], "sourcePaths": ["src/**"]}`, "src/a.py": "x = 1\n",
		".harness/reports/go.jsonl": "{\"Action\":\"pass\",\"Package\":\"p\",\"Test\":\"TestX\"}\n"})
	r := runHarness(t, dir, "check", "unit.02", "unit.03", "--json")
	var rep struct{ Violations, Warnings []violation }
	if err := json.Unmarshal([]byte(r.Stdout), &rep); err != nil || r.Code != 0 || len(rep.Violations) != 0 {
		t.Fatalf("只有 Python 源码时跳过：%d %v %s", r.Code, err, r.Stdout)
	}
	if len(rep.Warnings) != 1 || !strings.Contains(rep.Warnings[0].Msg, "python 不支持覆盖率采集，跳过 AC-021-2 与 AC-021-3") {
		t.Fatalf("应在输出中说明跳过：%+v", rep.Warnings)
	}
}

// ---- AC-021-3 模块行覆盖率阈值 ----

// coverageRepo 返回 MOD-01 阈值为 60、unit 报告中 MOD-01 覆盖率为 got 的项目。
func coverageRepo(t *testing.T, got float64) string {
	t.Helper()
	dir := goUnitRepo(t, nil)
	editFile(t, dir, "docs/architecture/overview.md", "- Core: true\n- Coverage: 80", "- Core: true\n- Coverage: 60")
	unitOK(t, dir)
	setModuleCoverage(t, dir, "MOD-01", got)
	return dir
}

func TestUnit_AC_021_3_N1(t *testing.T) {
	wantPass(t, checkJSON(t, coverageRepo(t, 72), "unit.03"))
}

func TestUnit_AC_021_3_E1(t *testing.T) {
	wantViolation(t, checkJSON(t, coverageRepo(t, 55), "unit.03"), "unit.03", "MOD-01 的行覆盖率为 55%", "低于阈值 60%")
}

func TestUnit_AC_021_3_E2(t *testing.T) {
	dir := coverageRepo(t, 72)
	editFile(t, dir, "docs/architecture/overview.md", "- Core: true\n- Coverage: 60\n", "- Core: true\n")
	wantViolation(t, checkJSON(t, dir, "unit.03"), "unit.03", "MOD-01 的 Core 为 true，但没有设置 Coverage")
}

func TestUnit_AC_021_3_B1(t *testing.T) {
	wantPass(t, checkJSON(t, coverageRepo(t, 60), "unit.03"))
}
