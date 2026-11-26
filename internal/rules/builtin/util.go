package builtin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"harness/internal/model"
	"harness/internal/rules/engine"
	"harness/internal/workspace"
)

// 产物编号的合法格式（docs/architecture/formats.md 第 2 节）。
var idFormat = map[string]*regexp.Regexp{
	"FR":   regexp.MustCompile(`^FR-\d{3,}$`),
	"NFR":  regexp.MustCompile(`^NFR-\d{3,}$`),
	"OQ":   regexp.MustCompile(`^OQ-\d{3,}$`),
	"TERM": regexp.MustCompile(`^TERM-\d{3,}$`),
	"CHG":  regexp.MustCompile(`^CHG-\d{3,}$`),
}

// requirements 返回全部 PRD 文件中的 FR 与 NFR，按文件和行号顺序。
func requirements(s *workspace.Snapshot) []*model.Entity {
	var out []*model.Entity
	for _, d := range s.PRDDocs() {
		for _, e := range d.Entities {
			if e.Kind == "FR" || e.Kind == "NFR" {
				out = append(out, e)
			}
		}
	}
	return out
}

func at(e *model.Entity, format string, args ...any) engine.Violation {
	return engine.Violation{File: e.File, Line: e.Line, Msg: fmt.Sprintf(format, args...)}
}

func positiveInt(v string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(v))
	return err == nil && n > 0
}

func oneOf(v string, allowed ...string) bool {
	for _, a := range allowed {
		if v == a {
			return true
		}
	}
	return false
}

// hasAC 判断列表项中是否至少有一条验收标准。
func hasAC(items []*model.Item) bool {
	for _, it := range items {
		if strings.HasPrefix(it.ID, "AC-") {
			return true
		}
	}
	return false
}
