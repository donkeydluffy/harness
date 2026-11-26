// Package app 是用例编排层：check、trace、lock、approve、phase next、门禁、init（MOD-02）。
package app

import (
	"sync"

	"harness/internal/rules/builtin"
	"harness/internal/rules/engine"
)

var registerOnce sync.Once

// registerRules 把内置规则注册到引擎，进程内只执行一次。
func registerRules() {
	registerOnce.Do(func() {
		for _, r := range builtin.Rules() {
			engine.Register(r)
		}
	})
}
