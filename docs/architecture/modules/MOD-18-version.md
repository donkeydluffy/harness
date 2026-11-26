# MOD-18 version 详细设计

- Status: Approved

## 接口

### IF-037 Current
- Module: MOD-18
- Covers: FR-049, FR-051

```go
// IF-037
// Current 返回当前二进制的版本（AC-049-1）：发布构建通过 -ldflags 注入的版本号；否则构建信息中的模块版本；
// 否则 vcs.revision 的前 12 位，vcs.modified 为 true 时加 -dirty；都取不到时为 dev。
func Current() string

// Release 判断版本是否为发布版本号：形如 0.2.0，不是 dev、提交哈希，也不含 -dirty（AC-051-2）。
func Release(v string) bool
```

## 内部设计

- 叶子模块，只用 `runtime/debug.ReadBuildInfo`。发布配置（`.goreleaser.yaml`，MOD-01）注入 `internal/version.injected`。
- `go install module@v0.2.0` 构建的二进制，模块版本为 `v0.2.0`，输出去掉前缀 `v`。
