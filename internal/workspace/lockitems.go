package workspace

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"harness/internal/lock"
	"harness/internal/model"
)

// Removed 判断需求实体是否为 Removed 状态。
func Removed(e *model.Entity) bool { return e.Fields["Status"] == "Removed" }

// CaseRemoved 判断用例是否已删除：文字以 Removed 开头（formats.md 第 2 节）。
func CaseRemoved(text string) bool { return strings.HasPrefix(strings.TrimSpace(text), "Removed") }

var chgIDRe = regexp.MustCompile(`^CHG-\d+$`)

// LockItems 计算当前产物应写入锁文件的条目（formats.md 第 7 节）。
// 键为 FR/NFR、用例、IF、CHG 编号与 config；已删除的需求和用例值前加 removed:。
func (s *Snapshot) LockItems() map[string]string {
	items := map[string]string{}
	for _, d := range s.PRDDocs() {
		for _, e := range d.Entities {
			if e.Kind != "FR" && e.Kind != "NFR" {
				continue
			}
			h := lock.Hash(e.Raw)
			if Removed(e) {
				h = lock.RemovedPrefix + h
			}
			items[e.ID] = h
			for _, it := range e.Items {
				for _, c := range it.Children {
					ch := lock.Hash(c.Text)
					if CaseRemoved(c.Text) || Removed(e) {
						ch = lock.RemovedPrefix + ch
					}
					items[c.ID] = ch
				}
			}
		}
	}
	for _, e := range s.Project.ByKind["IF"] {
		items[e.ID] = lock.Hash(e.Raw)
	}
	rows := map[string][]string{}
	for _, c := range s.Project.Changes {
		if chgIDRe.MatchString(c.ID) {
			rows[c.ID] = append(rows[c.ID], strings.Join([]string{c.ID, c.Date, c.Action, strings.Join(c.Targets, ", "), c.Note}, " | "))
		}
	}
	for id, r := range rows {
		items[id] = lock.Hash(strings.Join(r, "\n"))
	}
	if src, err := os.ReadFile(filepath.Join(s.Root, ConfigFile)); err == nil {
		items["config"] = lock.Hash(string(src))
	}
	return items
}
