package app_test

import "testing"

const suppDebt = "# 技术债清单\n\n- Status: Draft\n\n### DEBT-005 旧问题\n- Status: Resolved\n- Module: MOD-01\n\n已修复。\n\n### DEBT-014 依赖未处理错误\n- Status: Open\n- Module: MOD-01\n\n暂不处理删除失败。\n"

// suppRepo 返回含技术债清单的多语言项目；file 为源文件路径，src 为内容。
func suppRepo(t *testing.T, file, src string) string {
	t.Helper()
	return writeTree(t, map[string]string{
		"harness.json": `{"languages": ["go", "python", "typescript", "java"], "sourcePaths": ["internal/**", "src/**"]}`,
		"docs/debt.md": suppDebt,
		file:           src,
	})
}

func TestSupp_AC_042_1_N1(t *testing.T) {
	dir := suppRepo(t, "internal/a/a.go", "package a\n\nimport \"os\"\n\nfunc F() {\n\tos.Remove(\"x\") //nolint:errcheck // DEBT-014\n}\n")
	wantPass(t, checkJSON(t, dir, "supp.01"))
}

func TestSupp_AC_042_1_E1(t *testing.T) {
	dir := suppRepo(t, "src/a.py", "import os\n\n\nx = 1  # noqa\n")
	v := wantViolation(t, checkJSON(t, dir, "supp.01"), "supp.01", "noqa", "没有在同一行引用 Open 的技术债")
	if v.File != "src/a.py" || v.Line != 4 {
		t.Fatalf("应给出文件与行号：%+v", v)
	}
}

func TestSupp_AC_042_1_E2(t *testing.T) {
	dir := suppRepo(t, "src/a.ts", "// @ts-ignore DEBT-005\nconst x: number = 'a'\n")
	wantViolation(t, checkJSON(t, dir, "supp.01"), "supp.01", "DEBT-005", "已关闭")
}

func TestSupp_AC_042_1_B1(t *testing.T) {
	dir := suppRepo(t, "src/a.ts", "const tip = \"use @ts-ignore with care\"\nconst raw = `eslint-disable`\n")
	wantPass(t, checkJSON(t, dir, "supp.01"))
}

// Java 的抑制注解不是注释，按代码行识别；引用 Open 的技术债时通过。
func TestSupp_AC_042_1_E1_Annotation(t *testing.T) {
	src := "class A {\n  @SuppressWarnings(\"unchecked\")\n  void f() {}\n\n  @SuppressWarnings(\"rawtypes\") // DEBT-014\n  void g() {}\n}\n"
	c := checkJSON(t, suppRepo(t, "src/A.java", src), "supp.01")
	v := wantViolation(t, c, "supp.01", "@SuppressWarnings")
	if v.Line != 2 || len(c.vs) != 1 {
		t.Fatalf("只有第 2 行的注解没有引用技术债：%+v", v)
	}
}
