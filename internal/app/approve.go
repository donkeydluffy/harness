package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/donkeydluffy/harness/internal/gate"
	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// Terminal 是与批准人交互的终端（IF-024）。
type Terminal interface {
	IsInteractive() bool
	Confirm(prompt string) bool
	Notify(msg string)
	Identity() string
}

// ErrNotInteractive 表示不在交互式终端中；批准必须由人执行。
var ErrNotInteractive = errors.New("必须由人在交互式终端中执行，标准输入不是终端")

// ErrDeclined 表示批准人在确认提示处拒绝。
var ErrDeclined = errors.New("未确认，产物状态与锁文件均未改变")

// now 便于测试固定时间。
var now = func() time.Time { return time.Now().UTC().Truncate(time.Second) }

// staleAfter 之前开始的命令前快照视为代理异常退出后留下的残留，不再等待（AC-056-1/B1）。
const staleAfter = 10 * time.Minute

// pollEvery 是等待代理命令结束时的检查间隔。
var pollEvery = 500 * time.Millisecond

// errAgentStarted 表示人确认之后、写入锁文件之前有代理命令开始执行。
var errAgentStarted = errors.New("代理命令开始执行")

// Approve 由人批准（IF-024）：按批准单位批准（approveUnits）。target 可以是产物路径、prd（docs/PRD.md）或单位编号。
// 锁文件仍有 Release 1 的按文件批准时先由 harness lock 迁移；检查未通过时返回 *ViolationsError。
// 代理命令正在执行时先等待其结束，写入锁文件前再查一次，避免命令后比对把批准记录恢复掉（AC-056-1，ADR-018）。
//
// harness:IF-024
func Approve(root, target string, t Terminal) error {
	if !t.IsInteractive() {
		return ErrNotInteractive
	}
	registerRules()
	var accepted time.Time // 人已确认可以忽略的残留快照中最晚的时间
	for {
		if err := awaitAgent(root, t, &accepted); err != nil {
			return err
		}
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
		busy := func() bool { since, ok := gate.Running(root); return ok && since.After(accepted) }
		err = approveUnits(root, s, target, t, busy)
		if !errors.Is(err, errAgentStarted) {
			return err
		}
		t.Notify("确认之后有代理命令开始执行，锁文件未写入；等它结束后重新列出待批准单位。\n")
	}
}

// awaitAgent 等待正在执行的代理命令结束；命令前快照早于 staleAfter 时询问人是否继续，
// 确认后把它记入 accepted 不再等待，拒绝时返回 ErrDeclined。
func awaitAgent(root string, t Terminal, accepted *time.Time) error {
	notified := false
	for {
		since, ok := gate.Running(root)
		if !ok || !since.After(*accepted) {
			return nil
		}
		at := since.Local().Format("15:04:05")
		if time.Since(since) >= staleAfter {
			if !t.Confirm("代理命令的快照始于 " + at + "，可能是代理异常退出后留下的。仍要继续批准？输入 y 确认：") {
				return ErrDeclined
			}
			*accepted = since
			continue
		}
		if !notified {
			t.Notify("代理命令正在执行（始于 " + at + "），等它结束后继续批准……\n")
			notified = true
		}
		time.Sleep(pollEvery)
	}
}
