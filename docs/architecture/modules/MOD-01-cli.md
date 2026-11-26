# MOD-01 cli 详细设计

- Status: Approved

## 命令

| 命令 | 调用 |
|---|---|
| `harness init [--lang go,...] [--ci github|gitlab]` | IF-026 |
| `harness check [规则或前缀...] [--commits <范围>] [--msg-file <文件>] [--json]`；`check commits <范围>` 与 `--commits` 等价 | IF-022 |
| `harness trace [--format md|json]` | IF-022 |
| `harness lock` | IF-023 |
| `harness approve <产物路径>` | IF-024 |
| `harness phase list|next` | IF-024 |
| `harness hook <代理> <事件>` | IF-019 解码 → IF-025 → IF-019 编码 |

退出码：0 通过，1 有违规，2 内部错误（AC-N004-2）。本模块不对外提供接口。
