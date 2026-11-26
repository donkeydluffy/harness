# TypeScript/JavaScript 约定

harness 按以下约定识别测试与注释。

## 测试

- 测试文件：`*.test.ts`、`*.spec.js` 等
- 集成测试：文件名含 `integration`（如 `a.integration.test.ts`），或位于 `integration`、`e2e` 目录
- 测试名引用验收用例：把用例编号中的 `-` 与 `/` 换成 `_`，例如 `test('Init AC_001_1_N1', …)` 引用 `AC-001-1/N1`。

## 测试报告

用 JUnit 报告器（如 jest-junit、vitest `--reporter=junit`）输出到 `.harness/reports/*.xml`。

## 注释

TODO、FIXME、HACK、XXX 注释必须在同一行引用 Open 的 DEBT-n 或未完成的 T-n，例如 `// TODO(DEBT-1): 重试`。
