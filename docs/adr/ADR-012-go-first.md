# ADR-012 Release 2 的代码分析只做 Go；CI 下载固定版本

- Status: Accepted
- Relates: FR-050, FR-051, FR-052, FR-053, FR-035, MOD-01, MOD-13, MOD-14, MOD-18, EXT-01

## 背景

依赖约束、接口签名、机械改动校验都要解析源码。每种语言的 import 与声明解析是一套独立的工作，而 harness 自身是 Go 项目，只有 Go 能在自举中得到真实验证。另外，用户项目的 CI 用 `go install @latest` 安装 harness，版本会漂移，CI 里还得装 Go。

## 决策

- MOD-14 的 Analyzer 是语言的可选接口，Release 2 只有 Go 实现（`go/parser`、`go test -coverprofile`）；其他语言没有 Analyzer 时，相关规则跳过并在输出中说明。
- 接口签名快照提交到 `docs/architecture/interfaces.txt`，签名变化在评审 diff 中可见；规则 if.02 保证快照与源码一致。被否决的方案：只在检查时临时提取。锁文件比对仍然可行，但人在 diff 中看不到签名变化。
- 用户项目的 CI 下载生成时版本的发布归档，并用 checksums.txt 校验。harness 没有发布版本号时，init 拒绝生成。被否决的方案：`go install module@<版本>`，需要 Go 工具链，而且从源码构建，结果与发布包不完全相同。

## 后果

- NFR-006（新增语言）改到 Release 3；新增语言时实现 Packer 与 Analyzer，规则不改。
- 升级 harness 后需要重新运行 `harness init`，更新 CI 中固定的版本；cfg.02 会提示。
- cli 依赖 MOD-18 取当前版本，用于 `harness version` 与 init；EXT-01 的扩展路径加入 `internal/scaffold/templates/langs/<语言>/**`，新增语言时连同工具配置文件一起提供。
