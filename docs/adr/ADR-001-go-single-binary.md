# ADR-001 检查工具用 Go 单二进制实现

- Status: Proposed
- Relates: NFR-001, NFR-002, MOD-01

## 背景

检查工具要在每个使用 harness 的项目、三种操作系统、CI 中直接运行，且门禁单次耗时 < 200ms（NFR-003）。

## 决策

用 Go 实现，编译为单个静态二进制；模板通过 `go:embed` 打包。除 ADR-003 外不引入第三方依赖。

## 后果

- 分发只需一个文件；不要求目标机器安装 Python 或 Node。
- 需要为 6 个平台交叉编译与发布（AC-N002-1）。
- 不引入 YAML 与 Markdown 库，解析器自行实现（见 ADR-002）。
