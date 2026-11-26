# ADR-008 任务可关联验收标准；根目录文件的归属

- Status: Accepted
- Relates: FR-011, FR-013, FR-039, MOD-01, FLOW-01

## 背景

规划 Release 1 任务时，两条已批准的规则互相卡住：

- AC-039-1：任务 Done 时，所关联 FR 的全部用例必须通过。
- AC-011-6：跨模块任务必须引用一个调用流程，涉及的模块都在流程中。

FR-011、FR-013、FR-033、FR-037、FR-041、FR-045 等需求的实现横跨不在同一流程中的模块（例如规则在 MOD-05，skill 在 MOD-13）。按 FR 粒度拆分时，先完成的那部分任务永远到不了 Done；不拆则只能做成横跨 6 到 8 个模块的大任务，违背"任务独立、不相关改动拆开"。

另外，`go.mod`、`go.sum`、发布配置以及 `harness init` 写入的配置文件不在任何模块的路径内，包含它们的任务（task.03）和提交（commit.02）都会失败。

FLOW-01 原先没有列出 workspace 加载语言注册表和 git 仓库、规则调用语言接口这三步，与模块实际依赖不符。

## 决策

1. 任务的 Covers 可以写 FR、NFR 或 AC。完成定义只检查所关联条目涉及的用例；FR 被直接关联，或其每条 Active AC 都被关联，均视为被任务覆盖（PRD CHG-012）。
2. `go.mod`、`go.sum`、`.goreleaser.yaml`、`.github/workflows/release.yml` 归入 MOD-01 的 Paths，MOD-01 覆盖 NFR-002。
3. `harness init` 等命令写入的文件定义为生成文件（TERM-026），commit.02 与产物文件一样豁免；它们仍受门禁保护（gate.protect），内容由 cfg.01、cfg.02 校验。
4. FLOW-01 补充 `MOD-09 -> IF-004`、`MOD-09 -> IF-007`、`MOD-05 -> IF-003` 三步。

## 后果

- 任务可以按"一个验收标准 + 它涉及的模块"拆分，规则任务和 skill 任务各自能到 Done。
- task.02 的覆盖判定从"FR 被关联"变为"FR 或其全部 AC 被关联"，实现稍复杂。
- 用 FLOW-01 承载的任务可以同时涉及 MOD-01、02、03、04、05、06、07、08、09、14。
