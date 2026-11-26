package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"harness/internal/lock"
	"harness/internal/model"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
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

// Approve 由人批准产物：确认交互式终端，产物文档规则通过后请求确认，
// 把文件头 Status 改为 Approved（ADR 为 Accepted），并在锁文件追加批准记录（IF-024）。
// artifact 可以是产物路径，或 prd（docs/PRD.md）。检查未通过时返回 *ViolationsError。
func Approve(root, artifact string, t Terminal) error {
	if !t.IsInteractive() {
		return ErrNotInteractive
	}
	registerRules()
	rel := filepath.ToSlash(filepath.Clean(artifact))
	if strings.EqualFold(artifact, "prd") {
		rel = workspace.PRDIndex
	}
	if !workspace.Approvable(rel) {
		return fmt.Errorf("%s 不是可批准的产物（可批准：prd 或 docs/PRD.md、docs/architecture/*.md、docs/architecture/modules/*.md、docs/adr/*.md）", artifact)
	}
	s, err := workspace.Load(root, workspace.LoadOptions{})
	if err != nil {
		return err
	}
	docs := s.ApprovalDocs(rel)
	if len(docs) == 0 {
		return fmt.Errorf("%s 不存在", rel)
	}
	if rep := artifactViolations(s, docs); len(rep.Violations) > 0 {
		return &ViolationsError{Report: rep}
	}
	hash := s.ApprovalHash(rel)
	if !t.Confirm(fmt.Sprintf("批准 %s（%s）？输入 y 确认：", rel, hash)) {
		return ErrDeclined
	}
	if err := setStatus(filepath.Join(root, filepath.FromSlash(rel)), workspace.ApprovedStatus(rel)); err != nil {
		return err
	}
	l := s.Lock
	l.Approvals = append(l.Approvals, lock.Approval{Artifact: rel, Hash: hash, By: t.Identity(), At: now()})
	return l.Write(filepath.Join(root, lock.File))
}

// artifactViolations 运行产物文档规则（批准有效性规则除外），只保留落在被批准文件中的违规。
func artifactViolations(s *workspace.Snapshot, docs []*model.Doc) engine.Report {
	files := map[string]bool{}
	for _, d := range docs {
		files[d.Path] = true
	}
	sel := docRules()
	var ids []string
	for _, id := range sel.IDs {
		if !strings.HasPrefix(id, "approve.") {
			ids = append(ids, id)
		}
	}
	rep := engine.Run(s, engine.Selector{IDs: ids})
	var keep []engine.Violation
	for _, v := range rep.Violations {
		if files[v.File] {
			keep = append(keep, v)
		}
	}
	rep.Violations = keep
	return rep
}

// setStatus 把文件头的 Status 行改为 status；没有文件头 Status 时在标题后插入。
func setStatus(path, status string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	if i := workspace.HeaderStatusLine(lines); i >= 0 {
		lines[i] = "- Status: " + status
	} else {
		lines = append(lines[:1], append([]string{"", "- Status: " + status}, lines[1:]...)...)
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}
