# ADR-006 门禁判定与代理协议分离

- Status: Proposed
- Relates: FR-012, FR-044, NFR-007, MOD-11, MOD-12, EXT-02

## 背景

Claude Code、Codex、Cursor 的 hook 事件名、输入字段、拒绝方式各不相同（`docs/research/2026-11-25-landscape.md` 第 4 节），Release 3 才接入后两者。

## 决策

门禁判定（MOD-11）只处理与代理无关的 Request / Decision；每种代理一个适配器（MOD-12），负责事件解码、结果编码、hook 配置模板。cli 调用适配器，app 只接收 Request。

## 后果

- 新增代理只新增适配器目录与注册行，门禁判定保持封闭。
- 适配器解码失败一律判定为拒绝（失败即阻断）。
