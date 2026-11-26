# Java/Kotlin 约定

harness 按以下约定识别测试与注释。

## 测试

- 测试文件：`*Test`、`*Tests`、`*IT`
- 集成测试：`*IT`（Maven failsafe 约定），或位于 `integrationTest` 等含 integration 的目录
- 测试名引用验收用例：把用例编号中的 `-` 与 `/` 换成 `_`，例如 `void init_AC_001_1_N1()` 引用 `AC-001-1/N1`。

## 测试报告

Maven surefire/failsafe 或 Gradle 的 JUnit XML 报告复制到 `.harness/reports/`。

## 注释

TODO、FIXME、HACK、XXX 注释必须在同一行引用 Open 的 DEBT-n 或未完成的 T-n，例如 `// TODO(DEBT-1): 重试`。
