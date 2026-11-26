package workspace

import (
	"encoding/json"
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
// 键为 FR/NFR、用例、IF、CHG、ADR 编号与 config；已删除的需求和用例值前加 removed:。
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
	for _, e := range s.Project.ByKind["ADR"] {
		items[e.ID] = lock.Hash(e.Raw)
	}
	if src, err := os.ReadFile(filepath.Join(s.Root, ConfigFile)); err == nil {
		items["config"] = ConfigHash(src)
	}
	return items
}

// ConfigHash 返回 harness.json 的规范化哈希：按 JSON 解析，去掉键为 "//" 的注释项，键排序后紧凑序列化，
// 空白、键顺序与注释的变化不改变哈希。不是合法 JSON 时按文本规范化（formats.md 第 6 节）。
func ConfigHash(src []byte) string {
	var v any
	if err := json.Unmarshal(src, &v); err != nil {
		return lock.Hash(string(src))
	}
	b, _ := json.Marshal(stripComments(v))
	return lock.Hash(string(b))
}

func stripComments(v any) any {
	switch t := v.(type) {
	case map[string]any:
		delete(t, "//")
		for k, x := range t {
			t[k] = stripComments(x)
		}
	case []any:
		for i, x := range t {
			t[i] = stripComments(x)
		}
	}
	return v
}
