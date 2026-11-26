# MOD-05 builtin 详细设计

- Status: Approved

## 接口

### IF-014 Rules
- Module: MOD-05
- Covers: FR-003, FR-007, FR-011

```go
// IF-014
// Rules 返回全部内置规则；registry.go 中每条规则一行，供 app 在启动时注册。
func Rules() []engine.Rule
```

## 规则目录（Release 1）

每条规则一个文件 `internal/rules/builtin/<area>_<nn>.go`，样本在 `testdata/<规则编号>/{violation,ok}/`。

| 规则 | 检查 | 覆盖的验收标准 |
|---|---|---|
| prd.01 | 编号格式合法、全局唯一（含拆分文件） | AC-003-1, AC-003-4 |
| prd.02 | 需求必填字段与取值（Status/Release/Priority/Type），Current-Release 存在 | AC-003-2, AC-003-5, AC-040-2 |
| prd.03 | 锁文件中的需求与用例编号不得消失；Removed 编号不得复用 | AC-003-3, AC-019-3 |
| prd.04 | 功能需求、非功能需求、统一语言三章齐全且非空 | AC-002-3 |
| prd.05 | 有 Open 的待澄清项时 PRD 不得为 Approved；Closed 必须有 Resolution | AC-002-2 |
| prd.06 | 澄清记录按轮次标题组织、每轮不超过 5 问 | AC-002-1 |
| prd.07 | 用例编号格式、前缀一致，每条验收标准至少 N 加 B 或 E | AC-019-1, AC-019-2 |
| prd.08 | 哈希变化的需求必须有新 CHG，且动作与差异一致 | AC-004-1, AC-004-2 |
| prd.09 | 模糊词 | AC-040-1 |
| prd.10 | Quantitative 类 NFR 验收标准含数值与单位 | AC-040-2 |
| prd.11 | Scenario 类 NFR 字段齐全；无 Scenario 时须声明"不预留扩展" | AC-033-1, AC-033-5 |
| term.01 | 禁用别名 | AC-005-1 |
| term.02 | 标注为领域概念的名称在术语表中 | AC-005-2 |
| arch.01 | 模块、接口、流程、扩展点的必填字段；模块路径不重叠 | AC-006-1, AC-006-2, AC-006-3, AC-033-2 |
| arch.02 | 声明依赖无环 | AC-006-4 |
| arch.03 | 流程步骤与依赖声明一致 | AC-006-5 |
| arch.04 | 扩展点必须关联 Scenario 类 NFR | AC-033-3 |
| arch.05 | 扩展点有通过的示例扩展测试 | AC-033-4 |
| trace.01 | 纳入范围的 Active FR 被设计元素覆盖 | AC-007-1 |
| trace.02 | 设计元素不是孤儿 | AC-007-2 |
| trace.03 | 引用指向存在且有效的编号 | AC-007-3 |
| adr.01 | ADR 必填字段与正文章节 | AC-010-1 |
| adr.02 | Superseded 链有效且无环 | AC-010-2 |
| adr.03 | 架构声明的结构变化必须由同一提交中的 ADR 引用 | AC-010-3 |
| task.01 | 任务必填字段与取值 | AC-011-1 |
| task.02 | 纳入范围的 Active FR 被任务覆盖 | AC-011-2 |
| task.03 | 任务范围落在关联模块内 | AC-011-3 |
| task.04 | 至多一个 InProgress；未完成任务范围重叠须有前置关系 | AC-011-4 |
| task.05 | 单模块或引用覆盖所有涉及模块的流程 | AC-011-6 |
| task.06 | 扩展任务范围不含封闭模块 | AC-034-2 |
| task.07 | InProgress 任务的阶段前置条件（PRD、设计批准有效，无 Open 待澄清项） | AC-038-1, AC-038-2, AC-038-3, AC-011-5 |
| task.08 | Done 任务的完成定义 | AC-039-1, AC-039-2, AC-024-2 |
| commit.01 | Task 尾注唯一且有效 | AC-013-1 |
| commit.02 | 改动文件在任务范围内 | AC-013-2 |
| commit.03 | 提交行数上限 | AC-013-4 |
| commit.04 | 扩展任务提交不改封闭模块 | AC-034-1 |
| test.01 | 测试名引用的编号存在 | AC-020-1 |
| test.02 | 用例被通过的集成测试覆盖；skipped 不计；静态声明须出现在报告中 | AC-020-2, AC-020-3, AC-020-4 |
| debt.01 | 技术债必填字段 | AC-031-1 |
| debt.02 | TODO/FIXME/HACK/XXX 引用 Open 的 DEBT 或未完成的 T | AC-031-2 |
| debt.03 | Resolved 的 DEBT 不被引用 | AC-031-3 |
| cfg.01 | 门禁配置哈希变化须有 CHG 或 ADR | AC-041-1 |
| ci.01 | CI 配置包含 `harness check` | AC-043-2 |
| approve.01 | 产物批准有效性（哈希一致） | AC-037-3 |

另有两条规则由所属模块自行注册，以免 builtin 依赖上层模块：`phase.01`（阶段定义合法；移除内置规则须引用 ADR，AC-045-1、AC-045-2）在 MOD-10；`cfg.02`（生成文件与期望内容一致，AC-041-2、AC-017-2）在 MOD-13。

规则 `prd.03`、`prd.08`、`cfg.01` 需要锁文件；`adr.03`、`commit.*` 需要 git；`arch.05`、`test.*`、`task.08` 需要测试报告。快照缺少对应输入时，规则报告"输入缺失"违规而不是静默通过。
