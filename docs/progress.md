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
- 遗留：无。

### 2026-11-26 T-009
- 结论：Done。debt.01、debt.02、debt.03，注释按行拆分后检查，编号按数值归一（`DEBT-2` 与 `DEBT-002` 相同）。10 个用例（FR-031）集成测试通过；分别破坏三条规则的判定后 AC-031-1/E1、AC-031-2/E2、AC-031-3/E1 失败。
- 实现取舍：debt.03 只看注释中的引用，测试夹具字符串里出现的编号不算引用，否则这类测试会让技术债永远无法关闭。
- 用例说明：AC-031-2/B1 原文是 `# FIXME T-5`，Release 1 只内置 Go 语言包，集成测试用 `// FIXME T-5` 验证同一行为。
- 遗留：无。

### 2026-11-26 T-010
- 结论：Done。prd.11、arch.04、arch.05。T-010 关联的 AC-033-1、AC-033-3、AC-033-4 共 8 个用例集成测试通过；prd.11 同时实现 AC-033-5 的检查，3 个用例的测试也已写好（该 AC 由 T-020 关联）。逐一破坏各判定分支后对应的 E 用例失败。
- PRD 修改：NFR-006、NFR-007 缺少 CHG-010 声称已补的 Adds、Allowed-Paths、Closed-Modules 字段，按正文补齐并登记 CHG-015。PRD 内容在批准后又有变化，待用户重新确认。
- 实现取舍：arch.04 与 arch.05 不处理不存在的引用，由 trace.03 报告，避免同一问题报三次。
- 自检：arch.05 报告 EXT-04（新增阶段，NFR-008）没有示例扩展测试，属于 T-014 与 T-023 的范围。selfcheck.py 的扩展点检查已由 arch.04 取代。
- 遗留：EXT-04 的示例扩展测试由 T-014 补上（已完成）。

### 2026-11-26 T-011
- 结论：Done。task.01 ~ task.06。26 个用例（AC-011-1、2、3、4、6，AC-034-2）集成测试通过；逐一破坏各判定分支后对应用例失败。AC-011-3/E1 补了一个断言：路径属于存在但未关联的模块时也失败，原用例的路径不属于任何模块，单靠它测不出关联判断。
- 实现取舍：Removed 标记只用于需求和用例；AC-003-3 的正文以"Removed"开头，按用例规则判断会误报为已删除。规则样本改用最小夹具，完整夹具复制 12 份超出提交行数上限。
- 自检：`harness check task` 对本仓库 0 条违规，taskcheck.py 原型不再需要。
- 遗留：DEBT-002 仍为 Open，剩 PRD 中与用例无关的诊断和进度文件的诊断。

### 2026-11-26 T-012
- 结论：Done。commit.01 ~ commit.04；`harness check` 支持 `--commits <范围>`、`check commits <范围>` 与 `--msg-file`（commit-msg hook 传入即将形成的提交信息）。14 个用例（AC-013-1、2、4，AC-034-1）集成测试通过；逐一破坏各判定分支后对应用例失败。
- 设计澄清：pre-commit 时提交信息尚未写出，依赖 Task 尾注的规则只能在 commit-msg hook 中运行；为此 CheckOptions 增加 MessageFile，LoadOptions 增加 Message。MOD-01、MOD-09、flows.md 因此回到 Draft，待确认。生成文件暂按 TERM-026 列为 harness.json、harness.lock、.claude/settings.json、CODEOWNERS、.github/workflows/harness.yml，T-019 实现 init 后以 scaffold 的清单为准。
- 自检发现：`harness check commits 552eb0b..HEAD` 报告 24 条违规。11 个"docs: T-00x 完成"提交没有 Task 尾注；T-003、T-006、T-007、T-008、T-010 有 13 处改动超出任务范围（共享测试夹具、T-008 的 engine 接口调整等），当时应先扩大任务范围再改。历史提交不改写，提交规则从 920a059 起生效：此后每个提交（含文档提交）都带 Task 尾注，并在提交前用 `harness check commit adr.03 --msg-file` 检查；T-012 实现过程中已按此先扩大了自身范围（workspace.go、registry_test.go）。
- 遗留：无。

### 2026-11-26 T-013
- 结论：Done。`harness approve <产物>`、approve.01。7 个用例（AC-037-1、AC-037-3）集成测试通过，交互式终端用系统 `script` 命令提供的伪终端测试；逐一破坏终端判定、确认、检查、哈希比对后对应用例失败。
- 设计澄清：批准哈希不含文件头 Status 行，否则写入 Approved 本身会让批准失效；可批准产物与哈希计算放在 workspace（approval.go），任务范围相应扩大，MOD-09 增加对应接口。终端判定用 ioctl / GetConsoleMode，`/dev/null` 也是字符设备，只看文件类型会误判为终端。没有提供任何绕过交互的环境变量。
- 中断：实现过程中磁盘写满，用户清理后继续。
- 待用户操作：PRD、架构文档、详细设计、ADR 共 26 份产物需要用户在终端用 `harness approve` 补录批准；补录前 approve.01 报告违规，`harness lock` 被挡住。
- 遗留：产物批准由用户补录（已完成，见下一条记录）。

### 2026-11-26 T-013 批准补录
- 用户在终端用 `harness approve` 批准了 26 份产物（PRD、总览、格式规范、调用流程、14 个模块详细设计、8 个 ADR），批准记录写入 harness.lock，approve.01 为 0 条违规。此前回到 Draft 的 formats.md、MOD-01、MOD-02、MOD-04、MOD-05、MOD-09、flows.md 与 PRD 的 CHG-013 ~ CHG-015 一并获批。
- 之后任何产物内容变化都会使对应批准失效，需用户重新批准。

### 2026-11-26 T-014
- 结论：Done。phase 模块（内置四个阶段 JSON、order.json、项目配置合并、Current、Next、phase.01）、`harness phase list|next`、`TestExample_EXT_04`。AC-045-1、AC-045-2 共 8 个用例集成测试通过；AC-045-3 的 4 个 `phase next` 用例也已通过。逐一破坏各判定分支后对应用例失败。
- 规划调整：AC-045-3/E3（门禁拒绝代理修改当前阶段）依赖门禁，AC-045-3 改由 T-016 关联，T-014 只关联 AC-045-1、AC-045-2。AC-045-2/N2 的"项目规则"用内置规则 debt.01 代替，项目规则是 Release 2（FR-046）。
- 设计补充：formats.md 写明 `phases` 的两种写法（新增阶段用 after，修改已有阶段用 addExit 等、移除须引用 ADR）与规则引用写法；MOD-10 增加 Selector、Rule。两份文档因此回到 Draft，待用户重新批准。
- 范围：实现中把"非交互终端"错误改为 approve 与 phase 共用，改到了 approve.go；提交检查拦下后先扩大 T-014 范围再提交。
- 待用户操作：本仓库锁文件中的当前阶段为空，即 requirements。门禁（T-016）启用前需要用户在终端执行三次 `harness phase next` 推进到 implementation。
- 遗留：用户需在终端推进阶段到 implementation。

### 2026-11-26 T-015
- 结论：Done。task.07（InProgress 任务要求 PRD 与所引用设计有效批准、关联需求无 Open 待澄清项）、task.08（Done 任务的用例全部通过、进度记录含结论与遗留、全量检查通过）。19 个用例（FR-038、FR-039、AC-024-2，另含 AC-011-5/E1）集成测试通过；逐一破坏各判定分支后对应用例失败，AC-039-1/B2 的夹具因此改为同一 FR 下有两条 AC。
- 需求解释（用户确认）：AC-039-2 的"全量检查"不含 test.02 与 arch.05。这两条要求全部用例与扩展点，在 Release 1 完成前必然失败；完整覆盖由实现阶段的退出规则 `*` 保证。解释写在 task_08.go 的注释中；没有改 MOD-05，改了会让引用 IF-014 的 T-015 立即违反 task.07。
- 格式补充：OQ 可选 Relates 字段；进度记录须含 `- 结论：` 与 `- 遗留：` 两行，此前 6 条记录缺少遗留行，已补上。formats.md 仍为 Draft。
- 自检：task.07、task.08 对本仓库 0 条违规；全量检查只剩 test.02 对后续任务用例的覆盖违规。
- 遗留：formats.md、MOD-10 待用户重新批准；本仓库阶段仍为 requirements，待用户推进。

### 2026-11-26 T-016
- 结论：Done。gate 按 gate.protect → gate.artifact → gate.write 判定；Bash 命令用 mvdan.cc/sh v3.14.1 解析重定向、tee、cp、mv、rm、sed -i、git checkout/restore 等写入目标，以及 harness approve / phase next；无法静态判定的命令放行执行，命令前快照（大于 1 MiB 只记元数据），命令后恢复不允许的改动并反馈清单。claudecode 适配器解码 PreToolUse/PostToolUse/SessionStart/Stop，输入无法解析即拒绝；入口为 `harness gate <代理> <事件>`。18 个集成测试（AC-012-1、AC-012-2、AC-012-5、AC-037-2、AC-045-3/E3、AC-045-4，另含 FLOW-03）通过；逐一破坏各判定分支后对应用例失败，FLOW-03 的断言因此加上"已恢复"字样。
- 设计取舍：Edit 的写入后内容由适配器读取当前文件并应用替换得到，用来判定状态行是否被改为 Approved；读取失败时退回新增片段。仓库外路径不受本项目流程约束，直接放行。
- 遗留：gate.sensitive、gate.secret、gate.command 属于 T-017；hook 配置写入项目属于 init 任务；本仓库尚未启用 hook（阶段仍为 requirements）。

### 2026-11-26 T-017
- 结论：Done。gate.sensitive（读写与 Bash 参数、输入重定向命中 sensitivePaths，sensitiveAllow 豁免）、gate.secret（规则 secret.aws、secret.pem、secret.gcp、secret.github、secret.entropy，只检查新增行与命令文本）、gate.command（dangerousCommands 模式按词顺序匹配，--force-with-lease 不误判；rm 递归删除仓库外、~ 开头或含变量的路径）。17 个集成测试（FR-025）通过；逐一破坏各判定分支后对应用例失败。十六进制排除分支被证明多余（提交哈希熵不超过 4 比特且通常全小写），已删除。
- 设计取舍：整条命令只调用一次 harness（如 `harness lock`）时，命令后比对允许它改写锁文件、harness.json 与 hook 配置；夹带其他命令或 harness 只是参数时仍恢复。app/gate.go 因此纳入本任务范围。
- 遗留：危险命令按文本模式匹配，经变量或脚本间接执行的命令拦不住，靠命令后比对兜底（仅限仓库内文件）。

### 2026-11-26 T-018
- 结论：Done。写入产物后按文件选择规则前缀（PRD → prd、term；架构 → arch、trace、term；ADR → adr；任务 → task；技术债 → debt）运行文档规则并反馈；会话结束运行全量检查（engine.FullCheck，不含覆盖类规则），连续阻止 stopMaxBlocks 次后放行，未通过项写入 docs/progress.md 的"## 未完成"段，检查通过时清除；会话开始注入未完成项（最先）、阶段、进行中任务、未通过检查数、Open 待澄清项与冒烟结果（失败时最后 50 行）。12 个集成测试（AC-012-3、AC-012-4、AC-024-1、AC-024-3）通过；逐一破坏各判定分支（含两处计数清零）后对应用例失败。
- 设计取舍：会话结束检查与会话开始注入沿用 task.08 的解释（全量检查不含 test.02、arch.05），否则 Release 1 完成前每次会话都会被阻止 3 次。写后检查只运行不依赖测试报告与版本库的规则。
- 遗留：DEBT-003（task.08 仍重复维护排除列表）；Bash 命令改动产物后不运行写后检查，由会话结束检查兜底。
