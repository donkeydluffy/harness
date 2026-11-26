# MOD-11 gate 详细设计

- Status: Draft

## 接口

### IF-017 Decide
- Module: MOD-11
- Covers: FR-012, FR-024, FR-025, FR-037, FR-047, FR-048

```go
// IF-017
type Kind string // read / write / command / session_start / post_write / stop
type Request struct {
	Kind      Kind
	Root      string   // 项目根（绝对路径），由适配器从代理执行工具时的目录向上查找；快照、会话状态与判定都以它为根
	Cwd       string   // 代理执行工具时的目录（绝对路径）；command 中的相对路径从这里开始解析
	Paths     []string // 已解析为仓库相对路径
	Content   string   // write 时的新内容或补丁
	Command   string   // command 时的原始命令
	SessionID string
}
type Decision struct {
	Allow    bool
	Rule     string // 拒绝时的门禁规则编号，如 gate.write
	Reason   string
	Feedback string // 反馈给代理的文字（检查结果、会话注入内容）
}
func Decide(s *workspace.Snapshot, req Request) Decision
```

### IF-018 Snapshot
- Module: MOD-11
- Covers: FR-012

```go
// IF-018
// Before 在命令执行前记录仓库文件的 (路径, 大小, mtime) 与小文件内容副本，存于 .harness/snap/<会话>。
func Before(root, sessionID string) error
// After 比对命令前后状态；不被 allowed 允许的改动恢复为命令前内容，新建文件删除。
func After(root, sessionID string, allowed func(path string) bool) ([]Change, error)
```

## 判定顺序

1. `gate.protect`：写入锁文件、`harness.json`、hook 配置，或产物状态行改为 Approved，或执行 `harness approve` → 拒绝。
2. `gate.sensitive`：读写敏感路径 → 拒绝（`sensitiveAllow` 除外）。
3. `gate.secret`：写入内容命中密钥规则 → 拒绝。
4. `gate.command`：危险命令 → 拒绝；命令中解析出的写入目标按第 5、6 步判定。
5. `gate.write`：源码写入须有 InProgress 任务、路径在其范围内，且该任务的批准链（IF-028 Chain）全部有效；失效时拒绝并列出失效单位与原因（AC-047-4）。
6. 放行。`docs/` 下的产物任何时候都可写（AC-048-2），Release 1 的 `gate.artifact` 删除。

## Bash 命令解析

用 `mvdan.cc/sh/v3/syntax` 解析命令（ADR-003）：重定向目标、`tee`、`cp`、`mv`、`rm`、`touch`、`mkdir`、`sed -i`、`git checkout -- <path>`、`git restore` 的路径参数作为写入目标。无法静态判定的命令（脚本、解释器）放行到执行，由 IF-018 事后比对。

## 结束检查

`session_start` 注入 IF-028 Summarize 的批准状态摘要与建议的 skill（AC-048-3），取代当前阶段。

`stop` 请求运行全量检查；同一会话连续拒绝次数存于 `.harness/session/<会话>.json`，达到 `stopMaxBlocks` 后放行，并把未通过项写入 `docs/progress.md` 的"未完成"段。
