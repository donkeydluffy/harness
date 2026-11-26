# MOD-12 adapter 详细设计

- Status: Approved

## 接口

### IF-019 Adapter
- Module: MOD-12
- Covers: FR-012, FR-044

```go
// IF-019
type Adapter interface {
	Name() string
	// Decode 把代理 hook 的事件名和标准输入转换为门禁请求。
	Decode(event string, stdin []byte) (gate.Request, error)
	// Encode 把门禁判定转换为代理期望的标准输出和退出码。
	Encode(event string, d gate.Decision) (stdout []byte, exit int)
	// HookConfig 返回写入项目的 hook 配置文件路径与内容。
	HookConfig(binary string) (path string, content []byte)
}
func Register(a Adapter)
func Get(name string) (Adapter, bool)
```

## claudecode 适配器

| 事件 | 请求 | 输出 |
|---|---|---|
| SessionStart | session_start | `additionalContext` 注入 |
| PreToolUse（Read/Edit/Write/MultiEdit/NotebookEdit/Bash） | read / write / command | 拒绝时 `permissionDecision: "deny"` |
| PostToolUse（Edit/Write/MultiEdit/Bash） | post_write | 反馈检查结果；Bash 后执行 IF-018 After |
| Stop | stop | 拒绝时 `decision: "block"` |

Decode 失败时判定为拒绝（AC-012-5/E1），hook 配置设置 `onFailure: "block"`。
