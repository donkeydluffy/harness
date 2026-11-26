package app

import "github.com/donkeydluffy/harness/internal/scaffold"

// Init 在项目根目录生成骨架（IF-026）。
//
// harness:IF-026
func Init(root string, opt scaffold.InitOptions) (scaffold.Result, error) {
	return scaffold.Generate(root, opt)
}
