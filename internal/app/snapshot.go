package app

import (
	"os"
	"path/filepath"

	"github.com/donkeydluffy/harness/internal/workspace"
)

// InterfacesFile 是接口快照文件的路径。
const InterfacesFile = workspace.InterfacesFile

// Snapshot 从源码锚点提取接口签名（IF-035），写入接口快照文件 docs/architecture/interfaces.txt（AC-052-2）。
//
// harness:IF-039
func Snapshot(root string) error {
	s, err := workspace.Load(root, workspace.LoadOptions{WithSources: true})
	if err != nil {
		return err
	}
	p := filepath.Join(root, filepath.FromSlash(workspace.InterfacesFile))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, workspace.FormatInterfaces(s.ExtractInterfaces()), 0o644)
}
