package app

import (
	"fmt"
	"path/filepath"

	"harness/internal/lock"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// ViolationsError 表示命令因检查未通过而没有执行。
type ViolationsError struct{ Report engine.Report }

func (e *ViolationsError) Error() string {
	return fmt.Sprintf("检查未通过（%d 条违规）", len(e.Report.Violations))
}

// Lock 在全量检查通过后重算哈希写入锁文件；保留批准记录与阶段（IF-023）。
// 检查未通过时返回 *ViolationsError，锁文件不变。
func Lock(root string) error {
	rep, err := Check(root, CheckOptions{})
	if err != nil {
		return err
	}
	if len(rep.Violations) > 0 {
		return &ViolationsError{Report: rep}
	}
	s, err := workspace.Load(root, workspace.LoadOptions{})
	if err != nil {
		return err
	}
	l := s.Lock
	l.Items = s.LockItems()
	return l.Write(filepath.Join(root, lock.File))
}
