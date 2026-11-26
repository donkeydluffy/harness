package builtin

import (
	"regexp"

	"github.com/donkeydluffy/harness/internal/rules/engine"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// prd10：Quantitative 类 NFR 的每条验收标准至少包含一个数值和单位。
type prd10 struct{}

func (prd10) ID() string       { return "prd.10" }
func (prd10) Covers() []string { return []string{"AC-040-2"} }
func (prd10) Fix() string {
	return "在验收标准中写出数值和单位，如 \"p95 < 200ms\"、\"< 10s\"、\"≤ 400 行\""
}

var quantityRe = regexp.MustCompile(`\d+(\.\d+)?\s*(ms|毫秒|s|秒|min|分钟|h|小时|天|%|B|KB|KiB|MB|MiB|GB|GiB|字节|行|个|次|条|req/s|qps|rps)`)

func (prd10) Check(s *workspace.Snapshot) []engine.Violation {
	var out []engine.Violation
	for _, e := range requirements(s) {
		if e.Kind != "NFR" || e.Fields["Type"] != "Quantitative" {
			continue
		}
		for _, it := range e.Items {
			if !quantityRe.MatchString(it.Text) {
				out = append(out, engine.Violation{File: e.File, Line: it.Line,
					Msg: e.ID + " 是 Quantitative 类，验收标准 " + it.ID + " 没有数值和单位"})
			}
		}
	}
	return out
}
