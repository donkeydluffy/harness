# 进度记录

- Status: Draft
- Updated: 2026-11-26

任务变为 Done 时追加一条记录（AC-024-2）：标题为 `### <日期> T-n`，正文写结论与遗留问题。会话结束检查未通过的项也记在这里。

## 记录

### 2026-11-26 T-001
- 结论：Done。`harness check [规则或前缀] [--json]` 可用，退出码 0/1/2；model、workspace、engine、prd.01、prd.02 完成。17 个用例（AC-003-1、AC-003-2、AC-003-4、NFR-004）的集成测试全部通过（`go test ./...`）；故意删掉 prd.02 的 Priority 检查后 3 个测试失败，恢复后通过。本仓库 `harness check` 0 条违规。
- 调整：AC-003-5 的追溯部分依赖 trace.01，移到 T-006 验收；T-001 范围补充 `internal/rules/builtin/util.go` 与 `registry_test.go`。
- 遗留：DEBT-001 模块路径；DEBT-002 解析诊断未输出。task.08 尚未实现，本任务的完成定义由 `go test` 人工核对，T-015 上线后回溯校验。

### 2026-11-26 T-002
- 结论：Done。lock 模块（规范化、哈希、键排序且逐字节稳定的读写）、workspace 加载锁文件并计算锁定条目、`harness lock`（全量检查通过才写，保留批准与阶段）、prd.03、prd.08。17 个用例（FR-004、AC-003-3、AC-019-3）集成测试通过；把 prd.08 的动作比对改为恒通过后 3 个测试失败。本仓库已生成首份 `harness.lock`，此后需求变化必须登记 CHG。
- 调整：formats.md 补充锁文件条目定义、用例删除标记（文字以 Removed 开头）、进度记录标题 `### <日期> T-n`（原标题被解析为重复的任务实体，由 model 的仓库自检测试发现）。formats.md 因此回到 Draft，待用户重新确认。
- 遗留：无新增技术债。

### 2026-11-26 T-003
- 结论：Done。prd.04、prd.05、prd.06、prd.07、prd.09、prd.10。23 个用例（FR-002、AC-019-1、AC-019-2、FR-040）集成测试通过；分别放宽每轮问题上限、去掉行内代码屏蔽后，对应测试失败。
- 自举发现：prd.09 命中了 PRD 自身 FR-040 引用的词表，改为代码格式（CHG-013，语义不变）。PRD 内容因此变化，rc3 的批准随之失效，T-013 上线后需重新批准。
- 调整：所有内置规则任务的范围加入共用的 `internal/rules/builtin/util.go`。
- 过程问题：一次提交写到 693 行，超过 400 行上限，已拆成三个提交。T-012（commit.03）上线前，提交前先人工核对行数。
- 遗留：DEBT-002 部分处理。

### 2026-11-26 T-004
- 结论：Done。term.01、term.02。7 个用例（FR-005）集成测试通过；去掉统一语言表格行的豁免后全部检查类测试失败。
- 自举发现：term.01 命中 PRD 中 FR-005 的用例原文和 FR-015 的 `组件文件`，分别改为代码格式和 `界面代码文件`（CHG-014）。
- 调整：formats.md 增加可选字段 Domain（MOD、IF），定义"标注为领域概念"的方式。
- 遗留：无。

### 2026-11-26 T-005
- 结论：Done。arch.01（模块、接口、流程、扩展点字段，模块路径不重叠，模块接口列表与接口所属模块一致）、arch.02（依赖环，含自依赖）、arch.03（流程步骤与依赖一致）。20 个用例（FR-006、AC-033-2）集成测试通过；去掉"调用自身接口"的豁免后 B1 等测试失败。本仓库架构 0 条违规。
- 过程问题：又一次提交超过 400 行（511 行），已拆分；之后提交前用 scratchpad 中的 commitsize.sh 核对。
- 遗留：DEBT-002 继续缩小范围。

### 2026-11-26 T-006
- 结论：Done。trace.01（含 AC-003-5 的 Release 范围）、trace.02、trace.03、`harness trace --format md|json`。18 个用例（FR-007、AC-003-5）集成测试通过；让 Superseded 的 ADR 计入覆盖后 AC-007-1/B1 失败。本仓库 0 条违规。
- 调整：锁文件测试的夹具补充架构（`harness lock` 先跑全量检查，trace.01 上线后纯 PRD 夹具不再通过）；新增需求夹具改为 Release 2。
- 原型脚本：selfcheck.py 的追溯、悬空、环、流程检查已由 trace.*、arch.* 取代；扩展点关联检查保留到 T-010（arch.04），扇出检查属于 Release 2（FR-036）。taskcheck.py 保留到 T-011。
- 遗留：无。

### 2026-11-26 T-007
- 结论：Done。vcs（git 命令实现，尾注解析）、workspace 加载仓库与提交范围、adr.01、adr.02、adr.03。10 个用例（FR-010）集成测试通过；让 adr.03 忽略缺少的 ADR 引用后 AC-010-3/E1 失败。
- 设计澄清：MOD-05 中"输入缺失"的口径改为"指定了提交范围却不是 git 仓库、或需要测试报告却没有报告"；未指定范围时 git 规则检查暂存区。否则任何非 git 目录的全量检查都会失败，与 AC-001-1 冲突。MOD-05 因此回到 Draft，待确认。
- 规则样本：registry_test 支持 `repo/01`、`repo/02`、`repo/staged` 形式的 git 历史样本。
- 遗留：DEBT-002 中 ADR 文件的诊断已由 adr.01 报告。

### 2026-11-26 T-008
- 结论：Done。lang（注册表、Go 语言包，演示 EXT-01）、testreport（JUnit、go test -json）、workspace 加载源码与测试报告、test.01、test.02。13 个用例（FR-020）集成测试通过；让 test.02 把单元测试计入覆盖后 AC-020-4/E1 失败，让 test.01 忽略不存在的用例后 AC-020-1/E1 失败。
- 设计澄清：`harness lock` 原为"全量检查通过后"写锁文件，加入 test.02 后需求阶段无法锁定 PRD。engine 增加可选接口 `Needer`，需要源码、测试报告或 git 的规则声明输入，lock 只运行其余规则。MOD-02、MOD-04 因此回到 Draft，待确认。
- 自检：`go test -json ./... > .harness/reports/go.jsonl` 后全量检查有 192 条 test.02 违规，全部属于后续任务的需求；T-001 至 T-008 关联的用例都有通过的集成测试。`.harness/reports/` 加入 .gitignore。

### 2026-11-26 T-009
- 结论：Done。debt.01、debt.02、debt.03，注释按行拆分后检查，编号按数值归一（`DEBT-2` 与 `DEBT-002` 相同）。10 个用例（FR-031）集成测试通过；分别破坏三条规则的判定后 AC-031-1/E1、AC-031-2/E2、AC-031-3/E1 失败。
- 实现取舍：debt.03 只看注释中的引用，测试夹具字符串里出现的编号不算引用，否则这类测试会让技术债永远无法关闭。
- 用例说明：AC-031-2/B1 原文是 `# FIXME T-5`，Release 1 只内置 Go 语言包，集成测试用 `// FIXME T-5` 验证同一行为。

### 2026-11-26 T-010
- 结论：Done。prd.11、arch.04、arch.05。T-010 关联的 AC-033-1、AC-033-3、AC-033-4 共 8 个用例集成测试通过；prd.11 同时实现 AC-033-5 的检查，3 个用例的测试也已写好（该 AC 由 T-020 关联）。逐一破坏各判定分支后对应的 E 用例失败。
- PRD 修改：NFR-006、NFR-007 缺少 CHG-010 声称已补的 Adds、Allowed-Paths、Closed-Modules 字段，按正文补齐并登记 CHG-015。PRD 内容在批准后又有变化，待用户重新确认。
- 实现取舍：arch.04 与 arch.05 不处理不存在的引用，由 trace.03 报告，避免同一问题报三次。
- 自检：arch.05 报告 EXT-04（新增阶段，NFR-008）没有示例扩展测试，属于 T-014 与 T-023 的范围。selfcheck.py 的扩展点检查已由 arch.04 取代。

### 2026-11-26 T-011
- 结论：Done。task.01 ~ task.06。26 个用例（AC-011-1、2、3、4、6，AC-034-2）集成测试通过；逐一破坏各判定分支后对应用例失败。AC-011-3/E1 补了一个断言：路径属于存在但未关联的模块时也失败，原用例的路径不属于任何模块，单靠它测不出关联判断。
- 实现取舍：Removed 标记只用于需求和用例；AC-003-3 的正文以"Removed"开头，按用例规则判断会误报为已删除。规则样本改用最小夹具，完整夹具复制 12 份超出提交行数上限。
- 自检：`harness check task` 对本仓库 0 条违规，taskcheck.py 原型不再需要。
- 遗留：DEBT-002 仍为 Open，剩 PRD 中与用例无关的诊断和进度文件的诊断。

### 2026-11-26 T-012
- 结论：Done。commit.01 ~ commit.04；`harness check` 支持 `--commits <范围>`、`check commits <范围>` 与 `--msg-file`（commit-msg hook 传入即将形成的提交信息）。14 个用例（AC-013-1、2、4，AC-034-1）集成测试通过；逐一破坏各判定分支后对应用例失败。
- 设计澄清：pre-commit 时提交信息尚未写出，依赖 Task 尾注的规则只能在 commit-msg hook 中运行；为此 CheckOptions 增加 MessageFile，LoadOptions 增加 Message。MOD-01、MOD-09、flows.md 因此回到 Draft，待确认。生成文件暂按 TERM-026 列为 harness.json、harness.lock、.claude/settings.json、CODEOWNERS、.github/workflows/harness.yml，T-019 实现 init 后以 scaffold 的清单为准。
- 自检发现：`harness check commits 552eb0b..HEAD` 报告 24 条违规。11 个"docs: T-00x 完成"提交没有 Task 尾注；T-003、T-006、T-007、T-008、T-010 有 13 处改动超出任务范围（共享测试夹具、T-008 的 engine 接口调整等），当时应先扩大任务范围再改。历史提交不改写，提交规则从 920a059 起生效：此后每个提交（含文档提交）都带 Task 尾注，并在提交前用 `harness check commit adr.03 --msg-file` 检查；T-012 实现过程中已按此先扩大了自身范围（workspace.go、registry_test.go）。
