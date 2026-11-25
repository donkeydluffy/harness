# MOD-13 scaffold 详细设计

- Status: Draft

## 接口

### IF-020 Generate
- Module: MOD-13
- Covers: FR-001, FR-002, FR-013, FR-016, FR-043

```go
// IF-020
type InitOptions struct{ Languages []string; Agents []string; CI string; Binary string }
type Result struct{ Written, Skipped []string }
// Generate 写入模板；已存在且内容不同的文件跳过并列入 Skipped（AC-001-2）。
func Generate(root string, opt InitOptions) (Result, error)
```

### IF-021 Expected
- Module: MOD-13
- Covers: FR-041

```go
// IF-021
// Expected 返回"生成文件"的期望内容（hook 配置、git hook、CI 片段），供规则 cfg.02 比对。
func Expected(s *workspace.Snapshot) (map[string][]byte, error)
```

## 规则

本模块注册规则 `cfg.02`：生成文件与 IF-021 的期望内容一致。

## 模板

`internal/scaffold/templates/`：

- `artifacts/`：PRD、架构总览、formats、ADR、tasks、debt、progress 骨架。
- `skills/<阶段>/SKILL.md`：clarify（需求）、change（需求变更）、architect（架构）、plan（规划）、implement（实现）、test（测试）。每个 skill 的 frontmatter 声明结束前运行的 `harness check` 子命令。
- `agents/<代理>/`：hook 配置模板。
- `ci/`：GitHub Actions、GitLab CI。
- `githooks/`：commit-msg、pre-commit。
