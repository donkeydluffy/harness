package app_test

import (
	"strings"
	"testing"
)

const (
	apiGo   = "package api\n\n// Serve 启动服务。\n//\n// harness:IF-001\nfunc Serve() error {\n\treturn nil\n}\n"
	storeGo = "package store\n\n// harness:IF-002\nfunc Save(b []byte) error {\n\treturn nil\n}\n"
)

// anchorRepo 返回带架构设计与 Go 源码的项目；extra 中的文件覆盖或追加，值为空串时删除。
func anchorRepo(t *testing.T, extra map[string]string) string {
	t.Helper()
	files := traced(t, basePRD)
	files["go.mod"] = "module demo\n\ngo 1.21\n"
	files["internal/api/api.go"] = apiGo
	files["internal/store/store.go"] = storeGo
	for k, v := range extra {
		if v == "" {
			delete(files, k)
		} else {
			files[k] = v
		}
	}
	return writeTree(t, files)
}

func snapshotOK(t *testing.T, dir string) string {
	t.Helper()
	if r := runHarness(t, dir, "snapshot"); r.Code != 0 {
		t.Fatalf("harness snapshot 失败：%d %s%s", r.Code, r.Stdout, r.Stderr)
	}
	return readFile(t, dir, "docs/architecture/interfaces.txt")
}

// ---- AC-052-1 接口锚点 ----

func TestCheck_AC_052_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, anchorRepo(t, nil), "if.01"))
}

func TestCheck_AC_052_1_E1(t *testing.T) {
	dir := anchorRepo(t, map[string]string{"internal/store/store.go": "package store\n\nfunc Save(b []byte) error { return nil }\n"})
	wantViolation(t, checkJSON(t, dir, "if.01"), "if.01", "IF-002", "没有锚点")
}

func TestCheck_AC_052_1_E2(t *testing.T) {
	dir := anchorRepo(t, map[string]string{"internal/api/b.go": "package api\n\n// harness:IF-001\nfunc Other() {}\n"})
	v := wantViolation(t, checkJSON(t, dir, "if.01"), "if.01", "IF-001 有 2 个锚点")
	if !strings.Contains(v.Msg, "internal/api/api.go:5") || !strings.Contains(v.Msg, "internal/api/b.go:3") {
		t.Fatalf("应给出两处位置：%s", v.Msg)
	}
}

func TestCheck_AC_052_1_E3(t *testing.T) {
	dir := anchorRepo(t, map[string]string{"internal/api/api.go": "package api\n",
		"internal/store/serve.go": "package store\n\n// harness:IF-001\nfunc Serve() error { return nil }\n"})
	if v := wantViolation(t, checkJSON(t, dir, "if.01"), "if.01", "IF-001 属于 MOD-01"); v.File != "internal/store/serve.go" {
		t.Fatalf("位置：%+v", v)
	}
}

// ---- AC-052-2 harness snapshot ----

func TestSnapshot_AC_052_2_N1(t *testing.T) {
	if got := snapshotOK(t, anchorRepo(t, nil)); !strings.Contains(got, "IF-001\nfunc Serve() error\n") {
		t.Fatalf("快照应包含 IF-001 的签名：\n%s", got)
	}
}

func TestSnapshot_AC_052_2_B1(t *testing.T) {
	dir := anchorRepo(t, nil)
	if a, b := snapshotOK(t, dir), snapshotOK(t, dir); a != b {
		t.Fatalf("两次提取应逐字节一致：\n%s\n----\n%s", a, b)
	}
}

func TestSnapshot_AC_052_2_B2(t *testing.T) {
	dir := anchorRepo(t, nil)
	before := snapshotOK(t, dir)
	editFile(t, dir, "internal/api/api.go", "// Serve 启动服务。", "// Serve 启动 HTTP 服务。", "return nil", "return fmt.Errorf(\"x\")")
	if after := snapshotOK(t, dir); after != before {
		t.Fatalf("只改函数体与注释时快照不应变化：\n%s", after)
	}
	wantPass(t, checkJSON(t, dir, "if.02"))
}

// ---- AC-052-3 设计签名与快照一致 ----

func TestCheck_AC_052_3_N1(t *testing.T) {
	dir := anchorRepo(t, nil)
	snapshotOK(t, dir)
	wantPass(t, checkJSON(t, dir, "if.03"))
}

func TestCheck_AC_052_3_E1(t *testing.T) {
	dir := anchorRepo(t, map[string]string{"internal/store/store.go": strings.Replace(storeGo, "b []byte", "b []byte, n int", 1)})
	snapshotOK(t, dir)
	wantViolation(t, checkJSON(t, dir, "if.03"), "if.03", "IF-002", "设计为 func Save(b []byte) error", "代码为 func Save(b []byte, n int) error")
}

func TestCheck_AC_052_3_B1(t *testing.T) {
	dir := anchorRepo(t, nil)
	snapshotOK(t, dir)
	editFile(t, dir, "docs/architecture/modules/modules.md", "func Save(b []byte) error", "// Save 保存数据。\nfunc Save(\n\tb  []byte, // 数据\n) error")
	wantPass(t, checkJSON(t, dir, "if.03"))
}

// ---- AC-052-4 签名变化须登记 ----

// lockedAnchors 返回已生成快照、登记过 CHG-002 并执行过 harness lock 的项目，随后把 IF-002 的签名改为 Save(b []byte, n int) 并重新生成快照。
func lockedAnchors(t *testing.T) string {
	t.Helper()
	dir := anchorRepo(t, nil)
	snapshotOK(t, dir)
	addCHG(t, dir, "| CHG-002 | 2026-11-25 | 新增 | IF-002 | 锁定之前登记的变更不能说明之后的变化 |")
	lockOK(t, dir)
	editFile(t, dir, "internal/store/store.go", "b []byte", "b []byte, n int")
	snapshotOK(t, dir)
	return dir
}

func TestCheck_AC_052_4_N1(t *testing.T) {
	dir := lockedAnchors(t)
	addCHG(t, dir, "| CHG-003 | 2026-11-26 | 修改 | IF-002 | Breaking: Save 增加参数 n |")
	wantPass(t, checkJSON(t, dir, "if.04"))
}

func TestCheck_AC_052_4_E1(t *testing.T) {
	wantViolation(t, checkJSON(t, lockedAnchors(t), "if.04"), "if.04", "IF-002 签名有变化", "没有引用它的新 CHG", "MOD-01")
}

func TestCheck_AC_052_4_E2(t *testing.T) {
	dir := lockedAnchors(t)
	addCHG(t, dir, "| CHG-003 | 2026-11-26 | 修改 | IF-002 | Save 增加参数 n |")
	wantViolation(t, checkJSON(t, dir, "if.04"), "if.04", "IF-002", "Breaking:")
}

func TestCheck_AC_052_4_B1(t *testing.T) {
	dir := anchorRepo(t, nil)
	snapshotOK(t, dir)
	lockOK(t, dir)
	writeFile(t, dir, "internal/store/load.go", "package store\n\n// harness:IF-009\nfunc Load() ([]byte, error) { return nil, nil }\n")
	snapshotOK(t, dir)
	wantViolation(t, checkJSON(t, dir, "if.04"), "if.04", "IF-009 是新增的接口")
	addCHG(t, dir, "| CHG-002 | 2026-11-26 | 新增 | IF-009 | 新增 Load |")
	wantPass(t, checkJSON(t, dir, "if.04"))
}
