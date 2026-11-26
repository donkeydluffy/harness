package builtin

import (
	"strings"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// adr02：Superseded 的 ADR 必须用 Superseded-By 指向存在的 ADR，且取代链无环。
type adr02 struct{}

func (adr02) ID() string       { return "adr.02" }
func (adr02) Covers() []string { return []string{"AC-010-2"} }
func (adr02) Fix() string {
	return "为 Superseded 的 ADR 加上 - Superseded-By: ADR-nnn，指向存在且未被反向取代的 ADR"
}

func (adr02) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range s.Project.ByKind["ADR"] {
		if e.Fields["Status"] != "Superseded" {
			continue
		}
		by := strings.TrimSpace(e.Fields["Superseded-By"])
		if by == "" {
			out = append(out, at(e, "%s 为 Superseded，但没有 Superseded-By", e.ID))
			continue
		}
		if t := s.Project.ByID[by]; t == nil || t.Kind != "ADR" {
			out = append(out, at(e, "%s 的 Superseded-By 指向不存在的 %s", e.ID, by))
			continue
		}
		// 沿取代链前进，回到自身即为环。
		seen := map[string]bool{e.ID: true}
		path := []string{e.ID}
		for cur := by; cur != ""; {
			path = append(path, cur)
			if seen[cur] {
				out = append(out, at(e, "ADR 取代关系存在环：%s", strings.Join(path, " -> ")))
				break
			}
			seen[cur] = true
			next := s.Project.ByID[cur]
			if next == nil || next.Fields["Status"] != "Superseded" {
				break
			}
			cur = strings.TrimSpace(next.Fields["Superseded-By"])
		}
	}
	return out
}
