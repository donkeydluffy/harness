package lang

import "harness/internal/lang/golang"

// 内置语言，每种语言一行。
func init() {
	Register(golang.Lang{})
}
