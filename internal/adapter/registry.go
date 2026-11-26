package adapter

import "github.com/donkeydluffy/harness/internal/adapter/claudecode"

// 内置适配器，每个代理一行。
func init() {
	Register(claudecode.Adapter{})
}
