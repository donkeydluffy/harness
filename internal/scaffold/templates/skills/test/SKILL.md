---
name: test
description: 编写测试：根据验收用例编写集成测试，根据接口和核心模块编写单元测试。
check: harness check test
---

# 编写测试

## 集成测试对应验收用例

PRD 中每个当前 Release 的用例都要有一个状态为 passed 的集成测试。测试名中写用例编号，`-` 与 `/` 换成 `_`：`AC-001-1/N1` 写作 `AC_001_1_N1`。一个测试可以引用多个用例。

| 语言 | 集成测试文件 | 测试名示例 |
|---|---|---|
| go | `internal/<模块>/order_integration_test.go` | `TestCreate_AC_001_1_N1` |
| typescript | `tests/integration/order.test.ts` | `create AC_001_1_N1` |
| python | `tests/integration/test_order.py` | `test_create_AC_001_1_N1` |
| java | `src/test/java/OrderIT.java` | `create_AC_001_1_N1` |
| cpp | `tests/integration/order_test.cc` | `Create_AC_001_1_N1` |

各语言如何区分集成测试与单元测试见 `docs/lang/<语言>.md`。只被单元测试引用的用例不算覆盖。

## 单元测试对应接口与核心模块

- 接口测试名写接口编号：`IF-n` 写作 `IF_n`，如 `TestEnqueue_IF_004`。
- 扩展点至少有一个示例扩展测试，测试名写扩展点编号，如 `TestExample_EXT_01`。

## 测试要能失败

- 每个 E、B 类用例的测试要断言具体的错误信息或边界结果，不只断言"没有崩溃"。
- 写完后临时破坏被测分支，确认测试会失败，再恢复。
- 测试名引用的编号必须存在且未 Removed，否则 `harness check test` 失败。

## 结束

生成测试报告到 `.harness/reports/`，运行 `harness check test`。
