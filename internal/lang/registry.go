package lang

import (
	"github.com/donkeydluffy/harness/internal/lang/cpp"
	"github.com/donkeydluffy/harness/internal/lang/golang"
	"github.com/donkeydluffy/harness/internal/lang/java"
	"github.com/donkeydluffy/harness/internal/lang/python"
	"github.com/donkeydluffy/harness/internal/lang/typescript"
)

// 内置语言，每种语言一行。
func init() {
	Register(golang.Lang{})
	Register(typescript.Lang{})
	Register(python.Lang{})
	Register(java.Lang{})
	Register(cpp.Lang{})
}
