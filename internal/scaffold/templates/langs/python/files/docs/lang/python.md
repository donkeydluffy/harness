# Python 约定

harness 按以下约定识别测试与注释。

## 测试

- 测试文件：`test_*.py`、`*_test.py`
- 集成测试：位于 `integration` 目录（如 `tests/integration/`）或文件名含 `integration`
- 测试名引用验收用例：把用例编号中的 `-` 与 `/` 换成 `_`，例如 `def test_init_AC_001_1_N1()` 引用 `AC-001-1/N1`。

## 测试报告

`pytest --junitxml=.harness/reports/pytest.xml`

## 注释

TODO、FIXME、HACK、XXX 注释必须在同一行引用 Open 的 DEBT-n 或未完成的 T-n，例如 `# TODO(DEBT-n): 重试`。
