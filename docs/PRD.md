# PRD: harness

- Status: Approved
- Version: 1.0-rc3
- Current-Release: 1
- Updated: 2026-11-26

> 本文件是 harness 的唯一需求来源，同时也是 harness 自举的第一个受检产物。格式即规范：harness 的 `check prd` 会按本文件的结构解析。

## 1. 背景与目标

AI 编程代理（Claude Code 等）在没有约束时会"想到什么做什么"：需求没澄清就写代码，设计和需求对不上，改动范围随意扩散，UI 随手写颜色。现有的 Spec-Driven 框架（spec-kit、Kiro、OpenSpec、BMAD）把需求→设计→任务的对应关系交给大模型判断，没有确定性校验，也没有模块依赖和接口变更约束（见 `docs/research/2026-11-25-landscape.md`）。

harness 的目标：用一套 skill 引导流程，用一个确定性的检查工具和 hook 强制执行流程，让"需求 → 架构 → 任务 → 代码 → UI"每一步都有可机器校验的对应关系。

## 2. 统一语言（Glossary）

代码、文档、命令输出中只使用"术语"列的名称；"禁用别名"列出现在受检文档中视为违规。

| ID | 术语 | 代码名 | 定义 | 禁用别名 |
|---|---|---|---|---|
| TERM-001 | 产物 | Artifact | harness 管理的结构化文档：PRD、架构声明、详细设计、ADR、任务列表、设计 token | 交付物 |
| TERM-002 | 需求 | Requirement | PRD 中带编号的条目，分功能需求（FR）和非功能需求（NFR） | |
| TERM-003 | 验收标准 | AcceptanceCriterion | 需求下可判定通过与否的条目，编号 `AC-<需求号>-<序号>` | |
| TERM-004 | 待澄清项 | OpenQuestion | 尚未确认的需求问题，编号 OQ-n；存在未关闭的待澄清项时 PRD 不能进入 Approved | |
| TERM-005 | 变更记录 | ChangeEntry | PRD 变更日志中的一条，编号 CHG-n，声明新增/修改/删除了哪些需求或接口及原因 | |
| TERM-006 | 模块 | Module | 架构中的代码单元，编号 MOD-n，声明路径范围、允许依赖的模块、对外接口、覆盖的需求 | 组件 |
| TERM-007 | 接口 | Interface | 模块对外暴露的契约，编号 IF-n，带签名定义 | |
| TERM-008 | 调用流程 | Flow | 跨模块的调用序列，编号 FLOW-n | |
| TERM-009 | 架构决策 | Decision | 一条 ADR，编号 ADR-n | |
| TERM-010 | 任务 | Task | 任务列表中的一项，编号 T-n，必须关联需求和设计元素并声明文件范围 | 工单 |
| TERM-011 | 设计元素 | DesignElement | 模块、接口、调用流程、架构决策的统称 | |
| TERM-012 | 追溯链接 | TraceLink | 两个带编号条目之间的显式引用 | |
| TERM-013 | 锁文件 | Lockfile | `harness.lock`，记录每条需求和每个接口的内容哈希，用于检测未登记的变更 | |
| TERM-014 | 门禁 | Gate | 在代理操作前后或会话结束时执行的确定性检查，不通过即阻断 | |
| TERM-015 | 语言配置包 | Profile | 某种编程语言的格式化、静态检查、编码规范、依赖检查配置集合 | |
| TERM-016 | 设计 token | DesignToken | UI 设计规范中可被引用的颜色、字号、间距、圆角、阴影等命名值，遵循 W3C DTCG 格式 | |
| TERM-017 | 阶段 | Phase | 流程中的一个步骤，声明进入规则、退出规则、可写产物路径；内置需求、架构、规划、实现，项目可新增 | |
| TERM-018 | 验收用例 | AcceptanceCase | 验收标准下的一个具体测试情形，分正常（N）、边界（B）、错误（E） | 测试点 |
| TERM-019 | 测试报告 | TestReport | 测试运行生成的 JUnit XML 或 `go test -json` 输出，是"测试已通过"的唯一证据 | |
| TERM-020 | 核心模块 | CoreModule | 架构声明中标记为核心、必须设置覆盖率阈值的模块 | |
| TERM-021 | 技术债 | DebtItem | 已知但暂不处理的问题，编号 DEBT-n，登记在技术债清单中 | |
| TERM-022 | 扩展点 | ExtensionPoint | 架构中为某个扩展场景预留的位置，编号 EXT-n，声明承载接口、扩展路径、封闭模块 | 插槽 |
| TERM-023 | 扩展场景 | ExtensionScenario | 以非功能需求形式写出的"未来会新增什么、允许改动哪里"的变更场景 | |
| TERM-024 | 封闭模块 | ClosedModule | 实现某扩展场景时不得修改的模块 | |
| TERM-025 | 规则 | Rule | 一条确定性检查，有唯一编号、违规样本和合规样本；分内置规则和项目自定义规则 | 校验项 |
| TERM-026 | 生成文件 | GeneratedFile | 由 `harness init` 或其他 harness 命令写入项目的配置文件：项目配置、hook 配置、CI 配置、git hook、CODEOWNERS | |

## 3. 功能需求

### FR-001 项目初始化
- Status: Active
- Release: 1
- Priority: P0

在目标项目中一条命令生成 harness 所需的全部骨架：产物模板、hook 配置、skill、rule、所选语言的语言配置包。

- AC-001-1: `harness init --lang <语言列表>` 在空目录执行后，`harness check` 全部通过（骨架本身合法）。
  - AC-001-1/N1: 空目录执行 `harness init --lang go`，随后 `harness check` 退出码 0
  - AC-001-1/B1: `--lang` 同时指定 5 种语言，生成 5 套语言配置包，`harness check` 退出码 0
  - AC-001-1/E1: `--lang` 含不支持的语言名，退出码非 0，输出支持的语言列表，目录中不留下任何文件
- AC-001-2: 在已有 harness 骨架的目录重复执行时不覆盖用户已修改的文件，并列出跳过的文件。
  - AC-001-2/N1: 已 init 的目录中修改了 PRD.md，再次 init，PRD.md 内容不变，输出中列出 PRD.md 为跳过
  - AC-001-2/B1: 已 init 的目录中删除了一个模板文件，再次 init，只补回被删除的文件
  - AC-001-2/E1: 目标目录不可写时，退出码非 0，报告不可写的路径

### FR-002 需求澄清
- Status: Active
- Release: 1
- Priority: P0

提供需求澄清 skill，与开发或产品人员多轮对话直到需求清晰，结果写入 PRD。

- AC-002-1: 每轮最多提出 5 个问题，问题与答案追加到 PRD 的"澄清记录"章节，每轮一个标题（日期 + 轮次）。
  - AC-002-1/N1: 澄清记录中某一轮下有 5 条问答，`harness check prd` 通过
  - AC-002-1/B1: 某一轮下有 6 条问答，失败并指出该轮标题
  - AC-002-1/E1: 澄清记录中的问答不在任何轮次标题下，或标题缺少日期，失败
- AC-002-2: 无法当场确认的问题登记为待澄清项（OQ-n）。存在 Open 状态的待澄清项时，`harness check prd` 拒绝 PRD 状态为 Approved。
  - AC-002-2/N1: 全部待澄清项为 Closed 且带 Resolution，PRD 为 Approved，检查通过
  - AC-002-2/B1: 最后一个 Open 的待澄清项改为 Closed 后，同一 PRD 的 Approved 状态由失败变为通过
  - AC-002-2/E1: 存在 1 个 Open 的待澄清项且 PRD 为 Approved，失败并列出该 OQ
  - AC-002-2/E2: 待澄清项为 Closed 但没有 Resolution 字段，失败
- AC-002-3: PRD 必须包含功能需求、非功能需求、统一语言三个章节，缺一则检查失败。
  - AC-002-3/N1: 三个章节齐全，检查通过
  - AC-002-3/E1: 缺少统一语言章节，失败并指出缺失的章节名
  - AC-002-3/B1: 非功能需求章节存在但没有任何 NFR 条目，失败

### FR-003 PRD 结构校验
- Status: Active
- Release: 1
- Priority: P0

`harness check prd` 确定性校验 PRD 结构。

- AC-003-1: FR/NFR/TERM/OQ/CHG 编号唯一、格式合法。
  - AC-003-1/N1: 全部编号唯一且格式合法，通过
  - AC-003-1/E1: 两条需求都使用 FR-003，失败并给出两处行号
  - AC-003-1/E2: 编号写成 FR-3 或 FR003，失败并给出合法格式
  - AC-003-1/B1: 编号到达 FR-999 后新增 FR-1000，格式仍判为合法
- AC-003-2: 每条需求有 Status（Active / Removed）和 Priority（P0/P1/P2）字段，Active 需求至少有一条验收标准。
  - AC-003-2/N1: 需求有 Status、Priority 和 1 条验收标准，通过
  - AC-003-2/E1: 缺少 Priority 字段，失败
  - AC-003-2/E2: Status 取值为 Done，失败并列出合法取值
  - AC-003-2/B1: Removed 需求没有验收标准，通过
  - AC-003-2/E3: Active 需求没有验收标准，失败
- AC-003-3: Removed 的需求编号保留在文档中，不允许被新需求复用。
  - AC-003-3/N1: FR-005 标记 Removed 后仍保留在文档中，通过
  - AC-003-3/E1: 锁文件中存在的 FR-005 从文档中消失，失败
  - AC-003-3/E2: FR-005 已 Removed，又新增一条编号 FR-005 的 Active 需求，失败
- AC-003-4: PRD 可拆分：`docs/PRD.md` 为索引，包含背景、统一语言、非功能需求、待澄清项、澄清记录、变更日志；功能需求可分散在 `docs/prd/<领域>.md`。所有文件中的编号全局唯一。
  - AC-003-4/N1: FR 分布在 `docs/prd/a.md` 和 `docs/prd/b.md`，编号不重复，通过
  - AC-003-4/B1: 没有 `docs/prd/` 目录，全部 FR 在 PRD.md 中，通过
  - AC-003-4/E1: `docs/prd/a.md` 与 PRD.md 中都定义 FR-010，失败并给出两个文件位置
  - AC-003-4/E2: `docs/prd/a.md` 中定义了统一语言章节，失败（统一语言只能在索引中）
- AC-003-5: 每条需求有 Release 字段（正整数）；PRD 头部声明 Current-Release。追溯、任务、测试类的覆盖检查只针对 Release 不大于 Current-Release 的需求。
  - AC-003-5/N1: Current-Release 为 1，Release 2 的 FR 没有设计覆盖，追溯检查通过
  - AC-003-5/B1: Current-Release 由 1 改为 2 后，同一份产物中 Release 2 的未覆盖 FR 开始报告失败
  - AC-003-5/E1: 需求缺少 Release 字段，失败
  - AC-003-5/E2: PRD 头部缺少 Current-Release，失败

### FR-004 需求变更追踪
- Status: Active
- Release: 1
- Priority: P0

需求的新增、修改、删除都必须在 PRD 变更日志中登记，由锁文件做确定性比对。

- AC-004-1: 某条需求的内容哈希与锁文件不一致，且变更日志中没有引用该需求的新 CHG 条目时，检查失败。
  - AC-004-1/N1: 修改 FR-002 描述并新增引用 FR-002 的 CHG 条目，执行 `harness lock` 前检查通过
  - AC-004-1/E1: 修改 FR-002 描述但没有新 CHG 条目，失败并指出 FR-002
  - AC-004-1/B1: 只修改需求标题后的空白字符，哈希不变（规范化后计算），检查通过
  - AC-004-1/E2: 新 CHG 条目引用的是 FR-003 而改动的是 FR-002，失败
- AC-004-2: CHG 条目声明的动作（新增/修改/删除）必须与实际差异一致，例如声明"修改 FR-3"但 FR-3 实际被删除时失败。
  - AC-004-2/N1: CHG 声明新增 FR-020，FR-020 是锁文件中不存在的新需求，通过
  - AC-004-2/E1: CHG 声明修改 FR-003，FR-003 实际被改为 Removed，失败并提示应为删除
  - AC-004-2/E2: CHG 声明删除 FR-004，FR-004 内容未变，失败
  - AC-004-2/B1: 一个 CHG 编号下多行分别声明新增和修改不同需求，各自与差异一致，通过
- AC-004-3: `harness lock` 在检查通过后更新锁文件；锁文件纳入版本控制。
  - AC-004-3/N1: 检查通过后执行 `harness lock`，锁文件更新为当前哈希
  - AC-004-3/E1: 检查未通过时执行 `harness lock`，退出码非 0，锁文件不变
  - AC-004-3/B1: 没有锁文件的新项目首次执行 `harness lock`，生成锁文件

### FR-005 统一语言一致性
- Status: Active
- Release: 1
- Priority: P1

- AC-005-1: 受检产物（PRD、架构声明、详细设计、任务列表）中出现术语表"禁用别名"列的词时报告违规，给出文件和行号。
  - AC-005-1/N1: 受检产物中没有禁用别名，通过
  - AC-005-1/E1: 任务列表第 12 行出现禁用别名 `工单`，失败并给出文件和行号
  - AC-005-1/B1: 禁用别名只出现在代码格式中，不报告
  - AC-005-1/B2: 禁用别名出现在统一语言表格本身的"禁用别名"列，不报告
- AC-005-2: 架构声明中标注为领域概念的名称必须能在术语表"代码名"列中找到。
  - AC-005-2/N1: 架构声明中标注为领域概念的 `Order` 在术语表代码名列中存在，通过
  - AC-005-2/E1: 标注为领域概念的 `Ticket` 不在术语表中，失败
  - AC-005-2/B1: 未标注为领域概念的技术名称（如 `HttpServer`）不在术语表中，不报告

### FR-006 架构设计产物
- Status: Active
- Release: 1
- Priority: P0

提供架构设计 skill，产出宏观到微观三层设计：模块划分与依赖（概要）、模块接口定义（接口）、跨模块调用流程与模块内部设计（详细）。

- AC-006-1: 架构声明文件中每个模块有编号、名称、代码路径范围、允许依赖的模块列表、对外接口列表、覆盖的需求列表。
  - AC-006-1/N1: 模块含全部必填字段，通过
  - AC-006-1/E1: 模块缺少代码路径范围，失败并指出模块编号和字段
  - AC-006-1/B1: 允许依赖列表为空列表，通过（叶子模块）
  - AC-006-1/E2: 两个模块的代码路径范围重叠，失败
- AC-006-2: 每个接口有编号、所属模块、签名定义。
  - AC-006-2/N1: 接口含编号、所属模块、签名，通过
  - AC-006-2/E1: 接口的所属模块不存在，失败
  - AC-006-2/E2: 接口缺少签名，失败
- AC-006-3: 每个调用流程有编号、参与的接口序列、覆盖的需求。
  - AC-006-3/N1: 调用流程含编号、2 步接口序列、覆盖需求，通过
  - AC-006-3/B1: 调用流程只有 1 步，通过
  - AC-006-3/E1: 调用流程的接口序列为空，失败
- AC-006-4: 声明的模块依赖图中存在环时检查失败。
  - AC-006-4/N1: 依赖图无环，通过
  - AC-006-4/E1: A→B→C→A 三模块环，失败并输出环路径
  - AC-006-4/B1: 模块声明依赖自己，判为环并失败
- AC-006-5: 调用流程中每一步从模块 A 调用接口 IF-n 时，IF-n 必须存在，且其所属模块必须在 A 的允许依赖列表中，否则失败。
  - AC-006-5/N1: FLOW 中 MOD-1 调用 IF-2，IF-2 属于 MOD-2，MOD-2 在 MOD-1 允许依赖中，通过
  - AC-006-5/E1: 调用的 IF-9 不存在，失败
  - AC-006-5/E2: IF-2 所属模块不在调用方允许依赖中，失败并给出流程编号和步骤序号
  - AC-006-5/B1: 模块调用自身的接口，不要求自身在允许依赖中，通过

### FR-007 需求到设计的追溯
- Status: Active
- Release: 1
- Priority: P0

`harness check trace` 确定性检查需求与设计元素的对应关系，不使用大模型。

- AC-007-1: 每条 Active 的 FR 至少被一个设计元素覆盖，否则失败并列出未覆盖的 FR。
  - AC-007-1/N1: 每条 Release≤当前的 Active FR 都被至少一个设计元素覆盖，通过
  - AC-007-1/E1: FR-003 未被任何设计元素覆盖，失败并列出 FR-003
  - AC-007-1/B1: FR-003 只被一个状态为 Superseded 的 ADR 覆盖，判为未覆盖并失败
  - AC-007-1/B2: Removed 的 FR 未被覆盖，不报告
- AC-007-2: 每个设计元素至少引用一条需求或一条 ADR，否则报告为孤儿。
  - AC-007-2/N1: 每个设计元素引用至少一条需求或 ADR，通过
  - AC-007-2/E1: 模块 MOD-4 不引用任何需求或 ADR，报告孤儿并失败
- AC-007-3: 任何引用指向不存在或已 Removed 的编号时失败。
  - AC-007-3/N1: 全部引用指向存在且 Active 的编号，通过
  - AC-007-3/E1: 设计元素引用不存在的 FR-099，失败
  - AC-007-3/E2: 设计元素引用已 Removed 的 FR-005，失败
  - AC-007-3/B1: 引用 Superseded 的 ADR，失败并提示改为引用取代它的 ADR
- AC-007-4: 可输出覆盖矩阵（需求 × 设计元素），支持 Markdown 和 JSON。
  - AC-007-4/N1: `harness trace --format md` 输出需求 × 设计元素矩阵，行数等于纳入范围的需求数
  - AC-007-4/N2: `--format json` 输出可被 JSON 解析，内容与 Markdown 一致
  - AC-007-4/B1: 没有任何设计元素时输出空矩阵且不报错
  - AC-007-4/E1: `--format xml` 退出码非 0 并列出支持的格式

### FR-008 模块依赖约束
- Status: Active
- Release: 2
- Priority: P0

`harness check deps` 校验代码的实际依赖是否符合架构声明。

- AC-008-1: 代码中从模块 A 到模块 B 的引用，若 B 不在 A 的允许依赖列表中，报告违规，给出文件和行号。
- AC-008-2: 不属于任何模块路径范围的源文件报告为未归属文件。
- AC-008-3: 支持 TypeScript/JavaScript、Python、Go、Java/Kotlin、C/C++。
- AC-008-4: 声明的允许依赖在代码中从未被使用时，报告为过时声明并失败。

### FR-009 接口变更通知
- Status: Active
- Release: 2
- Priority: P0

接口变更必须被显式登记，并告知受影响的模块。

- AC-009-1: 接口签名的哈希与锁文件不一致，且没有引用该接口的 CHG 条目或 ADR 时，检查失败。
- AC-009-2: 检查输出依赖该接口所属模块的全部模块及其负责人，作为通知清单。
- AC-009-3: 破坏性变更（删除接口、修改签名）在 CHG 条目中必须标注 Breaking。
- AC-009-4: 接口签名有两个来源：详细设计中手写的签名（意图），以及从代码或契约文件提取的快照（事实）。两者不一致时检查失败。
- AC-009-5: 代码中用 `harness:IF-n` 注释标记接口实现处；锚点缺失、重复或不在所属模块路径范围内时检查失败。
- AC-009-6: `harness codeowners` 根据架构声明中模块的负责人生成 CODEOWNERS；接口所在文件的 owner 包含该接口所有下游模块的负责人，配合代码托管平台的分支保护强制审批。

### FR-018 接口快照提取
- Status: Active
- Release: 2
- Priority: P1

从代码或契约文件提取接口签名快照，提交到仓库，作为 FR-009 的"事实"来源。

- AC-018-1: 支持 OpenAPI、protobuf 契约文件，以及 TypeScript/JavaScript、Python、Go、Java/Kotlin、C/C++ 源码中带 `harness:IF-n` 锚点的声明。
- AC-018-2: 快照为规范化文本，同一份代码多次提取结果逐字节一致。

### FR-010 架构决策记录
- Status: Active
- Release: 1
- Priority: P1

- AC-010-1: 每条 ADR 有编号、标题、状态（Proposed / Accepted / Superseded）、背景、决策、后果、关联的需求或设计元素。
  - AC-010-1/N1: ADR 含全部必填字段，通过
  - AC-010-1/E1: 缺少后果字段，失败
  - AC-010-1/E2: 状态取值为 Done，失败
- AC-010-2: Superseded 的 ADR 必须指向取代它的 ADR，被指向的 ADR 必须存在。
  - AC-010-2/N1: ADR-2 为 Superseded 且指向存在的 ADR-5，通过
  - AC-010-2/E1: ADR-2 为 Superseded 但没有指向，失败
  - AC-010-2/E2: 指向不存在的 ADR-9，失败
  - AC-010-2/B1: ADR-2 与 ADR-5 互相指向对方为取代者，判为环并失败
- AC-010-3: 架构声明中模块的增删、允许依赖的变化、扩展点的变化，都必须在同一提交中由一条 ADR 引用，否则失败。
  - AC-010-3/N1: 提交新增一个模块，同一提交新增或修改了引用该模块的 ADR，通过
  - AC-010-3/E1: 提交给 MOD-1 增加允许依赖 MOD-3，同一提交没有 ADR 引用，失败
  - AC-010-3/B1: 提交只修改模块描述文字，不要求 ADR，通过

### FR-011 任务规划
- Status: Active
- Release: 1
- Priority: P0

提供任务规划 skill，根据需求和设计系统性生成任务列表；`harness check tasks` 校验任务与需求、设计的对应关系。

- AC-011-1: 每个任务有编号、标题、状态（Todo / InProgress / Done / Blocked）、关联需求（≥1，可以是 FR、NFR 或验收标准 AC）、关联设计元素（≥1）、文件范围、前置任务。
  - AC-011-1/N1: 任务含全部必填字段，通过
  - AC-011-1/N2: 任务关联需求为 `FR-002, AC-011-5`，通过
  - AC-011-1/E1: 任务没有关联设计元素，失败
  - AC-011-1/E2: 任务状态为 Doing，失败并列出合法状态
  - AC-011-1/B1: 前置任务列表为空，通过
  - AC-011-1/E3: 任务关联 `AC-011-9`，该验收标准不存在，失败
- AC-011-2: 每条 Active 的 FR 被任务覆盖：有任务直接关联该 FR，或该 FR 的每条 Active 验收标准都被某个任务关联（项目处于规划及之后阶段时）。
  - AC-011-2/N1: 每条纳入范围的 Active FR 至少被一个任务关联，通过
  - AC-011-2/E1: FR-007 未被任何任务关联，失败
  - AC-011-2/B1: 项目处于架构阶段（任务列表不存在），不检查此项
  - AC-011-2/B2: FR-011 未被直接关联，但 AC-011-1 ~ AC-011-6 分别被不同任务关联，通过
  - AC-011-2/E2: FR-011 未被直接关联，只有 AC-011-1 ~ AC-011-5 被任务关联，失败并指出 AC-011-6
- AC-011-3: 任务的文件范围必须落在其关联模块的路径范围内。
  - AC-011-3/N1: 任务文件范围 `internal/prd/**` 落在其关联模块路径内，通过
  - AC-011-3/E1: 任务文件范围包含其关联模块之外的 `internal/gate/x.go`，失败
  - AC-011-3/B1: 文件范围与模块路径完全相同，通过
- AC-011-4: 同一时间最多一个 InProgress 任务；两个未完成任务的文件范围重叠且彼此之间没有声明前置依赖时失败。
  - AC-011-4/N1: 只有一个 InProgress 任务，通过
  - AC-011-4/E1: 两个 InProgress 任务，失败
  - AC-011-4/E2: 两个 Todo 任务文件范围重叠且无前置关系，失败并列出重叠的路径
  - AC-011-4/B1: 两个任务范围重叠，但 T-2 声明前置任务 T-1，通过
  - AC-011-4/B2: 重叠的两个任务中一个为 Done，通过
- AC-011-5: 规划 skill 发现需求或设计不清时，先登记待澄清项并回到需求或架构阶段，不生成相应任务。
  - AC-011-5/N1: 规划 skill 文件包含"需求或设计不清时登记待澄清项并停止规划"的指令，静态测试通过
  - AC-011-5/E1: 任务关联的需求存在 Open 的待澄清项时，`harness check tasks` 失败（与 AC-038-3 共用判定）
- AC-011-6: 一个任务的文件范围默认只能落在 1 个模块内；跨模块任务必须引用一个调用流程（FLOW-n），且涉及的模块都在该流程中。
  - AC-011-6/N1: 任务文件范围全部在 1 个模块内，通过
  - AC-011-6/E1: 任务文件范围跨 2 个模块且没有引用 FLOW，失败
  - AC-011-6/B1: 任务跨 2 个模块并引用 FLOW-1，两个模块都在 FLOW-1 中，通过
  - AC-011-6/E2: 任务引用 FLOW-1，但其中一个模块不在 FLOW-1 中，失败

### FR-012 编码门禁
- Status: Active
- Release: 1
- Priority: P0

通过 Claude Code hook 硬拦截不符合流程的代码修改。

- AC-012-1: 没有 InProgress 任务时，对源代码路径的写操作（Edit、Write、Bash 写文件）被拒绝，并告知原因。
  - AC-012-1/N1: 存在 InProgress 任务，写入其范围内的源文件，门禁放行
  - AC-012-1/E1: 没有 InProgress 任务时 Edit 源文件，门禁拒绝并说明原因
  - AC-012-1/E2: 没有 InProgress 任务时 Bash 执行 `echo x > src/a.go`，门禁拒绝
  - AC-012-1/B1: 没有 InProgress 任务时写入产物文件 `docs/PRD.md`，门禁放行
- AC-012-2: 写入路径不在当前 InProgress 任务的文件范围内时被拒绝。
  - AC-012-2/N1: 写入路径在 InProgress 任务文件范围内，放行
  - AC-012-2/E1: 写入路径在范围外，拒绝并给出任务的文件范围
  - AC-012-2/B1: 写入路径通过 `..` 或符号链接解析后落在范围外，拒绝
  - AC-012-2/E2: Bash 命令 `sed -i` 修改范围外文件，拒绝
- AC-012-3: 写入产物文件后自动运行对应的检查，并把结果反馈给代理。
  - AC-012-3/N1: 代理写入 PRD.md 后，门禁运行 `harness check prd` 并把结果反馈给代理
  - AC-012-3/B1: 代理写入非产物文件，不运行产物检查
  - AC-012-3/E1: 写入后检查失败，反馈内容包含规则编号和行号
- AC-012-4: 会话结束前运行全量检查，不通过则阻止结束。
  - AC-012-4/N1: 会话结束时全量检查通过，允许结束
  - AC-012-4/E1: 全量检查失败，阻止结束并反馈失败项
  - AC-012-4/B1: 同一会话连续阻止 3 次（可配置）后第 4 次放行结束，未通过项写入进度文件，下次会话开始时优先注入；任务仍不能标 Done
  - AC-012-4/B2: 连续阻止 2 次后检查通过，正常放行，计数清零
- AC-012-5: 门禁判定逻辑与代理无关，通过代理适配器接入；本 Release 提供 Claude Code 适配器。
  - AC-012-5/N1: Claude Code 适配器把 PreToolUse 输入转换为门禁请求，门禁判定结果转换为 deny 输出
  - AC-012-5/E1: 适配器收到无法解析的输入时，判定为拒绝（失败即阻断）

### FR-013 提交粒度
- Status: Active
- Release: 1
- Priority: P1

两个不相关的改动必须拆开提交。

- AC-013-1: 每个提交信息带 `Task: T-n` 尾注；一个提交只能对应一个任务。
  - AC-013-1/N1: 提交信息含 `Task: T-3`，通过
  - AC-013-1/E1: 提交信息没有 Task 尾注，commit-msg hook 拒绝
  - AC-013-1/E2: 提交信息含两个 Task 尾注，拒绝
  - AC-013-1/E3: Task 尾注引用不存在的 T-99，拒绝
- AC-013-2: 提交中改动的文件必须都在该任务的文件范围内（产物文件和生成文件除外）。
  - AC-013-2/N1: 提交改动文件都在 T-3 文件范围内，通过
  - AC-013-2/B1: 提交同时改动范围内源文件和 `docs/tasks.md`，通过（产物文件除外）
  - AC-013-2/B2: 提交同时改动范围内源文件和 `harness.json`，通过（生成文件除外）
  - AC-013-2/E1: 提交含范围外的源文件，失败并列出该文件
- AC-013-3: `harness init` 安装本地 commit-msg 和 pre-commit git hook；`harness check commits <范围>` 供 CI 兜底，防止 `--no-verify` 绕过。
  - AC-013-3/N1: `harness init` 后 `.git/hooks/commit-msg` 与 `pre-commit` 存在并调用 harness
  - AC-013-3/N2: `harness check commits origin/main..HEAD` 对每个提交执行 AC-013-1/2/4 的检查
  - AC-013-3/E1: 用 `--no-verify` 提交的违规提交，被 `harness check commits` 发现并失败
  - AC-013-3/B1: 非 git 仓库中执行 init，跳过 git hook 安装并输出提示
- AC-013-4: 单个提交改动的行数（新增加删除，不含产物文件和生成文件）超过配置上限（默认 400）时失败。
  - AC-013-4/N1: 提交改动 400 行，通过
  - AC-013-4/E1: 提交改动 401 行，失败并给出行数和上限
  - AC-013-4/B1: 提交改动 1000 行，其中 900 行在产物文件和生成文件中，通过

### FR-014 语言配置包
- Status: Active
- Release: 2
- Priority: P0

为每种支持的语言提供格式化、静态检查、编码规范配置，`harness check lint` 统一调用。

- AC-014-1: 支持 TypeScript/JavaScript、Python、Go、Java/Kotlin、C/C++。
- AC-014-2: 语言工具缺失时给出安装提示并失败，不静默跳过。

### FR-015 UI 设计规范
- Status: Active
- Release: 3
- Priority: P1

提供 UI 规范 skill，产出设计规范和设计 token；`harness check ui` 禁止使用规范外的颜色和样式。

- AC-015-1: 设计 token 以 W3C DTCG 格式存放于单一文件，可生成 CSS 变量和 Tailwind 主题。
- AC-015-2: 样式文件和界面代码文件中出现 token 以外的颜色字面量、字号、间距、圆角、阴影值，或 Tailwind 任意值时报告违规，给出文件和行号。
- AC-015-3: 设计 token 中声明的前景色与背景色组合必须满足 WCAG 2.2 AA 对比度（正文 4.5:1，大字号 3:1），否则失败。

### FR-016 流程 skill 集
- Status: Active
- Release: 1
- Priority: P0

提供覆盖各阶段的 skill：需求澄清、需求变更、架构设计、任务规划、编码实现、UI 规范。

- AC-016-1: 每个 skill 结束前必须运行对应的 `harness check` 子命令并通过。
  - AC-016-1/N1: 每个 skill 文件声明其结束前要运行的 `harness check` 子命令，静态测试通过
  - AC-016-1/E1: 某个 skill 文件缺少该声明，harness 自身构建失败
- AC-016-2: 提供测试 skill，指导根据验收用例编写集成测试、根据接口和核心模块编写单元测试。
  - AC-016-2/N1: 测试 skill 文件存在，包含按验收用例编号命名测试的约定
  - AC-016-2/E1: 测试 skill 中的命名约定与 `harness check tests` 使用的提取规则不一致时，harness 自身测试失败

### FR-017 自举
- Status: Active
- Release: 1
- Priority: P0

harness 自身的开发受 harness 管理。

- AC-017-1: 本仓库的 PRD、架构声明、任务列表通过 `harness check`。
  - AC-017-1/N1: 在本仓库执行 `harness check`，退出码 0
  - AC-017-1/E1: 向本仓库 PRD 加入一条未被设计覆盖的 Release 1 需求，`harness check` 失败（CI 回归测试）
- AC-017-2: 本仓库启用 harness 自己的 hook。
  - AC-017-2/N1: 本仓库 `.claude/settings.json` 的 hook 指向本仓库构建的 harness 二进制
  - AC-017-2/E1: hook 配置与 `harness init` 生成的内容不一致时，AC-041-2 的检查失败

### FR-019 验收用例声明
- Status: Active
- Release: 1
- Priority: P0

每条验收标准必须显式列出要测试的用例，区分正常、边界、错误三类，使"是否考虑了边界"成为可检查的 PRD 内容。

- AC-019-1: 用例编号格式为 `AC-<需求号>-<序号>/<类别><序号>`，类别为 N（正常）、B（边界）、E（错误），例如 `AC-001-1/B1`。
  - AC-019-1/N1: `AC-001-1/B1` 判为合法
  - AC-019-1/E1: `AC-001-1/X1` 判为非法类别并失败
  - AC-019-1/E2: 用例前缀与其所在验收标准不一致（AC-001-2 下出现 `AC-001-1/N1`），失败
- AC-019-2: 每条 Active 需求的每条验收标准至少有一个 N 用例和一个 B 或 E 用例，否则 `harness check prd` 失败。
  - AC-019-2/N1: 验收标准有 N1 和 B1，通过
  - AC-019-2/B1: 验收标准只有 N1 和 E1（没有 B），通过
  - AC-019-2/E1: 验收标准只有 N1，失败
  - AC-019-2/E2: 验收标准只有 B1 和 E1（没有 N），失败
  - AC-019-2/B2: Release 大于 Current-Release 的需求没有用例，不报告
- AC-019-3: 用例编号只能新增或标记 Removed，不能删除或复用。
  - AC-019-3/N1: 新增用例 `/B2`，通过
  - AC-019-3/E1: 锁文件中存在的用例 `/E1` 从文档中消失，失败
  - AC-019-3/E2: 已 Removed 的用例编号被重新用于新用例，失败

### FR-020 集成测试追溯
- Status: Active
- Release: 1
- Priority: P0

`harness check tests` 校验每个验收用例都被真实运行并通过的集成测试覆盖。

- AC-020-1: 测试名中包含用例编号即视为引用该用例；一个测试可引用多个用例。
  - AC-020-1/N1: 测试名 `TestInit_AC_001_1_N1` 被识别为引用 `AC-001-1/N1`
  - AC-020-1/B1: 一个测试名中包含两个用例编号，两个用例都被记为引用
  - AC-020-1/E1: 测试名中的编号指向不存在的用例，失败
- AC-020-2: 检查读取测试运行生成的 JUnit XML 报告（Go 使用 `go test -json`），每个 Active 用例至少有一个状态为 passed 的集成测试，否则失败。
  - AC-020-2/N1: 每个纳入范围的 Active 用例都有一个 passed 的集成测试，通过
  - AC-020-2/E1: `AC-001-1/E1` 只有 failed 的测试，失败
  - AC-020-2/E2: 没有提供测试报告，失败并提示生成报告的命令
  - AC-020-2/B1: 同一用例一个测试 passed、另一个 failed，整体因 failed 测试失败
- AC-020-3: 状态为 skipped 的测试不计入覆盖；源码中声明了引用、报告中却没有出现的测试判为失败。
  - AC-020-3/N1: 引用用例的测试全部出现在报告中且 passed，通过
  - AC-020-3/E1: 引用某用例的唯一测试状态为 skipped，判为未覆盖并失败
  - AC-020-3/E2: 源码中声明引用了用例的测试在报告中不存在，失败
- AC-020-4: 集成测试与单元测试按语言配置包约定的目录或命名区分。
  - AC-020-4/N1: Go 配置包中 `*_integration_test.go` 或 `//go:build integration` 的测试判为集成测试
  - AC-020-4/E1: 只被单元测试引用的验收用例，判为未覆盖并失败

### FR-021 单元测试追溯
- Status: Active
- Release: 2
- Priority: P0

对外接口和核心逻辑必须有单元测试。

- AC-021-1: 每个 Active 接口至少有一个测试名包含其编号（IF-n）且状态为 passed 的单元测试。
- AC-021-2: 单独运行某接口的单元测试并采集覆盖率时，`harness:IF-n` 锚点所在函数的入口行必须被执行，否则判定该测试未真正调用接口。
- AC-021-3: 架构声明中每个模块可设置行覆盖率阈值；标记为核心的模块必须设置阈值。实际覆盖率低于阈值时失败。

### FR-022 测试有效性
- Status: Active
- Release: 2
- Priority: P0

确定性检测无效测试和削弱测试的行为。

- AC-022-1: 被追溯的测试函数中没有断言，或只有恒真断言时失败。
- AC-022-2: 接口 IF-n 的单元测试中 mock 了 IF-n 所属模块时失败。
- AC-022-3: 相对基线分支，被追溯测试的断言数减少、被追溯测试被删除、新增跳过标记时，必须有引用对应用例或接口的 CHG 条目，否则失败。

### FR-023 变异测试
- Status: Active
- Release: 3
- Priority: P2

- AC-023-1: `harness check mutation` 只对相对基线分支改动的代码运行变异测试，结果作为告警输出，不影响退出码。
- AC-023-2: 语言没有可用的变异测试工具时输出说明并跳过。

### FR-024 会话交接
- Status: Active
- Release: 1
- Priority: P1

长任务跨多个会话时，每个会话从一致的状态开始，结束时留下可接续的进度。

- AC-024-1: 会话开始时，门禁向代理注入当前阶段、InProgress 任务、未通过的检查、未关闭的待澄清项。
  - AC-024-1/N1: 会话开始时，注入内容包含当前阶段、InProgress 任务编号、未通过检查数、Open 待澄清项
  - AC-024-1/B1: 没有任务列表和待澄清项时，注入内容为"阶段：需求，无进行中任务"
- AC-024-2: 任务状态变为 Done 时，进度文件（`docs/progress.md`）必须追加一条记录（任务编号、日期、结论、遗留问题），否则检查失败。
  - AC-024-2/N1: T-3 变为 Done，`docs/progress.md` 有 T-3 记录，通过
  - AC-024-2/E1: T-3 变为 Done 但进度文件中没有 T-3 记录，失败
  - AC-024-2/E2: T-3 的记录缺少结论字段，失败
- AC-024-3: 项目声明冒烟测试命令时，会话开始先运行一次，并把结果注入代理。
  - AC-024-3/N1: 项目声明冒烟命令且命令成功，注入"冒烟通过"
  - AC-024-3/E1: 冒烟命令失败，注入失败输出的最后 50 行
  - AC-024-3/B1: 未声明冒烟命令，不运行，不报错

### FR-025 安全边界
- Status: Active
- Release: 1
- Priority: P0

- AC-025-1: 代理读取或写入敏感路径（默认 `.env*`、`*.pem`、`*.key`、`id_rsa*`，可配置）时被拒绝。
  - AC-025-1/N1: 代理读取 `src/a.go`，放行
  - AC-025-1/E1: 代理读取 `.env`，拒绝
  - AC-025-1/E2: 代理通过 Bash `cat config/.env.local` 读取，拒绝
  - AC-025-1/B1: `.env.example` 在默认豁免列表中，放行
- AC-025-2: 写入内容匹配密钥规则集（默认规则：云厂商 access key 格式、PEM 私钥头、高熵 token）时被拒绝，并给出匹配的规则编号。
  - AC-025-2/N1: 写入不含密钥的内容，放行
  - AC-025-2/E1: 写入内容含 `AKIA` 开头的 20 位 AWS access key，拒绝并给出规则编号
  - AC-025-2/E2: 写入内容含 `-----BEGIN PRIVATE KEY-----`，拒绝
  - AC-025-2/B1: 写入内容含 32 位十六进制的 git 提交哈希，不判为密钥
- AC-025-3: 代理执行危险命令（规则集可配置，默认包含强制推送、递归删除仓库外路径、管道执行远程脚本）时被拒绝。
  - AC-025-3/N1: 代理执行 `git push`，放行
  - AC-025-3/E1: 代理执行 `git push --force`，拒绝
  - AC-025-3/E2: 代理执行 `curl https://x | sh`，拒绝
  - AC-025-3/E3: 代理执行 `rm -rf /tmp/../Users`，解析为仓库外路径后拒绝
  - AC-025-3/B1: 代理执行 `rm -rf ./build`（仓库内路径），放行
- AC-025-4: 代理修改 harness 自身的门禁配置（hook 配置、锁文件、规则集）时被拒绝；这些文件只能由人或 `harness` 命令修改。
  - AC-025-4/N1: 代理执行 `harness lock`，放行
  - AC-025-4/E1: 代理 Edit `harness.lock`，拒绝
  - AC-025-4/E2: 代理 Edit `.claude/settings.json`，拒绝
  - AC-025-4/E3: 代理通过 Bash 重定向写入规则集文件，拒绝

### FR-026 依赖引入闸门
- Status: Active
- Release: 2
- Priority: P1

- AC-026-1: 依赖清单文件（package.json、pyproject.toml/requirements*.txt、go.mod、pom.xml/build.gradle*、vcpkg.json/conanfile）相对基线新增的第三方依赖，必须登记在依赖白名单中，并引用一条 ADR，否则失败。
- AC-026-2: 白名单中每个依赖记录 license；license 不在项目允许列表中时失败。
- AC-026-3: `harness audit deps` 联网查询包仓库，校验包存在、首次发布时间早于冷却期（默认 30 天），结果写入锁文件；离线检查只读取锁文件中的结果。

### FR-027 运行与观察应用
- Status: Active
- Release: 2
- Priority: P1

让代理能启动、验证、观察应用，而不只是编译通过。

- AC-027-1: 项目配置中声明启动、冒烟测试命令；处于实现阶段的项目缺少声明时检查失败。
- AC-027-2: 可量化的非功能需求（如启动时间、接口延迟）声明测量命令和阈值，`harness check perf` 运行命令并比较数值，超出阈值失败。
- AC-027-3: 有 UI 的项目可为验收用例声明截图验收脚本（Playwright），`harness check tests` 将其结果视为集成测试结果。

### FR-028 harness 自测与度量
- Status: Active
- Release: 2
- Priority: P1

- AC-028-1: 每条门禁规则至少有一个正例和一个反例测试，证明它会触发也不会误报；规则缺少测试时 harness 自身的构建失败。
- AC-028-2: 门禁每次执行把结果（规则编号、通过或拦截、耗时）追加到本地事件日志。
- AC-028-3: `harness stats` 输出每条规则的拦截次数、任务首次检查通过率、指定天数内从未触发的规则。

### FR-029 独立评审
- Status: Active
- Release: 3
- Priority: P2

- AC-029-1: 提供评审 skill，在新上下文中对照 PRD 验收标准和设计审查改动，输出问题清单。
- AC-029-2: 评审结论只作建议，不影响任何门禁的结果。

### FR-030 文档引用与过期检查
- Status: Active
- Release: 2
- Priority: P1

确定性发现自由文本文档（README、详细设计说明、入口指令文件、skill）的过期内容，替代定期清理代理。

- AC-030-1: 受检 Markdown 中以代码格式出现的仓库相对路径必须存在，否则失败。
- AC-030-2: 受检 Markdown 中出现的 `harness <子命令>` 必须是存在的子命令，否则失败。
- AC-030-3: 受检 Markdown 中引用的编号（FR、NFR、AC、IF、MOD、FLOW、ADR、T、DEBT）必须存在且不是 Removed，否则失败。
- AC-030-4: 模块路径下的代码在其详细设计文档最后一次修改之后被提交改动超过 N 次（默认 10，可配置）时输出"可能过期"告警，不影响退出码。

### FR-031 技术债登记
- Status: Active
- Release: 1
- Priority: P0

- AC-031-1: 技术债清单（`docs/debt.md`）中每项有编号 DEBT-n、所属模块、描述、状态（Open / Resolved）。
  - AC-031-1/N1: 技术债项含编号、所属模块、描述、状态，通过
  - AC-031-1/E1: 技术债项的所属模块不存在，失败
  - AC-031-1/E2: 状态取值为 Done，失败
- AC-031-2: 源码中的 `TODO`、`FIXME`、`HACK`、`XXX` 注释必须在同一行引用一个 Open 的 DEBT-n 或一个未完成的 T-n，否则失败。
  - AC-031-2/N1: `// TODO(DEBT-2): 重试` 引用 Open 的 DEBT-2，通过
  - AC-031-2/B1: `# FIXME T-5` 引用未完成的 T-5，通过
  - AC-031-2/E1: 裸写 `// TODO 以后再说`，失败并给出文件和行号
  - AC-031-2/E2: `// TODO(T-5)` 引用的 T-5 已 Done，失败
  - AC-031-2/B2: 字符串字面量 `"TODO"` 不判为注释，不报告
- AC-031-3: 状态为 Resolved 的 DEBT-n 仍被源码引用时失败。
  - AC-031-3/N1: Resolved 的 DEBT-1 没有被引用，通过
  - AC-031-3/E1: Resolved 的 DEBT-1 仍被源码引用，失败并列出引用位置

### FR-032 模块健康报告
- Status: Active
- Release: 2
- Priority: P2

替代代码质量评级：只列事实，不打分。

- AC-032-1: `harness stats --modules` 按模块列出：行覆盖率与阈值、依赖违规数、无效测试数、Open 技术债数、未覆盖的验收用例数、近 30 天门禁拦截次数。
- AC-032-2: 报告支持 Markdown 和 JSON 输出。

### FR-033 扩展场景与扩展点
- Status: Active
- Release: 1
- Priority: P0

可扩展性必须写成可验证的变更场景，并由架构中的扩展点承载。

- AC-033-1: 扩展性需求以非功能需求的形式写成变更场景，声明"新增什么"、"只允许新增或修改哪些路径"、"哪些模块不得修改"。
  - AC-033-1/N1: Type 为 Scenario 的 NFR 含"新增什么""允许路径""封闭模块"三个字段，通过
  - AC-033-1/E1: Scenario 类 NFR 缺少封闭模块字段，失败
- AC-033-2: 架构声明中每个扩展点有编号 EXT-n、承载接口（IF-n）、扩展代码路径、封闭模块列表、关联的扩展性需求。
  - AC-033-2/N1: EXT-1 含全部必填字段，通过
  - AC-033-2/E1: EXT-1 的承载接口不存在，失败
  - AC-033-2/E2: EXT-1 的封闭模块列表包含不存在的模块，失败
- AC-033-3: 扩展点没有关联扩展性需求时报告为过度设计并失败。
  - AC-033-3/N1: EXT-1 关联 NFR-006，通过
  - AC-033-3/E1: EXT-2 没有关联任何 Scenario 类 NFR，报告过度设计并失败
  - AC-033-3/E2: EXT-2 关联的是 Constraint 类 NFR，失败
- AC-033-4: 每个扩展点至少有一个测试名包含其编号、状态为 passed 的示例扩展测试。
  - AC-033-4/N1: 测试名含 `EXT-1` 的测试在报告中 passed，通过
  - AC-033-4/E1: 没有任何测试名含 `EXT-1`，失败
  - AC-033-4/B1: EXT-1 关联的扩展场景 Release 大于 Current-Release，不检查
- AC-033-5: 需求澄清 skill 必须询问"未来哪些东西会变化"，并把回答登记为扩展场景或明确记录为"不预留扩展"。
  - AC-033-5/N1: PRD 至少有一个 Type 为 Scenario 的 NFR，通过
  - AC-033-5/B1: PRD 没有 Scenario 类 NFR，但非功能需求章节写明"不预留扩展"，通过
  - AC-033-5/E1: PRD 既没有 Scenario 类 NFR，也没有"不预留扩展"声明，失败

### FR-034 开闭校验
- Status: Active
- Release: 1
- Priority: P0

- AC-034-1: 任务可声明"扩展 EXT-n"；这类任务的提交改动了该扩展点封闭模块路径下的文件时失败。
  - AC-034-1/N1: 扩展 EXT-1 的任务提交只改扩展路径，通过
  - AC-034-1/E1: 扩展 EXT-1 的任务提交改动了封闭模块的文件，失败并列出文件
  - AC-034-1/B1: 非扩展类任务改动该封闭模块，不受此规则约束
- AC-034-2: 这类任务的文件范围与封闭模块路径重叠时，`harness check tasks` 失败。
  - AC-034-2/N1: 扩展任务的文件范围不含封闭模块路径，通过
  - AC-034-2/E1: 扩展任务的文件范围包含封闭模块路径，`harness check tasks` 失败

### FR-035 模块封装
- Status: Active
- Release: 2
- Priority: P1

- AC-035-1: 每个模块在架构声明中列出公开路径（对外接口所在文件或目录）。
- AC-035-2: 跨模块引用的目标文件不在被引用模块的公开路径内时失败，给出文件和行号。

### FR-036 依赖稳定性与扇出
- Status: Active
- Release: 2
- Priority: P1

- AC-036-1: 按实际 import 图计算每个模块的不稳定度 I = 出向依赖数 /（出向依赖数 + 入向依赖数）。
- AC-036-2: 模块依赖了不稳定度比自己高出配置容差（默认 0）的模块时失败；允许在架构声明中逐条豁免，豁免必须引用 ADR。
- AC-036-3: 模块直接依赖的模块数超过配置上限（默认 7）时失败。

### FR-037 人工审批
- Status: Active
- Release: 1
- Priority: P0

产物的批准只能由人完成，批准后内容变化自动失效。

- AC-037-1: PRD、架构声明、详细设计、ADR 的状态只能通过 `harness approve <产物>` 变为 Approved；该命令要求交互式终端，非交互环境下拒绝执行。
  - AC-037-1/N1: 人在终端执行 `harness approve prd` 并确认，PRD 状态变为 Approved，锁文件新增批准记录
  - AC-037-1/E1: 标准输入不是终端（管道或代理执行）时，`harness approve` 拒绝并退出码非 0
  - AC-037-1/B1: 在确认提示处输入 n，状态不变，锁文件不变
  - AC-037-1/E2: PRD 存在检查失败项时执行 approve，拒绝批准
- AC-037-2: 门禁拒绝代理直接把产物状态改为 Approved，也拒绝代理执行 `harness approve`。
  - AC-037-2/N1: 代理编辑 PRD 的正文，放行
  - AC-037-2/E1: 代理把 PRD 的 `Status: Draft` 改为 `Status: Approved`，拒绝
  - AC-037-2/E2: 代理通过 Bash 执行 `harness approve prd`，拒绝
- AC-037-3: 锁文件记录每次批准的批准人、时间、被批准内容的哈希；产物内容与批准哈希不一致时，`harness check` 将其视为 Draft 并报告"批准已失效"。
  - AC-037-3/N1: 批准后内容未变，检查报告 PRD 为 Approved
  - AC-037-3/E1: 批准后修改一条验收标准，检查报告"批准已失效"，PRD 视为 Draft
  - AC-037-3/B1: 批准后只改变更日志之外的空白字符，规范化哈希不变，批准仍有效

### FR-038 阶段门禁
- Status: Active
- Release: 1
- Priority: P0

- AC-038-1: PRD 未处于有效的 Approved 状态时，任务不能进入 InProgress。
  - AC-038-1/N1: PRD 为有效 Approved，任务改为 InProgress，通过
  - AC-038-1/E1: PRD 为 Draft 时任务改为 InProgress，失败
  - AC-038-1/E2: PRD 批准已失效时任务改为 InProgress，失败
- AC-038-2: 任务引用的设计元素所在产物未处于有效的 Approved 状态时，该任务不能进入 InProgress。
  - AC-038-2/N1: 任务引用的设计元素所在产物有效 Approved，通过
  - AC-038-2/E1: 任务引用的模块所在的架构声明为 Draft，失败
- AC-038-3: 任务关联的需求存在 Open 的待澄清项时，该任务不能进入 InProgress。
  - AC-038-3/N1: 任务关联需求没有 Open 的待澄清项，通过
  - AC-038-3/E1: 待澄清项 OQ-3 关联 FR-002 且 Open，关联 FR-002 的任务改为 InProgress，失败

### FR-039 任务完成定义
- Status: Active
- Release: 1
- Priority: P0

- AC-039-1: 任务状态为 Done 时，其关联需求涉及的 Active 验收用例都必须在测试报告中有通过的测试，否则失败。关联 FR 或 NFR 时涉及其下全部用例，关联 AC 时只涉及该 AC 下的用例。
  - AC-039-1/N1: T-3 为 Done，其关联需求的全部用例在报告中 passed，通过
  - AC-039-1/E1: T-3 为 Done，其关联需求下 `AC-002-1/B1` 没有 passed 测试，失败
  - AC-039-1/B1: 关联需求下 Removed 的用例没有测试，不影响
  - AC-039-1/B2: T-3 为 Done，只关联 AC-011-5，AC-011-5 的用例 passed，FR-011 其他用例没有测试，通过
- AC-039-2: 任务状态为 Done 时，`harness check` 全部通过，且进度文件中有该任务的记录（见 AC-024-2）。
  - AC-039-2/N1: T-3 为 Done，全量检查通过且进度文件有记录，通过
  - AC-039-2/E1: T-3 为 Done，全量检查有失败项，失败

### FR-040 需求文本质量
- Status: Active
- Release: 1
- Priority: P0

参照 INCOSE《Guide for Writing Requirements》中可词法检查的规则。

- AC-040-1: FR 描述和验收标准中出现模糊词表中的词时失败，给出行号；默认词表为：`快速`、`友好`、`适当`、`尽量`、`等等`、`之类`、`如有必要`、`和/或`、`大约`、`灵活`、`高效`、`包括但不限于`（可配置）。代码格式中的文本不参与检查。
  - AC-040-1/N1: 验收标准文本不含模糊词，通过
  - AC-040-1/E1: 验收标准含 `响应要快速`，失败并给出行号和命中的词
  - AC-040-1/B1: 模糊词出现在代码格式中，不报告
  - AC-040-1/B2: `等待`、`相等` 中的 `等` 不命中词表（词表只含 `等等`）
- AC-040-2: 每条 NFR 有 Type 字段（Quantitative / Constraint / Scenario）；Type 为 Quantitative 的 NFR 每条验收标准至少包含一个数值和单位，否则失败。
  - AC-040-2/N1: Quantitative 类 NFR 的验收标准含"< 200ms"，通过
  - AC-040-2/E1: Quantitative 类 NFR 的验收标准为"响应及时"，失败
  - AC-040-2/B1: Constraint 类 NFR 的验收标准不含数值，通过
  - AC-040-2/E2: NFR 缺少 Type 字段，失败

### FR-041 门禁配置锁定
- Status: Active
- Release: 1
- Priority: P0

- AC-041-1: 门禁配置（覆盖率阈值、扇出上限、模糊词表、规则集、豁免列表）纳入锁文件；内容变化且没有引用它的 CHG 条目或 ADR 时失败。
  - AC-041-1/N1: 修改覆盖率阈值并新增引用该配置的 CHG 条目，通过
  - AC-041-1/E1: 把模糊词表删掉一个词且没有 CHG 或 ADR，失败
  - AC-041-1/B1: 门禁配置文件只改了注释，规范化哈希不变，通过
- AC-041-2: 由 harness 生成的文件（如 CODEOWNERS、主题文件、各语言工具配置）内容与重新生成的结果不一致时失败。
  - AC-041-2/N1: 生成文件与重新生成结果一致，通过
  - AC-041-2/E1: 手工修改 CODEOWNERS 一行，失败并提示重新生成的命令
  - AC-041-2/B1: 生成文件不存在时，提示运行生成命令并失败

### FR-042 抑制注释登记
- Status: Active
- Release: 2
- Priority: P1

- AC-042-1: 源码中的 lint 和类型检查抑制注释（`eslint-disable*`、`@ts-ignore`、`@ts-expect-error`、`# noqa`、`# type: ignore`、`# pylint: disable`、`//nolint`、`@SuppressWarnings`、`// NOLINT`）必须在同一行引用一个 Open 的 DEBT-n，否则失败。

### FR-043 CI 集成
- Status: Active
- Release: 1
- Priority: P0

- AC-043-1: `harness init` 生成 CI 配置（GitHub Actions 与 GitLab CI 模板），其中运行 `harness check` 和 `harness check commits`。
  - AC-043-1/N1: `harness init` 生成 `.github/workflows/harness.yml`，含 `harness check` 与 `harness check commits`
  - AC-043-1/N2: `harness init --ci gitlab` 生成 `.gitlab-ci.yml` 片段
  - AC-043-1/E1: `--ci` 取值不是 github 或 gitlab 时，退出码非 0 并列出支持的取值
- AC-043-2: 项目中存在的 CI 配置文件都不包含 `harness check` 步骤时失败。
  - AC-043-2/N1: 存在包含 `harness check` 的 CI 配置，通过
  - AC-043-2/E1: `.github/workflows/` 下有配置但都不含 `harness check`，失败
  - AC-043-2/B1: 项目没有任何 CI 配置文件，失败并提示运行 `harness init --ci`

### FR-044 多代理门禁
- Status: Active
- Release: 3
- Priority: P1

- AC-044-1: 提供 Codex 和 Cursor 适配器，接入 FR-012 的门禁判定。
- AC-044-2: 某个代理无法在写入前拦截时，用该代理可用的最强替代手段（写入后回滚或结束时阻断），并在文档中写明差异。

### FR-045 流程阶段定义
- Status: Active
- Release: 1
- Priority: P0

流程阶段以数据定义，内置阶段与项目新增阶段使用同一格式，阶段推进由人完成。

- AC-045-1: 每个阶段定义包含名称、顺序、进入规则列表、退出规则列表、可写产物路径、关联 skill；内置的需求、架构、规划、实现四个阶段用同一格式定义。
  - AC-045-1/N1: 内置四个阶段的定义通过校验
  - AC-045-1/B1: 第一个阶段的进入规则列表为空，通过
  - AC-045-1/E1: 阶段定义缺少退出规则字段，失败
  - AC-045-1/E2: 阶段定义引用不存在的规则编号，失败
- AC-045-2: 项目配置可新增阶段，或在已有阶段的进入、退出规则列表中追加规则；移除内置规则必须引用一条 ADR。
  - AC-045-2/N1: 项目配置新增"发布"阶段，顺序排在实现之后，`harness check` 通过，`harness phase list` 中出现该阶段
  - AC-045-2/N2: 项目配置给架构阶段的退出规则追加一条项目规则，推进时该规则被执行
  - AC-045-2/E1: 项目配置移除内置规则且没有引用 ADR，失败
  - AC-045-2/E2: 新增阶段与已有阶段同名，失败
- AC-045-3: 当前阶段记录在项目状态文件中，只能通过 `harness phase next` 推进；推进要求当前阶段全部退出规则和下一阶段全部进入规则通过，并要求交互式终端。
  - AC-045-3/N1: 规则全部通过，人在终端执行 `harness phase next`，当前阶段前进一步
  - AC-045-3/E1: 当前阶段有退出规则未通过，拒绝推进并列出未通过的规则
  - AC-045-3/E2: 非交互环境执行 `harness phase next`，拒绝
  - AC-045-3/E3: 代理编辑项目状态文件中的当前阶段，门禁拒绝
  - AC-045-3/B1: 已处于最后一个阶段时执行 `harness phase next`，提示已是最后阶段，退出码非 0
- AC-045-4: 门禁按当前阶段的可写产物路径判定产物写入；源码写入仍按 FR-012 判定。
  - AC-045-4/N1: 规划阶段写入 `docs/tasks.md`，放行
  - AC-045-4/E1: 需求阶段写入 `docs/tasks.md`，拒绝并给出当前阶段的可写路径
  - AC-045-4/B1: 实现阶段写入 `docs/PRD.md`，放行（实现阶段的可写路径包含 PRD，需求变更走 CHG）

### FR-046 项目自定义规则
- Status: Active
- Release: 2
- Priority: P1

使用 harness 的项目不重新编译即可增加硬检测规则。

- AC-046-1: 支持声明式规则：必填字段、枚举取值、引用存在、禁用正则，作用于指定产物类型或路径 glob。
- AC-046-2: 支持外部命令规则：harness 通过标准输入传入项目模型 JSON，命令通过标准输出返回违规 JSON（规则编号、文件、行号、消息、修复建议）；命令退出码非 0 或输出不合法时检查失败。
- AC-046-3: 自定义规则必须在项目配置中登记，纳入锁文件（见 FR-041）。
- AC-046-4: 每条自定义规则必须提供至少一个违规样本和一个合规样本；`harness check` 先运行样本，违规样本未被报告或合规样本被报告时，判定该规则失效并失败。
- AC-046-5: 自定义规则编号使用项目前缀，与内置规则编号冲突时失败。
- AC-046-6: 外部命令规则不得联网，由文档约定，harness 不强制。

## 4. 非功能需求

### NFR-001 确定性
- Status: Active
- Release: 1
- Type: Constraint
- Priority: P0

所有检查不调用大模型，相同输入产生相同输出。

- AC-N001-1: 除 `harness audit deps` 外，检查工具不发起任何网络请求。
  - AC-N001-1/N1: 在断网环境运行全部检查子命令，结果与联网时一致
  - AC-N001-1/E1: 源码静态测试：除 audit 包外引入 `net/http` 等网络包时 harness 自身构建失败

### NFR-002 单二进制分发
- Status: Active
- Release: 1
- Type: Constraint
- Priority: P0

检查工具用 Go 实现，编译为无运行时依赖的单个二进制。

- AC-N002-1: 提供 macOS（arm64/amd64）、Linux（amd64/arm64）、Windows（amd64/arm64）构建。
  - AC-N002-1/N1: 发布构建产出 6 个平台的二进制
  - AC-N002-1/E1: 任一平台交叉编译失败时，发布流程失败
- AC-N002-2: skill、hook 配置、rule、产物模板、语言配置包通过 go:embed 打包进二进制，`harness init` 写入项目；项目自包含，团队成员 clone 后只需安装二进制。
  - AC-N002-2/N1: 在只安装了 harness 二进制的机器上，`harness init` 生成全部模板
  - AC-N002-2/E1: 内嵌模板缺失（构建时被排除）时，harness 自身测试失败
- AC-N002-3: 二进制通过 GitHub Releases 和 Homebrew tap 分发。
  - AC-N002-3/N1: 打 tag 后，GitHub Release 包含 6 个平台的二进制和校验和
  - AC-N002-3/N2: Homebrew tap 的 formula 版本与 tag 一致
  - AC-N002-3/E1: Release 中某个二进制的校验和与文件不一致时，发布流程失败

### NFR-003 性能
- Status: Active
- Release: 1
- Type: Quantitative
- Priority: P1

- AC-N003-1: 写操作门禁（PreToolUse）单次耗时 < 200ms。
  - AC-N003-1/N1: 5000 个源文件、200 个任务的样例仓库上，PreToolUse 门禁 p95 < 200ms
  - AC-N003-1/E1: 超过 200ms 时性能基准测试失败
- AC-N003-2: 5000 个源文件的仓库全量检查 < 10s。
  - AC-N003-2/N1: 5000 个源文件的样例仓库全量检查 < 10s
  - AC-N003-2/B1: 50000 行的 PRD 上 `harness check prd` < 2s

### NFR-004 可操作的报错
- Status: Active
- Release: 1
- Type: Constraint
- Priority: P0

- AC-N004-1: 每条违规包含规则编号、文件、行号、修复建议。
  - AC-N004-1/N1: 每条违规输出含规则编号、文件、行号、修复建议
  - AC-N004-1/E1: harness 自身测试：任一规则的违规输出缺少修复建议时失败
- AC-N004-2: 支持 `--json` 输出；有违规时退出码非 0。
  - AC-N004-2/N1: `--json` 输出可被 JSON 解析，字段与文本输出一致；有违规时退出码 1，无违规时 0
  - AC-N004-2/E1: PRD 文件不可读等内部错误时退出码为 2，与"有违规"区分

### NFR-005 产物可读可 diff
- Status: Active
- Release: 1
- Type: Constraint
- Priority: P1

- AC-N005-1: 所有产物为 Markdown 或 YAML/JSON 文本，人可直接编辑，git diff 可读。
  - AC-N005-1/N1: `harness init` 生成的全部产物扩展名为 .md、.yaml 或 .json
  - AC-N005-1/E1: 锁文件为 JSON 并按键排序，同一状态两次生成逐字节一致

### NFR-006 扩展场景：新增一种编程语言
- Status: Active
- Release: 2
- Type: Scenario
- Priority: P1
- Adds: 一种项目语言的支持
- Allowed-Paths: 该语言的语言配置包、提取器文件、语言注册表文件
- Closed-Modules: 需求、追溯、任务、门禁

新增一种项目语言的支持时，只新增该语言的语言配置包和 import/接口提取器文件，并在语言注册表中登记一行；不修改需求、追溯、任务、门禁模块。

- AC-N006-1: 以示例语言实现一次扩展，提交只涉及语言配置包目录、提取器目录和注册表文件。

### NFR-007 扩展场景：新增一种编程代理
- Status: Active
- Release: 3
- Type: Scenario
- Priority: P1
- Adds: 一种编程代理的门禁支持
- Allowed-Paths: 该代理的 hook 适配器、配置模板、代理注册表文件
- Closed-Modules: 门禁判定

新增一种编程代理（如 Codex、Cursor 之外的代理）的门禁支持时，只新增该代理的 hook 适配器和配置模板，并在代理注册表中登记一行；不修改门禁判定逻辑。

- AC-N007-1: 以示例代理实现一次扩展，提交只涉及适配器目录、模板目录和注册表文件。

### NFR-008 扩展场景：新增一个流程阶段
- Status: Active
- Type: Scenario
- Release: 1
- Priority: P1
- Adds: 一个内置流程阶段
- Allowed-Paths: 阶段定义文件、该阶段的 skill 模板、阶段注册表文件
- Closed-Modules: 门禁判定、规则引擎、产物解析

- AC-N008-1: 以示例阶段实现一次扩展，提交只涉及阶段定义文件、skill 模板和阶段注册表文件。
  - AC-N008-1/N1: 示例阶段的扩展提交通过 FR-034 开闭校验
  - AC-N008-1/E1: 示例阶段的提交改动门禁判定模块时，开闭校验失败（回归测试）

### NFR-009 扩展场景：新增一条内置硬检测规则
- Status: Active
- Type: Scenario
- Release: 1
- Priority: P1
- Adds: 一条内置规则
- Allowed-Paths: 规则实现文件、规则样本目录、规则注册表文件
- Closed-Modules: 规则引擎、门禁判定、产物解析、阶段

- AC-N009-1: 以示例规则实现一次扩展，提交只涉及规则实现文件、规则样本目录和规则注册表文件。
  - AC-N009-1/N1: 示例规则的扩展提交通过 FR-034 开闭校验
  - AC-N009-1/E1: 示例规则的提交改动规则引擎模块时，开闭校验失败（回归测试）

## 5. 不做的事

- 不用大模型判断需求与设计的语义是否匹配；只校验显式声明的引用关系。
- 不替代各语言的编译器和测试框架。
- 不做需求管理 Web 界面。
- 不做大模型驱动的定期文档清理代理和代码质量打分（由 FR-030、FR-031、FR-032 的确定性检查替代）。
- 不做代理成本上限控制。

## 6. 待澄清项

### OQ-001 是否需要 Windows 支持
- Status: Closed
- Resolution: 支持 Windows（见 NFR-002）。

### OQ-002 第一版是否需要支持 Claude Code 以外的代理（Codex、Cursor 等）
- Status: Closed
- Resolution: 第一版支持 Claude Code、Codex、Cursor（见 AC-012-5）。

### OQ-003 接口的唯一来源
- Status: Closed
- Resolution: 两者都要：设计手写签名为意图，代码提取快照为事实，不一致即失败（见 AC-009-4、FR-018）。

接口签名是手写在架构声明里（与代码可能漂移），还是从代码或契约文件（OpenAPI、proto、TS 声明）提取？

### OQ-004 PRD 规模
- Status: Closed
- Resolution: 允许按领域拆分，PRD.md 为索引（见 AC-003-4）。

大项目是单个 PRD.md，还是允许按领域拆成多个文件（PRD 为索引）？

### OQ-005 "开发人员知情"的形式
- Status: Closed
- Resolution: 生成 CODEOWNERS，下游模块负责人成为接口文件 owner（见 AC-009-6）。

接口变更时，除了检查输出的通知清单，是否需要 CODEOWNERS 审批、PR 标签或其他通知渠道？

### OQ-006 提交粒度的执行位置
- Status: Closed
- Resolution: 本地 git hook + CI 兜底（见 AC-013-3）。

FR-013 通过本地 git hook、CI，还是两者都要？

### OQ-007 分发方式
- Status: Closed
- Resolution: 二进制内嵌全部模板，GitHub Releases + Homebrew tap（见 NFR-002）。

二进制通过 GitHub Releases、Homebrew，还是 `go install`？skill 和 hook 是否打包为 Claude Code 插件？

## 7. 澄清记录

每轮一个标题（日期 + 轮次），每轮最多 5 个问题（AC-002-1）。

### 2026-11-25 第 1 轮
- Q: 检查工具用什么语言实现？ A: Go 单二进制。
- Q: 需求变更怎么记录？ A: 直接改 PRD，并在变更日志登记；锁文件比对哈希。
- Q: 第一版支持哪些项目语言？ A: TypeScript/JavaScript、Python、Go、Java/Kotlin、C/C++。
- Q: 编码阶段拦截强度？ A: 硬拦截，无逃生开关。

### 2026-11-25 第 2 轮
- Q: 接口签名的唯一来源（OQ-003）？ A: 设计手写和代码提取两者都要。
- Q: 大项目 PRD 怎么组织（OQ-004）？ A: 允许按领域拆分，PRD.md 为索引。
- Q: 平台和代理范围（OQ-001、OQ-002）？ A: macOS/Linux/Windows；Claude Code、Codex、Cursor。
- Q: 分发方式（OQ-007）？ A: 二进制内嵌全部模板。

### 2026-11-25 第 3 轮
- Q: 接口变更的知情方式（OQ-005）？ A: 生成 CODEOWNERS。
- Q: 提交粒度检查位置（OQ-006）？ A: 本地 git hook + CI。

### 2026-11-25 第 4 轮
- 用户补充：增加测试 harness，集成测试覆盖所有需求的正常与边界情况，单元测试覆盖对外接口和核心逻辑，尽量用规则保证。
- Q: 补充哪些维度？ A: 会话交接、安全边界、依赖引入闸门、运行与观察应用、度量与自测、独立评审。
- Q: 自举 PRD 的验收用例何时补？ A: 先定范围和分期，再按期补。

### 2026-11-25 第 5 轮
- Q: 文档清理和质量评级？ A: 改为文档引用与过期检查、技术债登记、模块健康报告，不做大模型清理代理和打分。

### 2026-11-25 第 6 轮
- Q: 可扩展性要不要硬检测？ A: 要：扩展场景 + 扩展点 + 开闭校验、模块封装、依赖稳定性与扇出。

### 2026-11-25 第 7 轮
- Q: 各期还有哪些升为硬约束？ A: 人工审批防伪造与失效、阶段门禁、任务完成定义、需求文本质量、任务重叠与粒度、架构变更引用 ADR、流程与依赖一致、配置与阈值锁定、抑制注释登记、双向依赖一致、WCAG 对比度、CI 必须运行检查。

### 2026-11-25 第 8 轮
- 用户批准 v0.8 需求内容和三期分期方案。用例补齐后需重新批准（FR-037）。

### 2026-11-25 第 9 轮
- Q: 会话结束检查反复失败如何防死循环？ A: 连续阻止 3 次后放行，未通过项写入进度文件。
- Q: 是否 git init 并初始提交？ A: 是。

### 2026-11-25 第 10 轮
- 用户批准 PRD 1.0-rc1。`harness approve` 实现后由用户在终端补录批准哈希。

### 2026-11-25 第 11 轮
- 用户补充：架构需增加两个扩展点：新增流程阶段、在阶段中新增硬检测规则。
- Q: 由谁扩展？ A: harness 开发者（内置）和使用项目（不重新编译）都要。
- Q: 外部命令规则无法保证确定性怎么办？ A: 允许，但必须登记入锁文件并提供违规/合规样本自检。
- Q: 架构产物格式？ A: Markdown 字段格式 + JSON 配置，零第三方依赖。
- Q: Bash 写入判定？ A: 事前 shell 解析 + 事后快照比对回滚。
### 2026-11-26 第 12 轮
- 用户批准 PRD 1.0-rc2 与架构设计（总览、格式规范、14 个模块详细设计、调用流程、ADR-001 ~ ADR-007）。`harness approve` 实现后由用户在终端补录批准哈希。
### 2026-11-26 第 13 轮
- Q: 规划时发现，AC-039-1 要求任务完成时关联 FR 的全部用例通过，AC-011-6 又要求跨模块任务引用一个调用流程；跨多个流程的 FR 无法拆成多个任务。怎么处理？ A: 任务可关联到验收标准，完成定义只检查所关联 FR 或 AC 下的用例；FR 的每条 AC 都被关联也算覆盖。
- Q: go.mod、发布配置和 `harness init` 写入的配置文件不在任何模块路径内，归到哪里？ A: go.mod、go.sum、发布配置归 MOD-01；`harness init` 写入的文件定义为生成文件，提交范围检查与产物文件一样豁免。
### 2026-11-26 第 14 轮
- 用户批准 PRD 1.0-rc3、架构修订（ADR-008、MOD-01 路径、FLOW-01 步骤）与 Release 1 任务列表，进入实现阶段。`harness approve` 实现后由用户在终端补录批准哈希。

## 8. 变更日志

| ID | 日期 | 动作 | 对象 | 说明 |
|---|---|---|---|---|
| CHG-001 | 2026-11-25 | 新增 | FR-001..FR-017, NFR-001..NFR-005 | 初始草稿 |
| CHG-002 | 2026-11-25 | 新增 | FR-018 | 接口快照提取（OQ-003） |
| CHG-002 | 2026-11-25 | 修改 | FR-003, FR-009, FR-012, NFR-002 | PRD 可拆分、接口双来源、多代理门禁、三平台与内嵌分发（OQ-001..004, OQ-007） |
| CHG-003 | 2026-11-25 | 修改 | FR-009, FR-013 | CODEOWNERS 生成、git hook + CI（OQ-005, OQ-006） |
| CHG-004 | 2026-11-25 | 新增 | FR-019, FR-020, FR-021, FR-022, FR-023 | 测试 harness：验收用例、集成/单元测试追溯、测试有效性、变异测试 |
| CHG-004 | 2026-11-25 | 修改 | FR-016 | 新增测试 skill |
| CHG-005 | 2026-11-25 | 新增 | FR-024, FR-025, FR-026, FR-027, FR-028, FR-029 | 会话交接、安全边界、依赖闸门、运行与观察、自测与度量、独立评审 |
| CHG-005 | 2026-11-25 | 修改 | NFR-001 | 允许 `harness audit deps` 联网 |
| CHG-006 | 2026-11-25 | 新增 | FR-030, FR-031, FR-032 | 文档引用与过期检查、技术债登记、模块健康报告，替代文档清理代理和质量评级 |
| CHG-007 | 2026-11-25 | 新增 | FR-033, FR-034, FR-035, FR-036, NFR-006, NFR-007 | 可扩展性硬检测（含 harness 自身的两个扩展场景）：扩展场景与扩展点、开闭校验、模块封装、依赖稳定性与扇出 |
| CHG-008 | 2026-11-25 | 新增 | FR-037, FR-038, FR-039, FR-040, FR-041, FR-042, FR-043 | 硬约束补充：人工审批、阶段门禁、完成定义、需求文本质量、配置锁定、抑制注释、CI 集成 |
| CHG-008 | 2026-11-25 | 修改 | FR-006, FR-008, FR-010, FR-011, FR-013, FR-015, FR-025, NFR-001..NFR-007 | NFR 增加 Type 字段；流程与依赖一致、双向依赖、架构变更引用 ADR、任务重叠失败与粒度上限、提交行数上限、WCAG 对比度 |
| CHG-009 | 2026-11-25 | 新增 | FR-044 | 多代理门禁拆为独立需求（Release 3） |
| CHG-009 | 2026-11-25 | 修改 | FR-003, FR-012, 全部需求 | 增加 Release 字段与 Current-Release；AC-012-5 改为适配器接入 |
| CHG-010 | 2026-11-25 | 修改 | Release 1 全部需求、NFR-006、NFR-007 | 补齐 Release 1 验收用例（FR-019）；扩展场景补充 Adds/Allowed-Paths/Closed-Modules 字段；AC-002-1 澄清记录按轮次组织；AC-012-4 连续阻止 3 次后放行 |
| CHG-011 | 2026-11-25 | 新增 | FR-045, FR-046, NFR-008, NFR-009 | 流程阶段定义、项目自定义规则；新增阶段与新增规则两个扩展场景 |
| CHG-011 | 2026-11-25 | 修改 | TERM-017 | 阶段改为可定义、可扩展 |
| CHG-012 | 2026-11-26 | 新增 | TERM-026 | 生成文件 |
| CHG-012 | 2026-11-26 | 修改 | FR-011, FR-013, FR-039 | 任务可关联验收标准，完成定义按关联范围检查；提交范围检查豁免生成文件（第 13 轮） |
| CHG-013 | 2026-11-26 | 修改 | FR-040 | 词表与示例改为代码格式，避免 prd.09 命中需求自身引用的词表；语义不变 |
| CHG-014 | 2026-11-26 | 修改 | FR-005, FR-015 | 用例中引用的禁用别名改为代码格式；AC-015-2 的 `组件文件` 改为 `界面代码文件`，避免与禁用别名冲突；语义不变 |
| CHG-015 | 2026-11-26 | 修改 | NFR-006, NFR-007 | 补上 CHG-010 遗漏的 Adds、Allowed-Paths、Closed-Modules 字段，取值摘自正文；语义不变 |
