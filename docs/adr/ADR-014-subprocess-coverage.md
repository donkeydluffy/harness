# ADR-014 模块覆盖率计入集成测试与子进程

- Status: Proposed
- Relates: FR-021, IF-035, IF-031

## 背景

AC-021-3 按模块统计行覆盖率。最初的设计只全量运行单元测试（跳过名称含 `_AC_` 的集成测试），本仓库的 cli、app、gate 主要由集成测试覆盖：集成测试构建 harness 二进制并在子进程中运行。只算单元测试时 cli 为 0%、app 为 8.2%、gate 为 39.6%，与实际的测试程度不符，阈值失去意义。

`go test -cover` 会把测试进程的 `GOCOVERDIR` 改为自己的临时目录，子进程写入的覆盖率数据不会合并进 `-coverprofile`。

## 决策

- 统计模块覆盖率的全量运行包含集成测试。按接口运行的测试仍只取测试名含接口编号的测试。
- Go Analyzer 全量运行时设置环境变量 `HARNESS_COVERDIR` 为一个临时目录。测试构建被测二进制时，若该变量不为空，以 `go build -cover -coverpkg=<模块路径>/...` 构建，并把子进程的 `GOCOVERDIR` 设为该目录。运行结束后用 `go tool covdata textfmt` 转换，与 `-coverprofile` 合并。
- 被否决的方案：调低阈值到只算单元测试时的数值。cli、app 的阈值会变成 0 与 5，不再约束什么。
- 被否决的方案：为 cli、app、gate 补单元测试达到原阈值。要几千行与集成测试重复的测试。

## 后果

- `harness check unit` 的全量运行耗时与一次完整的 `go test ./...` 相当。
- 用户项目的测试以子进程运行自己的程序时，按同样的约定使用 `HARNESS_COVERDIR`，子进程的覆盖率才会计入；不使用时只统计测试进程内执行的代码。
