# ADR-004 锁文件记录哈希、批准与当前阶段

- Status: Superseded
- Superseded-By: ADR-009
- Relates: FR-004, FR-037, FR-041, MOD-06

## 背景

需求变更检测、批准有效性、门禁配置锁定、当前阶段都需要"人确认过的状态"，且必须防止代理篡改。

## 决策

统一放在仓库根目录的 `harness.lock`（JSON，键排序）。代理不能直接编辑锁文件（`gate.protect`）；`harness lock` 只在检查通过时更新哈希；批准与阶段推进只能由交互式终端的 `harness approve` 和 `harness phase next` 写入。

## 后果

- 一个受保护文件承载全部"受信状态"，门禁只需保护它与 `harness.json`、hook 配置。
- 多人并行修改时锁文件可能冲突；冲突解决后重新运行 `harness lock`。
