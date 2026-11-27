# ADR-017 阶段模块删除后的依赖与扩展点

- Status: Proposed
- Relates: FR-048, MOD-01, MOD-02, MOD-11, EXT-02, EXT-03, EXT-04

## 背景

ADR-009 决定以按单位批准取代阶段。Release 2 生效后（Current-Release 为 2），阶段模块 MOD-10 与它的代码目录一起删除（T-054、T-055、T-061）。架构声明中依赖 MOD-10、把它列为封闭模块的元素需要随之调整。

## 决策

- MOD-10 与扩展点 EXT-04（新增内置阶段）的 Status 改为 Removed，保留原有字段作为历史记录（AC-007-5）。
- MOD-01、MOD-02、MOD-11 的 Depends 去掉 MOD-10；MOD-02 同时去掉 MOD-03：去掉第 1 版的按文件批准后，app 不再直接使用 parser（deps.03）。
- EXT-02、EXT-03 的 Closed 去掉 MOD-10。
- 被否决的方案：保留 MOD-10 作为空模块，等以后需要时再恢复阶段。阶段已被批准链完全取代，空模块只会让依赖图与覆盖检查多一个没有代码的元素。

## 后果

- 依赖图少一个节点，仍然无环；门禁、cli 不再有按阶段判定的分支。
- 读到 Release 1 锁文件的 `phase` 字段时忽略，`harness lock` 写出后消失（AC-048-1）。
