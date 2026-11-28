# 技术债清单

- Status: Draft
- Updated: 2026-11-28

每项格式见 `docs/architecture/formats.md` 第 3 节：`### DEBT-nnn 标题`，字段 Status（Open / Resolved）、Module。

## 技术债

### DEBT-001 Go 模块路径暂为 harness
- Status: Resolved
- Module: MOD-01

仓库还没有远程地址，`go.mod` 的模块路径暂用 `harness`。T-027 发布前改为仓库地址，否则 `go install` 不可用。CI 模板的安装步骤同样使用这个路径（`internal/scaffold/scaffold.go` 的 modulePath），需一并修改。

已解决（fd3898e）：模块路径改为 `github.com/donkeydluffy/harness`，modulePath 与本仓库 CI 配置一并修改。

### DEBT-002 解析诊断未被任何规则报告
- Status: Resolved
- Module: MOD-05

`workspace.Snapshot.Diags`（代码块未闭合、用例不在验收标准下、用例编号重复）目前没有规则输出，问题会被静默忽略。T-003 起 prd.07 报告用例相关诊断，T-005 起 arch.01 报告架构文档的诊断。adr.01（T-007）、debt.01（T-009）、task.01（T-011）已报告各自文件的诊断。仍未报告的：PRD 中与用例无关的诊断（如代码块未闭合、实体编号重复），以及 `docs/progress.md` 的诊断。

已解决（c780404，T-032）：`diagOwner` 统一判定每条诊断由哪条规则报告，其余诊断（PRD 中未结束的代码块、跨文件重复的编号、进度文件等）由 doc.01 报告；测试确认每条诊断恰好被一条规则报告且归属正确。

### DEBT-003 task.08 自行维护全量检查的排除列表
- Status: Resolved
- Module: MOD-05

`engine.FullCheck` 定义了"全量检查"不含的覆盖类规则（test.02、arch.05），会话结束检查与会话开始注入使用它；task.08 仍在 task_08.go 中重复列出这两条规则。T-018 的模块范围不含 MOD-05，没有一并修改。以后修改 task.08 的任务应改为调用 `engine.FullCheck("task.08")`，避免两处列表不一致。

已解决（c780404，T-032）：task.08 改为调用 `engine.FullCheck("task.08")`。

### DEBT-004 找不到 harness 二进制时 hook 失败放行
- Status: Resolved
- Module: MOD-12

`harness init` 生成的 hook 命令是裸 `harness`，依赖 Claude Code 进程的 PATH。T-026 启用后，PATH 里没有 `~/go/bin` 的那段时间，hook 以退出码 127 失败，Claude Code 把它当作非阻止错误，写操作没有经过门禁。应让 PreToolUse 在找不到二进制时拒绝写操作（例如 hook 命令外包一层 `command -v harness || 输出拒绝`），并在 SessionStart 提示安装。

已解决（T-030）：hook 命令先用 `command -v` 确认 harness 可执行；找不到时 PreToolUse 拒绝、SessionStart 注入说明，PostToolUse 与 Stop 不阻止。

### DEBT-005 harness init 不忽略门禁运行时目录
- Status: Resolved
- Module: MOD-13

门禁在 `.harness/snap/`、`.harness/session/` 下保存快照和会话计数，属于运行时状态，不应提交；`harness init` 没有把它们加进 `.gitignore`，本仓库目前只能靠提交时不加这两个目录。应由 init 写入（或追加）`.gitignore` 条目，并把 `.gitignore` 的这部分计入生成文件。

已解决（T-033）：init 写入 `.harness/.gitignore`（`snap/`、`session/` 与它自身），不改项目根目录的 `.gitignore`，不需要提交；已有项目重新运行 `harness init` 即补上。

### DEBT-006 secret.entropy 把含大写与数字的长路径误判为密钥
- Status: Resolved
- Module: MOD-11

`internal/gate/security.go` 的 `tokenRe` 包含 `/`，整条绝对路径（如 `/private/tmp/claude-501/-Users-…-Desktop-…/<uuid>/…`）成为一个 token，同时含大小写字母和数字，熵超过 4.3，命令被拒绝。T-027 期间读取 scratchpad 文件的命令因此被拦。修复方向：以 `/`、`./`、`~/` 开头的 token 按路径处理，或按 `/` 分段后再判断熵；需补一条路径不误报、base64 密钥仍命中的用例。

已解决（388f331，T-028）：以 `/`、`./`、`~/` 开头的 token 按路径处理；含多个 `/` 的 base64 密钥仍命中。

### DEBT-007 跨全部模块的机械改动没有合规路径
- Status: Resolved
- Module: MOD-05

task.05 要求文件范围跨模块的任务引用覆盖这些模块的调用流程，task.03 要求范围在关联模块内，commit.03 限制单个提交 400 行。模块路径改名（DEBT-001）这类只改 import 的机械改动跨全部 14 个模块、约 400 行，不对应任何调用流程。T-027 期间由人在终端运行脚本并以 `--no-verify` 提交（fd3898e）。Release 2 应给机械改动一个受约束的通道，例如任务声明 `Kind: mechanical` 时免于 task.05，但提交只能改动 import 路径，由规则校验。

已解决（T-051，FR-050）：任务声明 `Kind: mechanical` 与 `Rewrite: 旧前缀 -> 新前缀` 并引用 ADR 后，免于文件范围、调用流程与提交行数三项约束；提交只能把 import 路径与模块声明行中的旧前缀替换为新前缀（AC-050-3）。

### DEBT-008 失败的非用例测试不被任何规则报告
- Status: Resolved
- Module: MOD-05

test.02 只看用例编号的测试，arch.05 只看扩展点编号的测试；名字含 IF 编号或不含编号的测试失败时，`harness check` 与 task.08 都通过。T-028 提交时 `TestParseRepoArtifacts_IF_001` 失败而检查全部通过。应增加一条规则：测试报告中存在失败的测试时报告（或把它纳入完成定义），并给出失败的测试名。

已解决（c780404，T-032）：test.03 报告测试报告中每个失败的测试，计入全量检查，task.08 与会话结束检查随之失败。

### DEBT-009 门禁不允许代理按 cfg.02 的修复办法重新生成生成文件
- Status: Resolved
- Module: MOD-11

commit.02/commit.03 豁免生成文件（TERM-026），门禁却按源码路径判定 `.github/workflows/harness.yml` 这类不属于任何模块的生成文件，拒绝 `rm` 后再 `harness init`。cfg.02 的修复办法因此只能由人执行（T-027 模板变化后即如此）。应让门禁允许删除生成文件，再由单独的 `harness init` 写回，写回内容由 cfg.02 校验。

已解决（df67d88，T-031）：整条命令只是 `rm` CI 配置（`workspace.RegeneratedFiles`）时放行并保留删除，单独一次 `harness init` 写回的内容保留。hook 配置 `.claude/settings.json` 不在其中，仍只能由人重新生成（AC-025-4）。

### DEBT-010 用户项目的 CI 安装 @latest，与生成文件的 harness 版本可能不一致
- Status: Resolved
- Module: MOD-13

生成的 CI 配置用 `go install …/cmd/harness@latest`，而生成文件来自本地那个版本的模板。harness 发布了改动模板的新版本后，用户项目的 CI 会因 cfg.02 失败，本地却通过；代理缓存还会让 `@latest` 落后于最新提交（T-027 中本仓库的 CI 装到了旧提交）。本仓库已改为从源码构建（`SelfHosted`）；用户项目应把版本固定为生成时的 harness 版本（取自构建信息），升级 harness 后由 cfg.02 提示重新生成。

已解决（T-047，FR-051）：生成的 CI 配置下载生成时所用 harness 版本的发布包并校验 checksums.txt；本地 harness 与 CI 固定的版本不同时，cfg.02 提示重新运行 `harness init`。

### DEBT-011 测试包构建失败不被任何规则报告
- Status: Resolved
- Module: MOD-08

`go test -json` 的包级结果（如测试包编译失败）在解析时被跳过，只保留测试级结果。包编译失败时其中的测试不出现在报告里：引用用例的测试由 test.02 报告为缺失，单元测试则无人报告。应把包级 fail 记为一条失败结果，由 test.03 报告。

已解决（T-034）：包级结果为 fail 而包内没有失败的测试时，解析结果增加一条名为 `testreport.PackageFailure` 的失败记录，由 test.03 报告。

### DEBT-012 门禁按仓库根目录解析命令中的相对路径，不跟随 cd
- Status: Resolved
- Module: MOD-11

`cd internal/x && mkdir -p y` 中的 `y` 被当作仓库根目录下的 `y` 判定。T-032 写规则样本时因此被误拒（改用仓库相对路径后通过）。反方向同样成立：`cd` 到范围外目录后写相对路径，命令前判定会按根目录放行，只能靠命令后比对恢复。应在解析命令时跟踪同一条命令里字面量的 `cd`，无法判定时交给命令后比对。

已解决（T-035）：门禁跟随顺序执行（`;`、`&&`）的顶层字面量 `cd`，之后的相对路径按新目录解析；子 shell、`||`、管道中的 `cd` 以及 `cd -`、无参数或含变量的目标使之后的相对路径无法判定，写入交给命令后比对，递归删除直接拒绝。

### DEBT-013 secret.entropy 把长的驼峰标识符判为密钥
- Status: Resolved
- Module: MOD-11

长度不少于 32、大小写混合且含数字的标识符（如由 Test、Parse、Go、JSON、Package、Failure 与用例编号拼成的测试名）熵超过 4.3，写入测试文件的命令被拒绝；T-034 中把测试名改短才通过，登记本条时原样引用该测试名也被拒绝。真实密钥很少由可读单词拼成：可考虑对按大小写边界能切成多个常见单词的 token 降权，或只对引号内、等号右侧的值计算熵。需补一条长标识符不误报、随机 token 仍命中的用例。

已解决（T-036）：计算熵之前排除由英文单词拼成的标识符（按大小写、数字与 `_` 切段后，短段与数字段少、像单词的段至少 3 个、非缩写段元音占比不低于 30%）。用随机 base62 token 估算，约百万分之十几会被放过；单元测试把上限定为万分之 0.5。

### DEBT-014 secret.entropy 把较长的仓库相对路径判为密钥
- Status: Open
- Module: MOD-11

T-028 只豁免以 `/`、`./`、`~/` 开头的路径。`docs/architecture/modules/` 加上大小写混合、含数字的模块文件名，去掉扩展名后长度超过 32，命令被拒绝（2026-11-26 修改架构时遇到，改用通配符绕过）。可考虑：token 含 `/` 且每一段都是标识符形状（T-036 的单词判定）或常见文件名时不按熵判定；随机密钥中的 `/` 两侧多为随机碎片，仍按熵判定。需补长相对路径不误报、含 `/` 的 base64 密钥仍命中的用例。

另一种形式：`=` 也在 token 字符集中，`S=/private/tmp/...` 这样的赋值使 token 以变量名开头而不被视为路径（2026-11-26 预演 Release 2 时遇到，给路径加引号绕过）。

### DEBT-015 shell 停在子目录时门禁把子目录当作仓库根
- Status: Resolved
- Module: MOD-01

`harness gate` 以进程当前目录为根（`internal/cli/cli.go` 的 `os.Getwd()`），没有使用适配器从 hook 输入的 cwd 向上查找到的项目根。代理的 shell 停在 `docs/architecture` 时，命令前快照建在 docs/architecture 下的 .harness/snap/，判定只加载该目录（没有任务、锁文件），命令后比对也只覆盖该目录（2026-11-26 实际发生）。后果是保护可被绕过：在子目录执行 `echo x > ../harness.lock`，路径被判为仓库外而放行，命令后比对也看不到它。修复：门禁使用项目根判定，并补"在子目录执行写保护文件的命令被拒绝"的回归测试。

已解决（T-039）：适配器在请求中填写项目根与代理的当前目录；`harness gate` 以项目根加载快照、保存命令前快照与会话状态，命令中的相对路径从代理的当前目录解析。回归测试覆盖子目录中经 `../` 写锁文件被拒绝、相对路径按子目录判定范围、快照建在项目根、命令后比对恢复整个仓库的改动。

### DEBT-016 cp 目标含变量时门禁把来源当作写入目标
- Status: Open
- Module: MOD-11

`writeTargets` 取 `cp`、`install`、`ln` 的最后一个非空操作数作为目标，而含变量的词被解析为空串后跳过。`cp -R /abs/docs /abs/harness.json "$S"/` 的目标是 `"$S"/`，门禁却把来源 `harness.json` 当作目标，按保护文件拒绝（2026-11-26 遇到）。方向是安全的，但判定错误。修复：最后一个操作数无法静态判定时，目标视为无法判定，交给命令后比对；补"目标含变量时不把来源当目标"的用例。


### DEBT-017 先改设计的接口签名无法批准
- Status: Resolved
- Module: MOD-02

`harness approve IF-n` 对待批准单位运行 if.03，要求设计签名与代码一致；而代码要等任务的批准链有效后才能修改（gate.write、task.07）。先改设计时两边互相等待，只能先回退设计、改代码、再恢复设计（2026-11-27 改 IF-035 时遇到）。修复方向：if.03 比对的是下游的代码，像其他要求下游覆盖的规则一样不参与批准前检查，只由全量检查要求一致；补"设计先改的接口能批准"的用例。
同类问题：InProgress 任务的批准因上游变化失效后，`harness approve T-n` 被 task.07 拒绝（task.07 报的正是这次批准要解决的失效），只能先把任务改回 Todo 再批准。task.07 也不应参与该任务自身的批准前检查。

已解决（T-059）：批准前检查不再运行 if.03 与 task.07，签名与代码的一致、任务批准的有效仍由全量检查要求。IF-024 的单元测试覆盖签名先改的接口与 InProgress 任务的重新批准。

### DEBT-018 已存在的 Git 钩子无法随 harness 升级更新
- Status: Open
- Module: MOD-13

`harness init` 不覆盖内容不同的已有文件，`.git/hooks/pre-commit` 也按此跳过；harness 升级后钩子中的规则列表（如 T-058 去掉的 phase）不会更新，只能删掉钩子再运行 init（2026-11-27 遇到）。钩子由 harness 生成、不由人编辑，可考虑让 init 识别自己生成的旧版钩子并替换，或让钩子只调用一个不带规则列表的子命令。

### DEBT-019 门禁把含 docs/PRD.md 的子目录当作项目根
- Status: Open
- Module: MOD-11

在 `internal/scaffold/templates/artifacts/docs` 下执行命令时，门禁向上找到 `templates/artifacts/docs/PRD.md`，把 `templates/artifacts` 当作项目根，在其中写入 `.harness/snap/<会话>/`；这些文件随后被嵌入模板，init 的集成测试失败（2026-11-27 遇到）。项目根应以 harness.json 或 git 仓库根判定，不以 docs/PRD.md 判定。

### DEBT-020 清除"未完成"段时连带删掉其后的进度记录
- Status: Open
- Module: MOD-11

门禁在会话结束检查连续阻止 3 次后把未通过项写入 docs/progress.md 末尾的"## 未完成"段；检查通过后清除该段，但段的结束只认下一个 `## ` 标题，而进度记录是 `### ` 标题，追加在该段之后的记录被一并删除（2026-11-27 T-059 的记录因此丢失，此前 T-049 的记录也丢过一次）。修复方向：把"未完成"段的正文限定为门禁自己写入的内容（如以结束标记收尾），或写入时放在文件开头、追加记录时插到它之前；补"段后追加的记录在清除后保留"的用例。

### DEBT-021 命令后比对把人同时写入的锁文件恢复掉
- Status: Open
- Module: MOD-11

门禁在代理命令前给受保护文件做快照，命令后发现内容变化就恢复为快照。人在自己的终端执行 `harness approve` 时，若代理的长命令（如 `harness check unit`，约 80 秒）正在运行，命令结束时 harness.lock 被判为代理改动并恢复，人的批准记录全部丢失，且没有任何提示给人（2026-11-27 一次批准 62 个单位后遇到）。修复方向：命令后比对只恢复由该命令进程写入的改动，或对受保护文件的变化先比对写入者（如锁文件中批准记录的身份与时间晚于命令开始时间则保留）；恢复前把被覆盖的内容另存，便于找回。
