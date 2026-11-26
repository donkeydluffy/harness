package app

import (
	"fmt"
	"path/filepath"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/phase"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// PhaseList 返回全部阶段与当前阶段名。
func PhaseList(root string) ([]phase.Phase, string, error) {
	registerRules()
	s, err := workspace.Load(root, workspace.LoadOptions{})
	if err != nil {
		return nil, "", err
	}
	phases, err := phase.Load(s)
	if err != nil {
		return nil, "", err
	}
	cur, err := phase.Current(s, phases)
	return phases, cur.Name, err
}

// PhaseNext 由人推进阶段：确认交互式终端，当前阶段退出规则与下一阶段进入规则通过后请求确认，
// 把新阶段写入锁文件（IF-024）。检查未通过时返回 *ViolationsError；已是最后阶段时返回 phase.ErrLastPhase。
func PhaseNext(root string, t Terminal) error {
	if !t.IsInteractive() {
		return ErrNotInteractive
	}
	registerRules()
	s, err := workspace.Load(root, workspace.LoadOptions{WithSources: true, WithTests: true})
	if err != nil {
		return err
	}
	phases, err := phase.Load(s)
	if err != nil {
		return err
	}
	cur, err := phase.Current(s, phases)
	if err != nil {
		return err
	}
	next, rep, err := phase.Next(s, phases)
	if err != nil {
		return err
	}
	if len(rep.Errors) > 0 {
		return rep.Errors[0]
	}
	if len(rep.Violations) > 0 {
		return &ViolationsError{Report: rep}
	}
	if !t.Confirm(fmt.Sprintf("从 %s 推进到 %s？输入 y 确认：", cur.Name, next.Name)) {
		return ErrDeclined
	}
	s.Lock.Phase = next.Name
	return s.Lock.Write(filepath.Join(root, lock.File))
}
