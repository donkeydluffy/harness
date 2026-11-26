package lang

import (
	"harness/internal/lang/cpp"
	"harness/internal/lang/golang"
	"harness/internal/lang/java"
	"harness/internal/lang/python"
	"harness/internal/lang/typescript"
)

// 内置语言，每种语言一行。
func init() {
	Register(golang.Lang{})
	Register(typescript.Lang{})
	Register(python.Lang{})
	Register(java.Lang{})
	Register(cpp.Lang{})
}
