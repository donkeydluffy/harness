# 调用流程

- Status: Approved

每个流程的 Steps 按调用顺序列出"调用方模块 -> 被调用接口"（格式见 `docs/architecture/formats.md` 第 4 节）。

### FLOW-01 全量检查
- Covers: FR-003, FR-007, FR-011, FR-017
- Steps: MOD-01 -> IF-022, MOD-02 -> IF-009, MOD-09 -> IF-010, MOD-09 -> IF-004, MOD-09 -> IF-001, MOD-09 -> IF-002, MOD-09 -> IF-006, MOD-09 -> IF-007, MOD-09 -> IF-008, MOD-02 -> IF-014, MOD-02 -> IF-012, MOD-05 -> IF-003

cli 调用 app.Check；app 加载快照（配置、语言注册表、产物、锁文件、git 仓库、测试报告），注册内置规则后运行引擎；规则通过语言接口提取注释和测试名；cli 按格式输出并设置退出码。

### FLOW-02 写操作门禁
- Covers: FR-012, FR-025, FR-037, FR-045
- Steps: MOD-01 -> IF-019, MOD-01 -> IF-025, MOD-02 -> IF-009, MOD-02 -> IF-017, MOD-11 -> IF-015, MOD-11 -> IF-016, MOD-01 -> IF-019

claudecode 适配器解码 PreToolUse 输入为门禁请求；app 用轻量快照（不读测试报告和源码）调用 gate.Decide；gate 查询当前阶段的可写路径；适配器把判定编码为 deny 输出。

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
- Covers: FR-024
- Steps: MOD-01 -> IF-019, MOD-01 -> IF-025, MOD-02 -> IF-017, MOD-11 -> IF-016, MOD-11 -> IF-012, MOD-01 -> IF-019

### FLOW-07 初始化
- Covers: FR-001, FR-013, FR-016, FR-043
- Steps: MOD-01 -> IF-026, MOD-02 -> IF-020, MOD-13 -> IF-004, MOD-13 -> IF-019, MOD-13 -> IF-015

### FLOW-08 人工批准与阶段推进
- Covers: FR-037, FR-038, FR-045
- Steps: MOD-01 -> IF-024, MOD-02 -> IF-009, MOD-02 -> IF-012, MOD-02 -> IF-016, MOD-02 -> IF-005, MOD-02 -> IF-006

app 先确认交互式终端，再运行检查；approve 记录产物哈希与批准人，phase next 校验退出与进入规则后写入新阶段。

### FLOW-09 提交检查
- Covers: FR-010, FR-013, FR-034
- Steps: MOD-01 -> IF-022, MOD-02 -> IF-009, MOD-09 -> IF-007, MOD-02 -> IF-012

git 的 commit-msg hook 调用 `harness check commit adr.03 --msg-file <信息文件>`，检查暂存区与即将形成的提交信息；CI 调用 `harness check commits <范围>`，检查范围内每个提交。快照中的 Repo 提供提交与改动，规则 `commit.*`、`adr.03` 判定。pre-commit 时提交信息尚未写出，不运行依赖 Task 尾注的规则。只改产物与锁文件的需求提交可用 Change 尾注代替 Task 尾注（AC-013-5），此时依赖任务的 `commit.02`、`commit.04` 不适用。

### FLOW-10 锁定
- Covers: FR-004, FR-041
- Steps: MOD-01 -> IF-023, MOD-02 -> IF-009, MOD-02 -> IF-012, MOD-02 -> IF-005, MOD-02 -> IF-006
