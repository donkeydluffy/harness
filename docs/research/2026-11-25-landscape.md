# 调研记录：AI 编程 harness 现状（2026-11-25）

三路调研的结论汇总。标"未验证"的条目在进入设计前需要再核实。

## 1. 现有 SDD（Spec-Driven Development）框架

| 框架 | 产物 | ID / 变更 | 确定性校验 | 任务→需求 |
|---|---|---|---|---|
| GitHub spec-kit v1.4.4 | `specs/###-feature/{spec,plan,research,data-model,tasks}.md`、`contracts/`、`memory/constitution.md` | `FR-001`、`SC-001`、`US1`、`T001`；`/clarify` 结果追加到 spec 的 Clarifications 段，无增删改 delta | 只有脚手架脚本；`/analyze` 覆盖率表由 LLM 推断 | 只标 `[US1]`，不标 FR |
| AWS Kiro | requirements.md（EARS）、design.md、`tasks.md`、`.kiro/steering/`、`.kiro/hooks/` | 直接改文件，无 delta/归档 | hooks 可跑 shell；Analyze 靠 LLM | 未验证 |
| OpenSpec v2.2.0 | `openspec/specs/<cap>/spec.md` + `openspec/changes/<id>/{proposal,design,tasks}.md` + delta，归档到 `changes/archive/` | 需求无数字 ID，按标题定位；`ADDED/MODIFIED/REMOVED/RENAMED` delta | `openspec validate --strict` 是真校验器（SHALL/MUST、Scenario、delta 合法性） | 任务不引用需求 |
| BMAD v6/v7 | PRD/SRS、initiative/epic/story、架构决策 `AD-n` | `FR-<n>`、`CAP-<n>`；有 withdrawn 状态 | check_ids.py（悬空/重复/撤回仍引用）、lint_spine.py（ADR 字段），但退出码恒 0，交给 LLM 决定 | 部分 |
| Agent OS v3 / Tessl | standards 注入 / spec registry | — | 未见校验器 | — |

来源：github.com/github/spec-kit、kiro.dev/docs/specs、github.com/Fission-AI/OpenSpec、github.com/bmad-code-org/BMAD-METHOD、github.com/buildermethods/agent-os、tessl.io。

**共同缺口（也就是本 harness 的定位）**：
1. 没有任何一个做到"每条 FR 至少被一个设计元素覆盖、每个任务至少关联 FR + 设计元素"的确定性覆盖矩阵，且以非 0 退出码阻断。
2. 都没有模块依赖规则和接口变更通知。
3. 术语表（统一语言）的一致性只有 LLM 判断。
4. UI token 约束都只是 prompt。

可借鉴：OpenSpec 的 delta + archive 变更模型和 strict 校验；BMAD 的 withdrawn 状态与 ADR 字段（Binds/Prevents/Rule）；spec-kit 的 `/clarify` 每轮最多 5 问、答案落盘到 Clarifications。

## 2. 确定性检查工具

- **需求追溯**：OpenFastTrace 要 JVM、ID 格式固定为 `type~name~rev`；Doorstop 每条需求一个 YAML；StrictDoc、sphinx-needs 不读普通 Markdown。结论：自研零依赖检查器（正则抽 ID → 建图 → 孤儿/悬空/未覆盖），变更检测用"每条需求内容哈希 + 已提交 lock 文件"。
- **模块依赖**：dependency-cruiser（JS/TS）、import-linter（Python）、go-arch-lint（Go）、ArchUnit（Java）、Nx tags、Bazel visibility。没有现成的跨语言方案。结论：一份语言中立的模块声明，生成各语言工具配置；或自研粗粒度 import 抽取。
- **接口变更**：oasdiff（OpenAPI）、buf breaking（protobuf）、api-extractor（TS `.api.md`）、cargo-semver-checks、griffe（Python）。推荐模式：接口快照提交进仓库 → CI 重新生成并 diff → 破坏性变更要求显式确认（commit trailer / 变更记录）→ CODEOWNERS 审批。
- **UI token**：W3C DTCG 规范 2025-10-28 发布首个稳定版 2025.10；Style Dictionary 支持 DTCG；Tailwind v4 用 `@theme { --color-*: initial; }` 清空默认色板；eslint-plugin-better-tailwindcss 禁任意值；stylelint-declaration-strict-value 禁裸色值。

未验证：oasdiff/buf/api-extractor 的具体 flag；Nx 配置写法；Style Dictionary v5 是否以 DTCG 2025.10 为基准。

## 3. Claude Code 扩展点（code.claude.com/docs）

- **Hooks**：PreToolUse 可硬拦截（exit 2 或 `permissionDecision: "deny"` JSON）；PostToolUse 只能事后反馈；Stop 可阻止结束（`decision: "block"`）。配置在 `.claude/settings.json` 或插件 `hooks/hooks.json`。注意 Bash 也能写文件，拦截写入时要覆盖 Bash。
- **Skills**：`.claude/skills/<name>/SKILL.md`，按 description 自动触发或 `/name` 手动触发；slash commands 已并入 skills。
- **Rules**：`.claude/rules/*.md` 是官方功能，`paths:` frontmatter 按 glob 在读写匹配文件时加载。
- **Plugin**：`.claude-plugin/plugin.json` + skills/agents/hooks，可通过 marketplace 分发。
- **跨工具**：Claude Code 在没有 CLAUDE.md 时读 AGENTS.md；可移植做法是 CLAUDE.md 里 `@AGENTS.md`。Codex 等的 hook 机制未查证。

未验证：hook stdin 字段名（如 `tool_input.file_path`）、`stop_hook_active` 防死循环字段、子代理工具调用是否触发主会话 PreToolUse。

## 4. Codex / Cursor 门禁能力与 Windows

| 代理 | 写前拦截 | 写后 | 结束时 | 配置 |
|---|---|---|---|---|
| Claude Code | PreToolUse deny（Edit/Write/Bash） | PostToolUse 只能反馈 | Stop `decision:block` 真阻止 | `.claude/settings.json` |
| Codex CLI/App | PreToolUse deny，matcher `apply_patch`；路径需从 `tool_input.command` 的补丁头 `*** Update File:` 解析 | PostToolUse 只能替换结果 | Stop block 变成续跑提示，不是硬阻止 | `.codex/hooks.json`；项目需 trust，hook 按哈希审核 |
| Cursor | `preToolUse` `permission:"deny"`，matcher `Write|StrReplace|Delete`（工具名来自社区，需实测），需 `failClosed:true` | `afterFileEdit` 无输出，挡不住 | `stop` 只能返回 `followup_message`，受 `loop_limit` 限制 | `.cursor/hooks.json` |

- Codex 官方说明 hooks "不是完整的强制边界"；IDE 扩展和 `codex exec` 是否执行 hooks 未验证。
- Cursor 论坛报告过 deny 不生效、Windows 改 hooks.json 后 hook 失效的 bug。
- 指令文件：Codex 读 AGENTS.md（默认上限 32 KiB），skills 在 `.agents/skills`；Cursor 读 `.cursor/rules/*.mdc`、AGENTS.md，skills 在 `.agents/skills` 或 `.cursor/skills`。
- Windows：Claude Code 用 exec 形式直接调 .exe 最稳；Codex 用 `commandWindows`，有报告 exit 2 被当失败放行，应改用 JSON deny；Cursor 经 PowerShell 执行。
- 回退三层：写后用 git 比对回滚未授权改动 → 结束时把失败反馈给代理 → git pre-commit + CI 兜底。

来源：learn.chatgpt.com/docs/hooks、cursor.com/docs/agent/hooks、code.claude.com/docs/en/hooks，以及 forum.cursor.com 相关帖子。

## 5. 测试追溯与测试质量（确定性）

- **ID 绑定**：唯一跨框架保证保留的是 JUnit XML 中 testcase 的 `name`；部分框架还能写 `<property>`（pytest `record_property`、Vitest `annotate`、GoogleTest `RecordProperty`）。Surefire 不输出 JUnit5 reportEntry，`@DisplayName` 在 XML 中不稳定。Go 用 `go test -json`（或 gotestsum 转 JUnit），ID 放测试名或子测试名。结论：约定 ID 写进测试名，property 作为补充。
- **静态 + 运行时双检**：静态扫描建立"声明的映射"，JUnit XML 证明"至少一个 testcase 为 passed"；skipped 不计覆盖；静态有、报告无即失败。
- **边界覆盖**：在 PRD 的每条 AC 下显式列出用例 `AC-001-1/N1`（正常）、`/B1`（边界）、`/E1`（错误），规则为每条 AC 至少一个 N 和一个 B 或 E，每个用例 ID 至少被一个通过的集成测试引用。借鉴 OpenFastTrace 的 `Needs: utest, itest` 与修订号过期标记。
- **接口测试真的调用了接口**：按测试分覆盖在 Python（coverage contexts）可行、Java 需 hack、Go/JS 不成熟。跨语言的确定性替代：只运行标注 IF-n 的测试并采集覆盖率，确认 `harness:IF-n` 锚点函数入口行被执行。
- **覆盖率阈值**：各语言覆盖率工具都能按路径统计，可按 MOD-n 设阈值。
- **变异测试**：PIT 最成熟（diff 模式需商业插件，开源免费）；Stryker 有 incremental；Mull 的 diff 模式为 experimental；Go 的 gremlins 未到 1.0 且慢。结论：只对 diff 运行，作为可选告警，不做默认硬门禁。
- **防 AI 作弊**：已知手法有删除或放宽断言、加 skip/xfail、恒真对象（arxiv 2511.21654、2511.18397）。确定性规则：每个测试至少一个断言、断言不能是常量、被追溯的测试不能 skipped、IF-n 的测试不能 mock IF-n 所在模块、相对基线断言数减少或测试删除或新增 skip 需人工确认；现成 lint：eslint-plugin-jest 的 `expect-expect` / `no-disabled-tests`、PMD `JUnitTestsShouldIncludeAssert`。

## 6. Harness engineering 实践中的其他维度

来源：anthropic.com/engineering/effective-harnesses-for-long-running-agents、openai.com/index/harness-engineering（2026-02-11）、martinfowler.com/articles/exploring-gen-ai/harness-engineering.html（2026-04-02）、anthropic.com/engineering/harness-design-long-running-apps（2026-03-24）、code.claude.com/docs/en/best-practices、thoughtworks.com/radar/techniques/harness-quality-evaluation、mitchellh.com/writing/my-ai-adoption-journey；Radar Vol.34 为第三方摘要，未核对原文。

按"最可能缺"排序：
1. **让代理运行和观察应用**：按 worktree 启动应用，浏览器截图/DOM，可查询的日志、指标、trace；性能预算写成可验证的目标。
2. **会话状态交接**：progress 文件、固定开场流程（init 脚本 + 冒烟测试）、结束时 commit 并更新进度。
3. **安全执行边界**：沙箱、deny 规则（如 `.env`）、secret 扫描（失败即阻断）、MCP/skill 来源审查。
4. **依赖引入闸门**：新增依赖需 allowlist 或 ADR；防 AI 幻觉包名被抢注（slopsquatting）：校验包存在、包龄、冷却期；lockfile 哈希；license/SBOM（推断，来源未专门讨论）。
5. **熵清理**：定期文档园丁、技术债登记、文档新鲜度与链接 lint。
6. **独立评审**：新上下文的评审代理对照 PRD/AC 审 diff，跳过风格问题。
7. **测试质量传感器**：变异测试；数据库迁移目录的写保护。
8. **度量和测试 harness 本身**：每个 hook 有测试证明会触发，统计首次通过率、返工率，监控从不触发的检查，模型升级后回测并删除多余部件；lint 报错附修复说明。

其他：入口指令文件（AGENTS.md）保持约 100 行、只做地图；重复犯错就补规则或工具；成本上限（max turns / budget）。
