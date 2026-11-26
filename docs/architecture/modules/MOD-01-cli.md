# MOD-01 cli 详细设计

- Status: Approved

## 命令

| 命令 | 调用 |
|---|---|
| `harness init [--lang go,...] [--ci github|gitlab] [--ci-version <版本>]` | IF-037 取版本 → IF-026 |
| `harness check [规则或前缀...] [--commits <范围>] [--msg-file <文件>] [--json]`；`check commits <范围>` 与 `--commits` 等价 | IF-022 |
| `harness trace [--format md|json]` | IF-022 |
| `harness lock` | IF-023 |
| `harness check lint` / `check unit` / `check smoke` | IF-038：先运行外部工具写报告，再运行同名区域的规则 |
| `harness snapshot` | IF-039 |
| `harness approve <产物路径、prd 或单位编号>` | IF-024 |
| `harness version`、`harness --version` | IF-037，输出 `harness <版本>`（AC-049-1） |
| `harness phase ...` | 已删除：输出"阶段已取消，改由批准状态判定（FR-048）"，退出码 1 |
| `harness hook <代理> <事件>` | IF-019 解码 → IF-025 → IF-019 编码 |

退出码：0 通过，1 有违规，2 内部错误（AC-N004-2）；只有告警时为 0。`check` 把全部子命令名传入 IF-022 的 `CheckOptions.Commands`，供文档引用检查（AC-030-2）。用法提示与 `harness help` 列出 version（AC-049-2）。本模块不对外提供接口。
