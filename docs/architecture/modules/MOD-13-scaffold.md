# MOD-13 scaffold 详细设计

- Status: Approved

## 接口

### IF-020 Generate
- Module: MOD-13
- Covers: FR-001, FR-002, FR-013, FR-016, FR-043

```go
// IF-020
type InitOptions struct {
	Languages, Agents []string
	CI, Binary        string
	Version           string // 当前 harness 版本（IF-037）
	CIVersion         string // --ci-version；为空时用 Version
}
type Result struct{ Written, Skipped []string }
// Generate 写入模板；已存在且内容不同的文件跳过并列入 Skipped（AC-001-2）。
// 各语言工具的配置文件（AC-014-6）只在缺失时写入，不属于生成文件。
// CI 固定的版本不是发布版本号（IF-037 Release）且项目不是 harness 自身仓库时返回错误，不写任何文件（AC-051-2）。
func Generate(root string, opt InitOptions) (Result, error)
```

### IF-021 Expected
- Module: MOD-13
- Covers: FR-041

```go
// IF-021
// Expected 返回"生成文件"的期望内容（hook 配置、git hook、CI 片段），供规则 cfg.02 比对。
// CI 片段按其中固定的版本生成期望内容（AC-051-3）。
func Expected(s *workspace.Snapshot) (map[string][]byte, error)
```

## 规则

本模块注册规则 `cfg.02`：生成文件与 IF-021 的期望内容一致；本地 harness 是发布版且与 CI 固定版本不同时，报告两个版本并提示重新生成（AC-051-3）。

## 模板

`internal/scaffold/templates/`：

- `artifacts/`：PRD、架构总览、formats、ADR、tasks、debt、progress 骨架。
- `skills/<名称>/SKILL.md`：clarify（需求）、change（需求变更）、architect（设计）、plan（任务）、implement（实现）、test（测试）。每个 skill 的 frontmatter 声明结束前运行的 `harness check` 子命令；正文不再提阶段（FR-048）。
- `agents/<代理>/`：hook 配置模板。
- `ci/`：GitHub Actions、GitLab CI。用户项目下载固定版本的发布归档并用 checksums.txt 校验（AC-051-1，ADR-012），依次运行 `harness check lint`、`harness check smoke`、`harness check`、`harness check commits`（AC-014-7、AC-055-2）；harness 自身仓库仍从源码构建。
- `langs/<语言>/files/`：工具配置文件（如 `ruff.toml`、`.prettierrc`、`.clang-format`）。
- `githooks/`：commit-msg、pre-commit。
