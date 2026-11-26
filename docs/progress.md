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

### 2026-11-26 T-019
- 结论：Done。`harness init --lang <语言列表> [--agent <代理列表>]` 写入内嵌模板：PRD、架构总览、任务、技术债、进度骨架，每种语言的配置包（`docs/lang/<语言>.md` 测试与注释约定，pack.json 的 sourcePaths 合并进 harness.json），以及 claudecode 的 `.claude/settings.json`。已存在且内容相同的文件不动，内容不同的跳过并列出；语言或代理不支持时不写任何文件。10 个集成测试（FR-001、AC-N002-2、NFR-005）通过；逐一破坏各判定分支后对应用例失败（目录创建与文件写入两处错误分支互为兜底，同时破坏才失败）。
- 设计澄清（用户确认）：AC-001-1/B1 要求 5 种语言，Release 1 新增 TypeScript/JavaScript、Python、Java/Kotlin、C/C++ 的最小语言包，只提供注释提取、测试分类、测试名，T-019 的范围因此纳入 internal/lang/**。完整的格式化与静态检查配置属于 FR-014（Release 2）。
- 设计取舍：prd.04（AC-002-3）要求三章各有条目，PRD 骨架带 TERM-001、FR-001、NFR-001 三个格式示例，Release 为 2，不进入当前 Release 的追溯与覆盖检查，澄清需求时替换。MOD-13 列出的 formats 参考与 ADR 模板没有放进骨架：放在 docs/architecture、docs/adr 下会被当作产物解析；改由 T-020 的 skill 说明格式。
- 遗留：T-021 的 git hook 文件（commit-msg、pre-commit）没有扩展名，与 AC-N005-1/N1 的"全部产物扩展名为 .md、.yaml 或 .json"冲突，届时需确认 git hook 是否算产物。

### 2026-11-26 T-020
- 结论：Done。六个 skill（clarify、change、architect、plan、implement、test）作为模板内嵌，`harness init` 写入 `.claude/skills/<名称>/SKILL.md`；frontmatter 的 check 字段声明结束前运行的 `harness check` 子命令。静态测试：每个 skill 的 check 选择器都选中已注册规则、内置阶段引用的 skill 都存在；测试 skill 中五种语言的命名示例逐一放进临时项目，用 test.01、test.02 实际提取验证计入覆盖，并以不按约定命名的测试作对照。5 个用例（AC-016-1、AC-016-2、AC-011-5/N1）通过；逐一破坏 skill 内容后对应测试失败。
- 设计取舍：skill 写入需要改 scaffold.go 与 init 的期望文件清单，T-020 的 Design 因此加上 IF-026、FLOW-07。FR-016 描述中的 UI 规范 skill 属于 FR-015 的范围，不在本任务。
- 遗留：无。

### 2026-11-26 T-021
- 结论：Done。git 仓库中 init 安装可执行的 commit-msg（`harness check commit --msg-file`）与 pre-commit（产物文档规则）hook，非 git 仓库跳过并提示；`--ci github|gitlab` 生成 `.github/workflows/harness.yml` 或 `.gitlab-ci.yml`，运行 `harness check` 与 `harness check commits <范围>`；scaffold 注册 cfg.02，比对 hook 配置与已存在的 CI 配置，不一致时指出第一处不同的行并给出重新生成命令。10 个集成测试（AC-013-3、AC-043-1、AC-041-2）通过，其中 commit-msg hook 被实际执行并拒绝没有 Task 尾注的提交；逐一破坏各判定分支后对应用例失败。变异检查发现一处破坏导致构建失败被误判为存活，已改用能编译的变异重做。
- 设计取舍：cfg.02 只在存在 harness.json 的项目（执行过 init）上运行；本仓库在 T-026 启用前没有 harness.json，提前写入 `.claude/settings.json` 会让门禁在当前会话中生效。git hook 不在版本库中，cfg.02 不比对，CI 环境里才不会误报。T-018 的会话夹具手写了 harness.json，因此补上生成的 hook 配置。
- 需求解释：AC-N005-1/N1 的".yaml"按"YAML 文本"理解，`.yml` 同样允许；AC-043-1/N1 指定 `harness.yml`，GitLab 也要求 `.gitlab-ci.yml`。git hook 位于 `.git/hooks`，不属于项目产物。
- 遗留：CI 模板的安装命令依赖 DEBT-001 的模块路径；`.gitlab-ci.yml` 不在 commit 规则的生成文件豁免列表中（builtin 的 generatedFiles），属于 MOD-05，留给后续修改该模块的任务。

### 2026-11-26 T-022
- 结论：Done。cfg.01：harness.json 的规范化哈希（JSON 解析、去掉 `"//"` 注释项、键排序）相对锁文件变化时，须有锁定之后新增的 CHG（对象写 harness.json）或新增的 ADR（Relates 含 harness.json）；锁文件因此记录 ADR 编号，prd.08 跳过对象为 harness.json 的 CHG。ci.01：CI 配置（GitHub Actions、GitLab CI、CircleCI、Azure、Bitbucket、Jenkinsfile）至少一份含 `harness check`。7 个集成测试（AC-041-1、AC-043-2，另含 ADR 说明配置变化）通过；逐一破坏各判定分支后对应用例失败，AC-043-2/E1 的夹具因此改为提到 harness 但不运行 harness check。
- 需求解释：AC-041-1/B1 的"只改了注释"按 JSON 中的 `"//"` 注释项理解，空白与键顺序同样不算变化，已写入 formats.md 第 6 节。ci.01 与 cfg.02 一样只检查有 harness.json 的项目，本仓库在 T-026 之前不受影响。
- 遗留：本仓库的 harness.lock 还没有 ADR 条目，下次 `harness lock` 时补上；formats.md 仍为 Draft，待用户重新批准。

### 2026-11-26 T-023
- 结论：Done。临时 git 仓库直接使用本仓库的 docs/architecture/overview.md，EXT-03、EXT-04 的 Paths 与 Closed 改动后测试随之变化。示例阶段 release（阶段定义、order.json、skill 模板）与示例规则 demo.01（规则文件、样本、registry.go）的扩展提交通过 commit.04；同一提交再改 internal/gate 或 internal/rules/engine 时失败并列出文件。4 个用例（NFR-008、NFR-009）通过。
- 遗留：无。

### 2026-11-26 T-024
- 结论：Done。静态测试对模块内每个包（除 internal/audit）用 `go list -deps` 检查不依赖 net、net/http、crypto/tls 等网络包，并以 testdata 中引入 net/http 的包作对照；临时给 lock 包加 net/http 导入后测试失败。断网用代理变量指向关闭的端口模拟，check（文本与 JSON）、trace、phase list 的输出与联网时及重复运行逐字节一致，两次 lock 生成的锁文件一致。2 个用例（NFR-001）通过。
- 遗留：断网是用代理变量模拟的，没有真正隔离网络；不引入网络包的静态测试保证了这一点。

### 2026-11-26 T-025
- 结论：Done。样例仓库 5000 个 Go 源文件、200 个任务：PreToolUse 门禁 40 次（含进程启动）p95 约 15ms；全量检查约 0.4s，并用最后一个源文件里不合规的 TODO 确认全部源文件被扫描；50020 行 PRD 的 `harness check prd` 约 55ms。AC-N003-1/E1 用构造的耗时序列验证 p95 计算与超阈值判失败。4 个用例（NFR-003）通过（本机 Apple Silicon，数字随机器变化，阈值留有一个数量级以上余量）。
- 遗留：无。

### 2026-11-26 T-026
- 结论：Done。用户在终端安装 harness（`go install ./cmd/harness`）、重新批准 PRD、formats.md、MOD-10，并把阶段推进到 implementation；随后对本仓库执行 `harness init --lang go`，写入 harness.json、`.claude/settings.json`、六个 skill、GitHub Actions 配置、`docs/lang/go.md` 与 git hook（已有产物全部跳过），CHG-016 说明了 harness.json 的新增，`harness lock` 通过。之后的两次提交都经过本地 commit-msg 与 pre-commit hook。5 个测试（FR-017，另含生成文件豁免）通过；AC-017-1/N1 对仓库根目录运行除依赖测试报告与暂存区之外的全部规则，因为测试运行时报告文件正在被重写。
- 设计澄清（用户确认，第 15 轮）：hook 调用 PATH 上的 `harness`；`harness init` 写入的 skill 纳入生成文件（TERM-026），`.gitlab-ci.yml` 一并加入豁免。
- 更正：T-025 标为 Done 时性能测试的一条注释含未登记的 TODO 字样，task.08 随即报出，已在 4ce2ff8 修正。
- 遗留：门禁 hook 在 Claude Code 重新加载配置后生效；CI 配置的安装步骤依赖 DEBT-001 的模块路径，T-027 处理。

### 2026-11-26 T-028
- 结论：Done。T-027 期间门禁在 `d=$(mktemp -d)` 这类只含赋值的命令上 panic：CallExpr 的 Args 为空，`args[1:]` 越界，进程以退出码 2 结束，Claude Code 把它当作拒绝。改为跳过无参数的调用，赋值里命令替换的写入照常判定。secret.entropy 不再把以 `/`、`./`、`~/` 开头的路径当作密钥（DEBT-006），含多个 `/` 的 AWS 示例 secret 仍命中。两个回归测试通过，各自的变异都被抓到。
- 更正：提交 388f331 时测试报告里有一条失败（`TestParseRepoArtifacts_IF_001` 断言任务数为 27，新增 T-028 后变为 28），提交前没有核对失败数；由 T-029 修正。
- 遗留：没有规则要求非用例测试（如 IF 编号的单元测试）通过，`harness check` 因此没有报告这条失败，见 DEBT-008。

### 2026-11-26 T-029
- 结论：Done。`TestParseRepoArtifacts_IF_001` 改为下限断言（任务数至少 29），只保证本仓库产物无诊断地解析且没有漏读整类实体；全量测试 0 失败。
- 遗留：无。

### 2026-11-26 T-027
- 结论：Done。用户确认后建立仓库 donkeydluffy/harness（先私有，后改公开）与公开的 donkeydluffy/homebrew-tap，模块路径改为 `github.com/donkeydluffy/harness`（DEBT-001，用户在终端运行脚本并以 `--no-verify` 提交，fd3898e）。`.goreleaser.yaml` 构建 6 个平台的归档、checksums.txt 与 Homebrew formula；release.yml 先发草稿，校验 6 个归档、校验和与 formula 版本后公开，测试直接运行工作流中的校验命令（3 个变异均被抓到）。用户设置 HOMEBREW_TAP_TOKEN 并推送 v0.1.0 后，发布运行 45785485863 通过：Release 公开，含 6 个归档与 checksums.txt，tap 中 Formula/harness.rb 的版本为 0.1.0；本地下载后 `shasum -c` 全部 OK，darwin/arm64 二进制可运行。6 个用例（AC-N002-1、AC-N002-3）通过。
- 期间修正：CI 模板运行测试、生成报告（Go 语言包声明 ciTest，cfg.02 按 harness.json 的语言生成期望内容）；harness 自身的仓库在 CI 中从源码构建（代理缓存的 @latest 落后于最新提交）。门禁的两个缺陷由 T-028、T-029 修复。
- 遗留：goreleaser 的 `brews` 已标为弃用，工作流固定 v2.20.3 继续使用；没有在本机执行 `brew install` 验证 formula（会改动本机环境）；DEBT-004、005、007、008、009、010 待后续 Release 处理。

### 2026-11-26 T-030
- 结论：Done。生成的 hook 命令改为 `command -v harness >/dev/null 2>&1 || {…}; exec harness gate claudecode <事件>`：找不到 harness 时 PreToolUse 输出拒绝（AC-012-5/E1 的失败即阻断），SessionStart 注入"harness 未安装"，PostToolUse 与 Stop 不阻止。`TestGate_MissingBinaryFailsClosed` 用 /bin/sh 按 Claude Code 的方式执行生成的命令，分别在 PATH 不含与含 harness 时验证；两个变异都被抓到（其中一个起初因编译失败看似存活，换成可编译的变异后确认）。用户在终端重新生成了本仓库的 `.claude/settings.json`。
- 遗留：新的 hook 配置在 Claude Code 重新加载后生效。

### 2026-11-26 T-031
- 结论：Done。workspace 新增 `RegeneratedFiles`（GitHub、GitLab 的 CI 配置）。整条命令只是 `rm`（可带 `-f`）这些文件时，门禁命令前放行、命令后保留删除；单独一次 `harness init` 写回它们时保留。复合命令、`-r`、混入其他路径都拒绝，通配符照常命令前放行、命令后恢复。最初把 `.claude/settings.json` 也列入清单，对照 AC-025-4（代理修改 hook 配置须拒绝）后移除：代理删除 hook 配置而不写回会让下次会话失去门禁。回归测试通过，三个变异都被抓到。
- 遗留：hook 配置的模板变化后仍需由人运行 `rm .claude/settings.json && harness init`。

### 2026-11-26 T-032
- 结论：Done。新增 test.03：测试报告中每个失败的测试报一条（含测试名与包），计入全量检查；在本仓库的报告里追加一条失败的单元测试后，test.03 与 task.08 都报告。新增 doc.01：报告其他结构规则不负责的解析诊断；`util.go` 的 `diagOwner` 统一判定每条诊断的归属，adr.01、arch.01、debt.01、prd.07、task.01 改为按它取诊断，原先架构文件中含"用例"的诊断会被 prd.07 与 arch.01 各报一次，现在只报一次。task.08 改用 `engine.FullCheck`。测试确认每条诊断恰好被一条规则报告且归属正确；三个变异中一个（架构诊断改归 doc.01）起初存活，补上归属断言后被抓到。
- 遗留：测试包构建失败仍不被报告（DEBT-011，在 MOD-08）；门禁不跟随命令中的 cd（DEBT-012）。

### 2026-11-26 T-033
- 结论：Done。原计划在项目根目录的 `.gitignore` 追加条目，但 `.gitignore` 不属于任何模块，代理运行 init 写它会被门禁恢复、提交会被 commit.02 拒绝；改为 init 写入 `.harness/.gitignore`（`snap/`、`session/`、`.gitignore`），门禁本就放行 `.harness/` 下的写入，文件忽略自身因而不需要提交。测试确认快照、会话目录与该文件都不出现在 `git status`，测试报告照常出现，根目录 `.gitignore` 不被创建；去掉 `session/` 的变异被抓到。AC-N005-1 约束的是产物，`.harness/` 是运行时目录，测试跳过它（与"git hook 不是产物"的判断一致）。本仓库已重新运行 init，`git status` 不再显示 `.harness/`。
- 遗留：无。

### 2026-11-26 T-034
- 结论：Done。`ParseGoJSON` 记录包级结果为 fail 的包；包内没有失败的测试时（构建失败、TestMain 失败、测试外的 panic、超时）追加一条名为 `PackageFailure` 的失败记录。包因其中的测试失败而失败时不重复记录。单元测试覆盖构建失败、测试全过但包失败、测试失败、通过四种情况，两个变异都被抓到；在本仓库报告中追加一条构建失败记录后，test.03 报告了它。
- 遗留：写测试时 secret.entropy 把较长的测试名判为密钥，登记 DEBT-013。

### 2026-11-26 T-035
- 结论：Done。门禁解析命令时维护当前目录：顺序执行的顶层字面量 `cd` 被跟随，多次 `cd` 累积，写入、敏感路径与递归删除都按当前目录解析。`cd` 目标无法静态判定（子 shell、`||`、`cd -`、无参数、含变量）时，之后相对路径的写入命令前放行、由命令后比对恢复，递归删除拒绝；绝对路径照常判定。集成测试覆盖范围内外写入、`..` 跳出、仓库外递归删除、无法判定目录的五种写法及命令后恢复，五个变异都被抓到。
- 遗留：无。

### 2026-11-26 T-036
- 结论：Done。计划中的条件（80% 的字母落在 3 个字符以上的段）用一百万个随机 base62 token 验证时放过了 0.8%，安全上不可接受；逐步收紧为：短段至多 2 个、数字段至多 3 个、像单词的段至少 3 个、非全大写段的元音占比不低于 30%。放过率降到约百万分之 14，7 个样例测试名全部不误报；元音阈值取 0.33 时含 HTTP、KB 的基准测试名会误报，因此取 0.30，并把全大写缩写排除在元音比例之外。单元测试覆盖长测试名不误报、GitHub/Stripe 形式的随机 token 与 AWS 示例 secret 仍命中、随机 token 放过率上限；集成测试确认写入与命令都放行。5 个变异被抓到，单独去掉"数字段至多 3 个"的变异没有测。
- 遗留：含多个短段的长测试名（如带 `AC_012_2_N1` 这类后缀又有 `Cd` 之类短词）不满足标识符条件，若熵又偏高仍会误报，目前样例中这类名字熵都低于阈值。

### 2026-11-26 Release 2 需求（第 16、17 轮）
- 结论：PRD 新增 TERM-027、FR-047（按条目批准）、FR-048（取消流程阶段）、AC-046-7（批准前检查），均为 Release 2；FR-047、FR-048 用 `Replaces` 声明取代 FR-038、FR-045、NFR-008、AC-037-3、AC-024-1，旧需求标注 `Replaced-By`。Release 1 范围内新增 AC-013-5（需求提交用 Change 尾注，T-037）与 AC-003-6（取代关系按 Release 生效，T-038）。PRD 改回 Draft 等待批准。T-029 的 Covers 由 FR-003 收窄为 AC-003-1：它只修改了编号解析的测试，关联整条 FR-003 使新增的 AC-003-6 计入它的完成定义。
- 遗留：Current-Release 仍为 1，等 Release 2 其余需求（DEBT-007、DEBT-010、FR-014、FR-015）澄清后再改。

### 2026-11-26 T-037
- 结论：Done。commit.01 接受只改 `docs/` 与 `harness.lock` 的提交用一个 `Change: CHG-n` 尾注代替 Task 尾注；CHG 须在变更日志中存在，不能与 Task 尾注同时出现，不能有多个。5 个用例的集成测试与 5 个变异都通过验证。流程上的疏漏：拆任务前没有先更新设计，代码提交后才补上 formats.md、flows.md、MOD-05 的规则目录，由用户重新批准；trace.01 只按 FR 检查设计覆盖，新增 AC 不会触发它，FR-047 的批准链实现后才能拦住这类情况。会话结束检查在用户批准设计前连续阻止 3 次后放行，写入的"未完成"段落在批准后已无失败项，已删除。
- 遗留：DEBT-014（较长的相对路径被判为密钥）、DEBT-015（shell 停在子目录时门禁把子目录当作仓库根）。

### 2026-11-26 T-039
- 结论：Done。先更新设计（IF-017 增加 Root 与 Cwd，MOD-12 说明由适配器查找项目根），用户批准后实现：适配器填写 Root 与 Cwd，`harness gate` 以 Root 为根，Root 为空时退回进程当前目录；命令解析的起始目录取 Cwd，再跟随命令中的 `cd`。集成测试在子目录中触发 hook，覆盖经 `../../harness.lock` 写锁文件被拒绝、相对路径的写入与 Write 工具按子目录判定范围、快照建在项目根而不是子目录、命令后比对删除范围外新建的文件并恢复锁文件；三处改动分别回退都使测试失败。本会话让 shell 停在 `docs/architecture` 后执行命令，子目录中不再出现 `.harness/`。
- 遗留：无。

### 2026-11-26 T-038
- 结论：Done。新增 `replacedBy`：取代者为 Active 且 Release 不大于 Current-Release 时，被取代的需求或验收标准按 Removed 判定。trace.01、task.02、test.02 不再要求它被覆盖，trace.03 报告"引用了已被 X 取代的 Y"；新增 prd.12 校验 Replaces 引用存在、Replaced-By 互相对应、被取代者的 Release 小于取代者。开工时发现任务范围漏了 trace_01.go、task_02.go、test_02.go（同属 MOD-05），补进范围后实现。6 个用例的集成测试覆盖取代前后的设计、测试、任务覆盖与三种一致性错误；12 个变异中最初有 4 个存活（B1 中被取代的需求仍有设计覆盖，任务覆盖与整条需求被取代的路径没有测到），补测试后全部被抓到。在临时副本中把 Current-Release 改为 2 预演：trace.03 列出 23 处引用待清理（FR-045 17 处、FR-038 4 处、NFR-008 2 处），trace.01 列出 15 处 Release 2 需求尚无设计覆盖。
- 遗留：DEBT-014 补充一种赋值形式的误报；登记 DEBT-016（cp 目标含变量时把来源当作目标）。

### 2026-11-26 T-040
- 结论：新增 MOD-18（internal/version）；`harness version` 与 `--version` 输出版本，发布构建经 .goreleaser.yaml 的 ldflags 注入版本号，源码构建取模块版本（Go 1.24 起为含提交哈希的伪版本，`+dirty` 改为 `-dirty`），没有模块版本时取提交哈希前 12 位；用法与 help 列出 version。IF-020、IF-021、IF-037 补锚点与单元测试。变异测试：去掉 -dirty 转换，单元测试与 AC-049-1/B1 都失败。
- 遗留：无。

### 2026-11-26 T-041
- 结论：新增 MOD-15（internal/approval）：批准单位（需求、术语、待澄清项、设计元素、ADR、任务、各文件正文）及其哈希（不含 Status，PRD 正文不含变更日志）、批准链与状态、待批准单位、批准记录。锁文件第 2 版的 `units` 记录批准内容与上游哈希；锁文件为第 2 版时 `harness approve` 按单位批准，列出新增、修改、失效的单位与逐行差异，只运行这些单位自身的检查（不要求下游覆盖）并要求上游已批准；approve.01 以告警列出未批准与失效单位及原因，task.07 按批准链判定。engine 的告警放入 Report.Warnings，不计入违规与退出码（与 IF-012 设计的 Report 结构有出入，T-048 对齐设计时一并处理）。锁文件为第 1 版时行为不变；本仓库的迁移在 T-042。变异测试：approval 5 处、approve 流程 5 处均被测试发现。
- 遗留：task.07 的批准链分支由 T-043 的 AC-047-4 用例覆盖。

### 2026-11-26 T-042
- 结论：`harness lock` 把有按文件批准记录的第 1 版锁文件迁移为第 2 版：仍有效的文件批准（只看每个产物最近一次）换算为文件内全部单位的批准，沿用原批准人与时间；已失效文件中的单位不记录；有效与失效的产物写入审计记录 `unit-migration`。没有批准记录的第 1 版锁文件（含新建的）不迁移，保证两次 lock 结果一致（AC-N001-1）。`phase` 字段保留到 T-054，因为门禁在 T-043 前仍按阶段判定产物写入。变异测试 4 处：3 处被发现，`!l.ByUnit()` 为等价变异（迁移后 Legacy 为空）。
- 遗留：本仓库的锁文件在本任务完成后迁移；之后任务开工前需由人按单位批准任务。
