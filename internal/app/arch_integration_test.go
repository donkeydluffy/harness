package app_test

import (
	"strings"
	"testing"
)

const archOverview = `# 架构总览

- Status: Draft

### MOD-01 api
- Status: Active
- Paths: internal/api/**
- Public: internal/api/api.go
- Depends: MOD-02
- Interfaces: IF-001
- Covers: FR-001
- Core: true
- Coverage: 80

### MOD-02 store
- Status: Active
- Paths: internal/store/**
- Public: internal/store/store.go
- Depends:
- Interfaces: IF-002
- Covers: FR-002
- Core: false
- Coverage: 70

### EXT-01 新增存储
- Status: Active
- Interface: IF-002
- Paths: internal/store/<实现>/**
- Closed: MOD-01
- Relates: NFR-001
`

const archModules = "# 模块详细设计\n\n- Status: Draft\n\n### IF-001 Serve\n- Module: MOD-01\n- Covers: FR-001\n\n```go\n// IF-001\nfunc Serve() error\n```\n\n### IF-002 Save\n- Module: MOD-02\n- Covers: FR-002\n\n```go\n// IF-002\nfunc Save(b []byte) error\n```\n"

const archFlows = `# 调用流程

- Status: Draft

### FLOW-01 保存
- Covers: FR-001
- Steps: MOD-01 -> IF-001, MOD-01 -> IF-002
`

// arch 返回架构夹具；edits 为 文件名、旧文本、新文本 的三元组序列。
func arch(t *testing.T, edits ...string) map[string]string {
	t.Helper()
	files := map[string]string{
		"docs/architecture/overview.md":        archOverview,
		"docs/architecture/modules/modules.md": archModules,
		"docs/architecture/flows.md":           archFlows,
	}
	for i := 0; i+2 < len(edits); i += 3 {
		p := "docs/architecture/" + edits[i]
		if !strings.Contains(files[p], edits[i+1]) {
			t.Fatalf("替换目标不存在：%s %q", p, edits[i+1])
		}
		files[p] = strings.Replace(files[p], edits[i+1], edits[i+2], 1)
	}
	return files
}

func checkArch(t *testing.T, files map[string]string) checked {
	t.Helper()
	return checkJSON(t, writeTree(t, files), "arch")
}

// ---- AC-006-1 模块必填字段、路径不重叠 ----

func TestCheck_AC_006_1_N1(t *testing.T) { wantPass(t, checkArch(t, arch(t))) }

func TestCheck_AC_006_1_E1(t *testing.T) {
	c := checkArch(t, arch(t, "overview.md", "- Paths: internal/api/**\n", ""))
	wantViolation(t, c, "arch.01", "MOD-01", "Paths")
}

func TestCheck_AC_006_1_B1(t *testing.T) { wantPass(t, checkArch(t, arch(t))) } // MOD-02 的 Depends 为空

func TestCheck_AC_006_1_E2(t *testing.T) {
	c := checkArch(t, arch(t, "overview.md", "- Paths: internal/store/**", "- Paths: internal/api/store/**"))
	wantViolation(t, c, "arch.01", "MOD-02", "internal/api/store/**", "重叠")
}

// ---- AC-006-2 接口 ----

func TestCheck_AC_006_2_N1(t *testing.T) { wantPass(t, checkArch(t, arch(t))) }

func TestCheck_AC_006_2_E1(t *testing.T) {
	c := checkArch(t, arch(t, "modules/modules.md", "- Module: MOD-02", "- Module: MOD-09"))
	wantViolation(t, c, "arch.01", "IF-002", "MOD-09 不存在")
}

func TestCheck_AC_006_2_E2(t *testing.T) {
	c := checkArch(t, arch(t, "modules/modules.md", "```go\n// IF-002\nfunc Save(b []byte) error\n```\n", ""))
	wantViolation(t, c, "arch.01", "IF-002", "缺少签名")
}

// ---- AC-006-3 调用流程 ----

func TestCheck_AC_006_3_N1(t *testing.T) { wantPass(t, checkArch(t, arch(t))) }

func TestCheck_AC_006_3_B1(t *testing.T) {
	wantPass(t, checkArch(t, arch(t, "flows.md", "MOD-01 -> IF-001, MOD-01 -> IF-002", "MOD-01 -> IF-002")))
}

func TestCheck_AC_006_3_E1(t *testing.T) {
	c := checkArch(t, arch(t, "flows.md", "- Steps: MOD-01 -> IF-001, MOD-01 -> IF-002", "- Steps:"))
	wantViolation(t, c, "arch.01", "FLOW-01", "Steps")
}

// ---- AC-006-4 依赖无环 ----

func TestCheck_AC_006_4_N1(t *testing.T) { wantPass(t, checkArch(t, arch(t))) }

func TestCheck_AC_006_4_E1(t *testing.T) {
	third := "\n### MOD-03 cache\n- Status: Active\n- Paths: internal/cache/**\n- Public: internal/cache/cache.go\n- Depends: MOD-01\n- Interfaces:\n- Covers: FR-001\n- Core: false\n- Coverage: 60\n"
	c := checkArch(t, arch(t,
		"overview.md", "- Depends:\n- Interfaces: IF-002", "- Depends: MOD-03\n- Interfaces: IF-002",
		"overview.md", "\n### EXT-01", third+"\n### EXT-01"))
	wantViolation(t, c, "arch.02", "MOD-01 -> MOD-02 -> MOD-03 -> MOD-01")
}

func TestCheck_AC_006_4_B1(t *testing.T) {
	c := checkArch(t, arch(t, "overview.md", "- Depends:\n- Interfaces: IF-002", "- Depends: MOD-02\n- Interfaces: IF-002"))
	wantViolation(t, c, "arch.02", "MOD-02 -> MOD-02")
}

// ---- AC-006-5 流程步骤与依赖一致 ----

func TestCheck_AC_006_5_N1(t *testing.T) { wantPass(t, checkArch(t, arch(t))) }

func TestCheck_AC_006_5_E1(t *testing.T) {
	c := checkArch(t, arch(t, "flows.md", "MOD-01 -> IF-002", "MOD-01 -> IF-009"))
	wantViolation(t, c, "arch.03", "FLOW-01", "第 2 步", "IF-009 不存在")
}

func TestCheck_AC_006_5_E2(t *testing.T) {
	c := checkArch(t, arch(t, "flows.md", "MOD-01 -> IF-001, MOD-01 -> IF-002", "MOD-02 -> IF-002, MOD-02 -> IF-001"))
	wantViolation(t, c, "arch.03", "FLOW-01 第 2 步", "MOD-01 不在 MOD-02 的 Depends 中")
}

func TestCheck_AC_006_5_B1(t *testing.T) {
	wantPass(t, checkArch(t, arch(t, "flows.md", "MOD-01 -> IF-001, MOD-01 -> IF-002", "MOD-02 -> IF-002")))
}

// ---- AC-033-2 扩展点字段 ----

func TestCheck_AC_033_2_N1(t *testing.T) { wantPass(t, checkArch(t, arch(t))) }

func TestCheck_AC_033_2_E1(t *testing.T) {
	c := checkArch(t, arch(t, "overview.md", "- Interface: IF-002", "- Interface: IF-077"))
	wantViolation(t, c, "arch.01", "EXT-01", "IF-077 不存在")
}

func TestCheck_AC_033_2_E2(t *testing.T) {
	c := checkArch(t, arch(t, "overview.md", "- Closed: MOD-01", "- Closed: MOD-01, MOD-08"))
	wantViolation(t, c, "arch.01", "EXT-01", "MOD-08 不存在")
}
