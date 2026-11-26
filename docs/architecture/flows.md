# 调用流程

- Status: Draft

每个流程的 Steps 按调用顺序列出"调用方模块 -> 被调用接口"（格式见 `docs/architecture/formats.md` 第 4 节）。

调用 IF-015、IF-016 的步骤属于 Release 1，在 MOD-10 删除（FR-048）时一并去掉。

### FLOW-01 全量检查
- Covers: FR-003, FR-007, FR-011, FR-017, FR-030, FR-035, FR-042, FR-053, FR-054
- Steps: MOD-01 -> IF-022, MOD-02 -> IF-009, MOD-09 -> IF-010, MOD-09 -> IF-004, MOD-09 -> IF-001, MOD-09 -> IF-002, MOD-09 -> IF-006, MOD-09 -> IF-007, MOD-09 -> IF-008, MOD-09 -> IF-033, MOD-02 -> IF-014, MOD-02 -> IF-036, MOD-02 -> IF-013, MOD-02 -> IF-012, MOD-05 -> IF-003, MOD-05 -> IF-035, MOD-05 -> IF-028

cli 调用 app.Check；app 加载快照（配置、语言注册表、产物、锁文件、git 仓库、测试报告），注册内置规则后运行引擎；规则通过语言接口提取注释和测试名；cli 按格式输出并设置退出码。

Release 2：快照另读 lint 与 unit 报告；app 构造项目自定义规则并先用样本验证；依赖、封装、接口锚点类规则经 IF-035 解析 Go 源码，批准类规则经 IF-028 取批准状态。cli 把子命令名放入 CheckOptions，供文档引用检查。

### FLOW-02 写操作门禁
- Covers: FR-012, FR-025, FR-037, FR-047, FR-048
- Steps: MOD-01 -> IF-019, MOD-01 -> IF-025, MOD-02 -> IF-009, MOD-02 -> IF-017, MOD-11 -> IF-015, MOD-11 -> IF-016, MOD-11 -> IF-028, MOD-01 -> IF-019

claudecode 适配器解码 PreToolUse 输入为门禁请求；app 用轻量快照（不读测试报告和源码）调用 gate.Decide；产物写入不再按阶段判定；源码写入时 gate 查询 InProgress 任务的批准链（IF-028 Chain）；适配器把判定编码为 deny 输出。

### FLOW-03 命令执行后比对
- Covers: FR-012
- Steps: MOD-01 -> IF-019, MOD-01 -> IF-025, MOD-02 -> IF-017, MOD-11 -> IF-018, MOD-01 -> IF-019

PreToolUse(Bash) 放行时 gate 调用 Before 记录状态；PostToolUse(Bash) 时调用 After，恢复范围外改动并把恢复清单反馈给代理。

### FLOW-04 写后产物检查
- Covers: FR-012
- Steps: MOD-01 -> IF-019, MOD-01 -> IF-025, MOD-02 -> IF-017, MOD-11 -> IF-012, MOD-01 -> IF-019

PostToolUse 写入产物文件后，gate 按文件类型选择规则前缀（如 `prd.`）运行引擎，结果作为反馈返回代理。

### FLOW-05 会话结束检查
- Covers: FR-012, FR-024
- Steps: MOD-01 -> IF-019, MOD-01 -> IF-025, MOD-02 -> IF-009, MOD-02 -> IF-017, MOD-11 -> IF-012, MOD-01 -> IF-019

### FLOW-06 会话开始注入
- Covers: FR-024, FR-048
- Steps: MOD-01 -> IF-019, MOD-01 -> IF-025, MOD-02 -> IF-017, MOD-11 -> IF-016, MOD-11 -> IF-028, MOD-11 -> IF-012, MOD-01 -> IF-019

注入批准状态摘要（IF-028 Summarize）与建议的 skill，取代当前阶段。

### FLOW-07 初始化
- Covers: FR-001, FR-013, FR-014, FR-016, FR-043, FR-051
- Steps: MOD-01 -> IF-037, MOD-01 -> IF-026, MOD-02 -> IF-020, MOD-13 -> IF-004, MOD-13 -> IF-034, MOD-13 -> IF-019, MOD-13 -> IF-015, MOD-13 -> IF-037

cli 取当前版本放入 InitOptions；scaffold 按语言配置包生成配置与缺失的工具配置文件，CI 固定版本不是发布版本号时拒绝。

### FLOW-08 人工批准
- Covers: FR-037, FR-047, FR-054
- Steps: MOD-01 -> IF-024, MOD-02 -> IF-009, MOD-02 -> IF-029, MOD-02 -> IF-036, MOD-02 -> IF-012, MOD-02 -> IF-016, MOD-02 -> IF-005, MOD-02 -> IF-006

app 先确认交互式终端，列出范围内待批准的单位与差异，运行批准前检查（单位所在文件的结构规则、上游已批准、登记在该单位类型上的自定义规则），人确认后写入按单位的批准记录。Release 1 的 phase next 随 MOD-10 删除。

### FLOW-09 提交检查
- Covers: FR-010, FR-013, FR-034, FR-050
- Steps: MOD-01 -> IF-022, MOD-02 -> IF-009, MOD-09 -> IF-007, MOD-02 -> IF-012, MOD-05 -> IF-035

git 的 commit-msg hook 调用 `harness check commit adr.03 --msg-file <信息文件>`，检查暂存区与即将形成的提交信息；CI 调用 `harness check commits <范围>`，检查范围内每个提交。快照中的 Repo 提供提交与改动，规则 `commit.*`、`adr.03` 判定。pre-commit 时提交信息尚未写出，不运行依赖 Task 尾注的规则。只改产物与锁文件的需求提交可用 Change 尾注代替 Task 尾注（AC-013-5），此时依赖任务的 `commit.02`、`commit.04` 不适用。机械改动任务的提交由 `commit.05` 经 IF-035 逐行判定只改了 import 路径与模块声明行（AC-050-3）。

### FLOW-10 锁定
- Covers: FR-004, FR-041, FR-047, FR-048
- Steps: MOD-01 -> IF-023, MOD-02 -> IF-009, MOD-02 -> IF-012, MOD-02 -> IF-029, MOD-02 -> IF-005, MOD-02 -> IF-006

Release 2：锁文件为 Release 1 格式时经 IF-029 Migrate 迁移为按单位批准，并删除 phase 字段。

### FLOW-11 外部工具报告
- Covers: FR-014, FR-021, FR-055
- Steps: MOD-01 -> IF-038, MOD-02 -> IF-009, MOD-02 -> IF-030, MOD-02 -> IF-031, MOD-02 -> IF-032, MOD-16 -> IF-034, MOD-16 -> IF-035, MOD-16 -> IF-033, MOD-02 -> IF-012

`harness check lint|unit|smoke`：app 按 kind 调用 toolrun；toolrun 按语言配置包（或项目覆盖）运行工具、Go 用 IF-035 采集覆盖率，写报告；app 重新加载快照，运行同名区域的规则。

### FLOW-12 接口快照
- Covers: FR-052
- Steps: MOD-01 -> IF-039, MOD-02 -> IF-009, MOD-02 -> IF-035

`harness snapshot`：app 对 sourcePaths 内的 Go 文件提取锚点声明，按接口编号排序写入 `docs/architecture/interfaces.txt`。
