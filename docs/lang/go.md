# Go 约定

harness 按以下约定识别测试与注释。

## 测试

- 测试文件：`*_test.go`
- 集成测试：文件名以 `_integration_test.go` 结尾，或带 `//go:build integration` 约束
- 测试名引用验收用例：把用例编号中的 `-` 与 `/` 换成 `_`，例如 `TestInit_AC_001_1_N1` 引用 `AC-001-1/N1`。

## 测试报告

`go test -json ./... > .harness/reports/go.jsonl`

## 注释

TODO、FIXME、HACK、XXX 注释必须在同一行引用 Open 的 DEBT-n 或未完成的 T-n，例如 `// TODO(DEBT-n): 重试`。
