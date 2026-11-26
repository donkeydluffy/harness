// Package adapter 把各编程代理的 hook 输入输出转换为门禁请求与判定（MOD-12）。
// 新增代理是扩展点 EXT-02：实现 Adapter 并在 registry.go 登记一行。
package adapter

import (
	"sort"
	"sync"

	"harness/internal/gate"
)

// Adapter 是一种编程代理的 hook 适配器（IF-019）。
type Adapter interface {
	Name() string
	// Decode 把代理 hook 的事件名和标准输入转换为门禁请求。
	Decode(event string, stdin []byte) (gate.Request, error)
	// Encode 把门禁判定转换为代理期望的标准输出和退出码。
	Encode(event string, d gate.Decision) (stdout []byte, exit int)
	// HookConfig 返回写入项目的 hook 配置文件路径与内容。
	HookConfig(binary string) (path string, content []byte)
}

var (
	mu       sync.RWMutex
	adapters = map[string]Adapter{}
)

// Register 注册适配器；名称重复时 panic（IF-019）。
func Register(a Adapter) {
	mu.Lock()
	defer mu.Unlock()
	if _, dup := adapters[a.Name()]; dup {
		panic("adapter: 适配器重复注册 " + a.Name())
	}
	adapters[a.Name()] = a
}

// Get 按名称返回适配器（IF-019）。
func Get(name string) (Adapter, bool) {
	mu.RLock()
	defer mu.RUnlock()
	a, ok := adapters[name]
	return a, ok
}

// Names 返回已注册的适配器名称，按字母排序。
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	var out []string
	for n := range adapters {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
