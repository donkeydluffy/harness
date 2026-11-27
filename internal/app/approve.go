package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Terminal 是与批准人交互的终端（IF-024）。
type Terminal interface {
	IsInteractive() bool
	Confirm(prompt string) bool
	Identity() string
}

// ErrNotInteractive 表示不在交互式终端中；批准与阶段推进必须由人执行。
var ErrNotInteractive = errors.New("必须由人在交互式终端中执行，标准输入不是终端")

// ErrDeclined 表示批准人在确认提示处拒绝。
var ErrDeclined = errors.New("未确认，产物状态与锁文件均未改变")

// now 便于测试固定时间。
var now = func() time.Time { return time.Now().UTC().Truncate(time.Second) }

// Approve 由人批准（IF-024）：按批准单位批准（approveUnits）。target 可以是产物路径、prd（docs/PRD.md）或单位编号。
// 锁文件仍有 Release 1 的按文件批准时先由 harness lock 迁移；检查未通过时返回 *ViolationsError。
//
// harness:IF-024
func Approve(root, target string, t Terminal) error {
	if !t.IsInteractive() {
		return ErrNotInteractive
	}
	registerRules()
	s, err := workspace.Load(root, workspace.LoadOptions{})
	if err != nil {
		return err
	}
	if !s.Lock.ByUnit() {
		if len(s.Lock.Legacy) > 0 {
			return fmt.Errorf("锁文件是 Release 1 的按文件批准格式，先运行 harness lock 迁移")
		}
		s.Lock.Version, s.Lock.Units = lock.UnitVersion, map[string]lock.UnitApproval{}
	}
	return approveUnits(root, s, target, t)
}
