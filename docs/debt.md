# 技术债清单

- Status: Draft
- Updated: 2026-11-26

每项格式见 `docs/architecture/formats.md` 第 3 节：`### DEBT-nnn 标题`，字段 Status（Open / Resolved）、Module。

## 技术债

### DEBT-001 Go 模块路径暂为 harness
- Status: Resolved
- Module: MOD-01

仓库还没有远程地址，`go.mod` 的模块路径暂用 `harness`。T-027 发布前改为仓库地址，否则 `go install` 不可用。CI 模板的安装步骤同样使用这个路径（`internal/scaffold/scaffold.go` 的 modulePath），需一并修改。

已解决（fd3898e）：模块路径改为 `github.com/donkeydluffy/harness`，modulePath 与本仓库 CI 配置一并修改。

### DEBT-002 解析诊断未被任何规则报告
- Status: Open
- Module: MOD-05

`workspace.Snapshot.Diags`（代码块未闭合、用例不在验收标准下、用例编号重复）目前没有规则输出，问题会被静默忽略。T-003 起 prd.07 报告用例相关诊断，T-005 起 arch.01 报告架构文档的诊断。adr.01（T-007）、debt.01（T-009）、task.01（T-011）已报告各自文件的诊断。仍未报告的：PRD 中与用例无关的诊断（如代码块未闭合、实体编号重复），以及 `docs/progress.md` 的诊断。

### DEBT-003 task.08 自行维护全量检查的排除列表
- Status: Open
- Module: MOD-05

`engine.FullCheck` 定义了"全量检查"不含的覆盖类规则（test.02、arch.05），会话结束检查与会话开始注入使用它；task.08 仍在 task_08.go 中重复列出这两条规则。T-018 的模块范围不含 MOD-05，没有一并修改。以后修改 task.08 的任务应改为调用 `engine.FullCheck("task.08")`，避免两处列表不一致。

### DEBT-004 找不到 harness 二进制时 hook 失败放行
- Status: Open
- Module: MOD-12

`harness init` 生成的 hook 命令是裸 `harness`，依赖 Claude Code 进程的 PATH。T-026 启用后，PATH 里没有 `~/go/bin` 的那段时间，hook 以退出码 127 失败，Claude Code 把它当作非阻止错误，写操作没有经过门禁。应让 PreToolUse 在找不到二进制时拒绝写操作（例如 hook 命令外包一层 `command -v harness || 输出拒绝`），并在 SessionStart 提示安装。

### DEBT-005 harness init 不忽略门禁运行时目录
- Status: Open
- Module: MOD-13

门禁在 `.harness/snap/`、`.harness/session/` 下保存快照和会话计数，属于运行时状态，不应提交；`harness init` 没有把它们加进 `.gitignore`，本仓库目前只能靠提交时不加这两个目录。应由 init 写入（或追加）`.gitignore` 条目，并把 `.gitignore` 的这部分计入生成文件。

### DEBT-006 secret.entropy 把含大写与数字的长路径误判为密钥
- Status: Resolved
- Module: MOD-11

`internal/gate/security.go` 的 `tokenRe` 包含 `/`，整条绝对路径（如 `/private/tmp/claude-501/-Users-…-Desktop-…/<uuid>/…`）成为一个 token，同时含大小写字母和数字，熵超过 4.3，命令被拒绝。T-027 期间读取 scratchpad 文件的命令因此被拦。修复方向：以 `/`、`./`、`~/` 开头的 token 按路径处理，或按 `/` 分段后再判断熵；需补一条路径不误报、base64 密钥仍命中的用例。

已解决（388f331，T-028）：以 `/`、`./`、`~/` 开头的 token 按路径处理；含多个 `/` 的 base64 密钥仍命中。

### DEBT-007 跨全部模块的机械改动没有合规路径
- Status: Open
- Module: MOD-05

task.05 要求文件范围跨模块的任务引用覆盖这些模块的调用流程，task.03 要求范围在关联模块内，commit.03 限制单个提交 400 行。模块路径改名（DEBT-001）这类只改 import 的机械改动跨全部 14 个模块、约 400 行，不对应任何调用流程。T-027 期间由人在终端运行脚本并以 `--no-verify` 提交（fd3898e）。Release 2 应给机械改动一个受约束的通道，例如任务声明 `Kind: mechanical` 时免于 task.05，但提交只能改动 import 路径，由规则校验。

### DEBT-008 失败的非用例测试不被任何规则报告
- Status: Open
- Module: MOD-05

test.02 只看用例编号的测试，arch.05 只看扩展点编号的测试；名字含 IF 编号或不含编号的测试失败时，`harness check` 与 task.08 都通过。T-028 提交时 `TestParseRepoArtifacts_IF_001` 失败而检查全部通过。应增加一条规则：测试报告中存在失败的测试时报告（或把它纳入完成定义），并给出失败的测试名。


### DEBT-009 门禁不允许代理按 cfg.02 的修复办法重新生成生成文件
- Status: Open
- Module: MOD-11

commit.02/commit.03 豁免生成文件（TERM-026），门禁却按源码路径判定 `.github/workflows/harness.yml` 这类不属于任何模块的生成文件，拒绝 `rm` 后再 `harness init`。cfg.02 的修复办法因此只能由人执行（T-027 模板变化后即如此）。应让门禁允许删除生成文件，再由单独的 `harness init` 写回，写回内容由 cfg.02 校验。

### DEBT-010 用户项目的 CI 安装 @latest，与生成文件的 harness 版本可能不一致
- Status: Open
- Module: MOD-13

生成的 CI 配置用 `go install …/cmd/harness@latest`，而生成文件来自本地那个版本的模板。harness 发布了改动模板的新版本后，用户项目的 CI 会因 cfg.02 失败，本地却通过；代理缓存还会让 `@latest` 落后于最新提交（T-027 中本仓库的 CI 装到了旧提交）。本仓库已改为从源码构建（`SelfHosted`）；用户项目应把版本固定为生成时的 harness 版本（取自构建信息），升级 harness 后由 cfg.02 提示重新生成。
