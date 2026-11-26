# ADR-002 产物用 Markdown 字段格式，配置用 JSON

- Status: Accepted
- Relates: NFR-005, FR-003, FR-006, MOD-03, MOD-09

## 背景

产物既要人直接编辑、git diff 可读，又要被确定性解析。可选：全部 Markdown、Markdown + YAML、JSON。

## 决策

所有产物用 Markdown，机器读取"编号标题 + `- Key: value` 字段"结构（`docs/architecture/formats.md`）。项目配置 `harness.json` 与锁文件 `harness.lock` 用 JSON，由 Go 标准库解析。

## 后果

- 一个解析器覆盖 PRD、架构、ADR、任务、技术债；零第三方依赖。
- JSON 配置不能写注释，说明写在 `docs/architecture/formats.md`。
- 字段格式的语法错误由 MOD-03 以 Diagnostic 报告，规则 `prd.01`、`arch.01` 等给出修复建议。
