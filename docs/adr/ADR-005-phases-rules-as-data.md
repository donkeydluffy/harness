# ADR-005 阶段与规则可注册、可配置

- Status: Superseded
- Superseded-By: ADR-010
- Relates: FR-046, NFR-009, MOD-04, MOD-05, MOD-10, EXT-03, EXT-04

## 背景

用户要求两个扩展点：新增流程阶段、在阶段中新增硬检测规则；扩展主体既包括 harness 开发者，也包括使用项目。

## 决策

- 规则实现统一的 Rule 接口（IF-011），内置规则每条一个文件，在 `registry.go` 中各占一行；项目规则（Release 2）由 `harness.json` 构造为同一接口。
- 阶段是数据：内置阶段为 `internal/phase/builtin/*.json`，项目阶段写在 `harness.json` 的 `phases`，合并后按顺序排列；阶段只引用规则编号。
- 规则编号格式 `<区域>.<两位序号>`，项目规则加前缀 `x.`。

## 后果

- 新增内置规则或阶段只新增文件和注册行，引擎、门禁、解析器保持封闭（由 EXT-03、EXT-04 的开闭校验保证）。
- 少数规则需要上层模块的能力（`phase.01`、`cfg.02`），由所属模块自行注册，避免 builtin 依赖上层。
