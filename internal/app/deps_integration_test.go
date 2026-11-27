package app_test

import (
	"strings"
	"testing"
)

const depsOverview = "# 架构总览\n\n- Status: Draft\n\n" +
	"### MOD-01 api\n- Status: Active\n- Paths: internal/api/**\n- Public: internal/api/api.go\n- Depends:\n- Interfaces:\n- Covers: FR-001\n- Core: false\n\n" +
	"### MOD-02 store\n- Status: Active\n- Paths: internal/store/**\n- Public: internal/store/store.go\n- Depends: MOD-01\n- Interfaces:\n- Covers: FR-001\n- Core: false\n\n" +
	"### MOD-03 cache\n- Status: Active\n- Paths: internal/cache/**\n- Public: internal/cache/cache.go\n- Depends:\n- Interfaces:\n- Covers: FR-001\n- Core: false\n"

// depsRepo 返回三个模块的 Go 项目：MOD-02 store import MOD-01 api 的公开包；extra 中的文件覆盖或追加。
func depsRepo(t *testing.T, extra map[string]string) string {
	t.Helper()
	files := map[string]string{
		"harness.json":                  `{"languages": ["go"], "sourcePaths": ["internal/**"]}`,
		"go.mod":                        "module demo\n\ngo 1.21\n",
		"docs/architecture/overview.md": depsOverview,
		"internal/api/api.go":           "package api\n\nfunc Serve() {}\n",
		"internal/api/route/route.go":   "package route\n\nfunc R() {}\n",
		"internal/store/store.go":       "package store\n\nimport \"demo/internal/api\"\n\nfunc Save() { api.Serve() }\n",
		"internal/cache/cache.go":       "package cache\n",
	}
	for k, v := range extra {
		files[k] = v
	}
	return writeTree(t, files)
}

func importing(pkg string, paths ...string) string {
	return "package " + pkg + "\n\nimport (\n\t\"" + strings.Join(paths, "\"\n\t\"") + "\"\n)\n"
}

// ---- AC-053-1 跨模块 import 须在 Depends 中 ----

func TestDeps_AC_053_1_N1(t *testing.T) {
	wantPass(t, checkJSON(t, depsRepo(t, nil), "deps.01"))
}

func TestDeps_AC_053_1_E1(t *testing.T) {
	dir := depsRepo(t, map[string]string{"internal/api/api.go": "package api\n\nimport _ \"demo/internal/store\"\n\nfunc Serve() {}\n"})
	v := wantViolation(t, checkJSON(t, dir, "deps.01"), "deps.01", "MOD-01 import 了 MOD-02", "不在 MOD-01 的 Depends 中")
	if v.File != "internal/api/api.go" || v.Line != 3 {
		t.Fatalf("应给出文件与行号：%+v", v)
	}
}

func TestDeps_AC_053_1_B1(t *testing.T) {
	dir := depsRepo(t, map[string]string{
		"internal/store/disk/disk.go": "package disk\n",
		"internal/store/extra.go":     importing("store", "fmt", "github.com/google/uuid", "demo/internal/store/disk"),
	})
	wantPass(t, checkJSON(t, dir, "deps.01"))
}

func TestDeps_AC_053_1_B2(t *testing.T) {
	dir := depsRepo(t, map[string]string{"internal/api/api_test.go": importing("api", "testing", "demo/internal/store")})
	wantPass(t, checkJSON(t, dir, "deps.01"))
}

// ---- AC-053-2 未归属文件 ----

func TestDeps_AC_053_2_N1(t *testing.T) {
	wantPass(t, checkJSON(t, depsRepo(t, nil), "deps.02"))
}

func TestDeps_AC_053_2_E1(t *testing.T) {
	dir := depsRepo(t, map[string]string{"internal/x/y.go": "package x\n"})
	if v := wantViolation(t, checkJSON(t, dir, "deps.02"), "deps.02", "internal/x/y.go", "不属于任何模块"); v.File != "internal/x/y.go" {
		t.Fatalf("应列出该文件：%+v", v)
	}
}

func TestDeps_AC_053_2_B1(t *testing.T) {
	wantPass(t, checkJSON(t, depsRepo(t, map[string]string{"scripts/a.go": "package main\n"}), "deps.02"))
}

// ---- AC-053-3 过时的 Depends ----

func TestDeps_AC_053_3_N1(t *testing.T) {
	wantPass(t, checkJSON(t, depsRepo(t, nil), "deps.03"))
}

func TestDeps_AC_053_3_E1(t *testing.T) {
	dir := depsRepo(t, map[string]string{"docs/architecture/overview.md": strings.Replace(depsOverview, "- Depends: MOD-01\n", "- Depends: MOD-01, MOD-03\n", 1)})
	wantViolation(t, checkJSON(t, dir, "deps.03"), "deps.03", "MOD-02 的 Depends 声明了 MOD-03", "过时声明")
}

func TestDeps_AC_053_3_B1(t *testing.T) {
	dir := depsRepo(t, map[string]string{
		"docs/architecture/overview.md": strings.Replace(depsOverview, "- Depends: MOD-01\n", "- Depends: MOD-01, MOD-03\n", 1),
		"internal/store/store_test.go":  importing("store", "testing", "demo/internal/cache"),
	})
	wantViolation(t, checkJSON(t, dir, "deps.03"), "deps.03", "MOD-02 的 Depends 声明了 MOD-03")
}

// ---- AC-035-2 跨模块 import 指向 Public ----

func TestDeps_AC_035_2_N1(t *testing.T) {
	wantPass(t, checkJSON(t, depsRepo(t, nil), "deps.04"))
}

func TestDeps_AC_035_2_E1(t *testing.T) {
	dir := depsRepo(t, map[string]string{"internal/store/route.go": importing("store", "demo/internal/api/route")})
	v := wantViolation(t, checkJSON(t, dir, "deps.04"), "deps.04", "demo/internal/api/route", "不在 MOD-01 的 Public")
	if v.File != "internal/store/route.go" || v.Line != 4 {
		t.Fatalf("应给出文件与行号：%+v", v)
	}
}

func TestDeps_AC_035_2_B1(t *testing.T) {
	dir := depsRepo(t, map[string]string{"internal/api/use.go": importing("api", "demo/internal/api/route")})
	wantPass(t, checkJSON(t, dir, "deps.04"))
}

func TestDeps_AC_035_2_B2(t *testing.T) {
	dir := depsRepo(t, map[string]string{"internal/store/store_test.go": importing("store", "testing", "demo/internal/api/route")})
	wantPass(t, checkJSON(t, dir, "deps.04"))
}
