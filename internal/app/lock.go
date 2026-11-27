package app

import (
	"fmt"
	"path/filepath"

	"github.com/donkeydluffy/harness/internal/approval"
	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// ViolationsError 表示命令因检查未通过而没有执行。
type ViolationsError struct{ Report engine.Report }

func (e *ViolationsError) Error() string {
	return fmt.Sprintf("检查未通过（%d 条违规）", len(e.Report.Violations))
}

// Lock 在产物文档检查通过后重算哈希写入锁文件；保留批准记录（IF-023）。
// 有按文件批准记录的第 1 版锁文件同时迁移为按单位批准（AC-047-6）；没有批准记录的第 1 版锁文件
// （包括新建的）在 Release 2 生效前保持第 1 版，第 1 版逻辑删除时（T-055）一并按第 2 版处理。
// 需要源码、测试报告或 git 的规则不参与：只有 PRD、还没有代码时就要能锁定。
// 检查未通过时返回 *ViolationsError，锁文件不变。
//
// harness:IF-023
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
	if !l.ByUnit() {
		migrate(s, l)
	}
	return l.Write(filepath.Join(root, lock.File))
}

// migrate 把第 1 版锁文件的按文件批准换算为按单位批准，有换算结果时写入审计记录 unit-migration（AC-047-6）。
func migrate(s *workspace.Snapshot, l *lock.Lock) {
	l.Version = lock.UnitVersion
	if len(l.Legacy) == 0 {
		l.Units = map[string]lock.UnitApproval{}
		return
	}
	units, raw := approval.Migrate(s)
	l.Units = units
	l.Legacy = []lock.Approval{}
	l.Audit["unit-migration"] = raw
}
