package workspace

import (
	"path"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/model"
)

// Approvable 判断产物是否需要人工批准：PRD 索引（覆盖全部 PRD 文件）、架构文档、详细设计、ADR（FR-037）。
func Approvable(rel string) bool {
	dir := path.Dir(rel)
	return rel == PRDIndex || (strings.HasSuffix(rel, ".md") &&
		(dir == "docs/architecture" || dir == "docs/architecture/modules" || dir == "docs/adr"))
}

// ApprovedStatus 返回产物批准后的文件头状态：ADR 为 Accepted，其余为 Approved。
func ApprovedStatus(rel string) string {
	if path.Dir(rel) == "docs/adr" {
		return "Accepted"
	}
	return "Approved"
}

// ApprovalDocs 返回批准产物覆盖的文档：PRD 索引覆盖全部 PRD 文件，其余为文件本身。
func (s *Snapshot) ApprovalDocs(rel string) []*model.Doc {
	if rel == PRDIndex {
		return s.PRDDocs()
	}
	if d := s.Doc(rel); d != nil {
		return []*model.Doc{d}
	}
	return nil
}

// ApprovalHash 返回产物当前内容的批准哈希。文件头的 Status 行不参与计算：
// harness approve 写入 Approved 不应让刚记录的批准失效。
func (s *Snapshot) ApprovalHash(rel string) string {
	var parts []string
	for _, d := range s.ApprovalDocs(rel) {
		lines := append([]string(nil), d.Lines...)
		if i := HeaderStatusLine(lines); i >= 0 {
			lines = append(lines[:i], lines[i+1:]...)
		}
		parts = append(parts, "<<"+d.Path+">>\n"+strings.Join(lines, "\n"))
	}
	return lock.Hash(strings.Join(parts, "\n"))
}

var statusLineRe = regexp.MustCompile(`^- Status:\s*(.*)$`)

// HeaderStatusLine 返回文件头 Status 字段所在行的下标（从 0 开始）：第一个二级或三级标题之前的
// 第一个 "- Status:" 行。没有时返回 -1。
func HeaderStatusLine(lines []string) int {
	for i, l := range lines {
		if strings.HasPrefix(l, "## ") || strings.HasPrefix(l, "### ") {
			return -1
		}
		if statusLineRe.MatchString(l) {
			return i
		}
	}
	return -1
}

// HeaderStatus 返回文档文件头的 Status 取值，没有时为空串。
func HeaderStatus(d *model.Doc) string {
	if i := HeaderStatusLine(d.Lines); i >= 0 {
		return strings.TrimSpace(statusLineRe.FindStringSubmatch(d.Lines[i])[1])
	}
	return ""
}
