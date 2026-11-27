# ADR-013 Analyzer.Coverage 同时返回可执行行

- Status: Proposed
- Relates: FR-021, IF-035, IF-031

## 背景

AC-021-3 要求按模块统计行覆盖率并与 Coverage 阈值比较。IF-035 的 `Coverage` 只返回已执行的行，算不出分母：模块里哪些行含有语句，只有覆盖率文件知道（执行次数为 0 的语句块也在其中）。

## 决策

- `Coverage` 增加返回值 `statements`：覆盖率文件中全部语句块覆盖的行（文件 -> 行号），与 `executed` 同一格式。模块行覆盖率 = 模块路径内 `executed` 的行数 / `statements` 的行数 × 100。
- 被否决的方案：用模块源文件的非空、非注释行作分母。import、类型声明与右括号都会算作未覆盖，结果明显偏低，阈值失去意义。
- 被否决的方案：toolrun 自己解析覆盖率文件。覆盖率文件的格式属于具体语言，应留在 Analyzer 内。

## 后果

- IF-035 的签名变化，T-048 已实现的 Go Analyzer 随 T-049 修改；`harness snapshot` 更新接口快照，if.04 由本 ADR 说明这次变化。
- 一个模块没有任何语句行时（只有类型声明）不计算覆盖率，unit.03 不检查它的阈值。
