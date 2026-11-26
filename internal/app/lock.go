package app

import (
	"fmt"
	"path/filepath"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// ViolationsError 表示命令因检查未通过而没有执行。
type ViolationsError struct{ Report engine.Report }

func (e *ViolationsError) Error() string {
	return fmt.Sprintf("检查未通过（%d 条违规）", len(e.Report.Violations))
}

// Lock 在产物文档检查通过后重算哈希写入锁文件；保留批准记录与阶段（IF-023）。
// 需要源码、测试报告或 git 的规则不参与：需求阶段就要能锁定 PRD。
// 检查未通过时返回 *ViolationsError，锁文件不变。
func Lock(root string) error {
	registerRules()
	rep, err := run(root, docRules(), workspace.LoadOptions{})
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
