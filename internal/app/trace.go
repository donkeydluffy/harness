package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"harness/internal/workspace"
)

// TraceRow 是覆盖矩阵的一行：一条纳入范围的需求及覆盖它的有效设计元素。
type TraceRow struct {
	Requirement string   `json:"requirement"`
	Title       string   `json:"title"`
	Elements    []string `json:"elements"`
}

// TraceFormats 是 harness trace 支持的输出格式。
var TraceFormats = []string{"md", "json"}

// Trace 输出需求 × 设计元素的覆盖矩阵（IF-022），格式为 md 或 json。
func Trace(root, format string) ([]byte, error) {
	if format != "md" && format != "json" {
		return nil, fmt.Errorf("不支持的格式 %q，支持：%s", format, strings.Join(TraceFormats, ", "))
	}
	s, err := workspace.Load(root, workspace.LoadOptions{})
	if err != nil {
		return nil, err
	}
	rows := traceRows(s)
	if format == "json" {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		err := enc.Encode(struct {
			Rows []TraceRow `json:"rows"`
		}{rows})
		return buf.Bytes(), err
	}
	var b strings.Builder
	b.WriteString("| 需求 | 标题 | 设计元素 |\n|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", r.Requirement, r.Title, strings.Join(r.Elements, ", "))
	}
	return []byte(b.String()), nil
}

// traceRows 计算纳入范围的 Active 需求及覆盖它的有效设计元素。
// 覆盖口径与规则 trace.01 相同：Covers 或 Relates 引用该需求，且元素不是 Superseded 或 Removed。
func traceRows(s *workspace.Snapshot) []TraceRow {
	cover := map[string][]string{}
	for _, d := range s.Project.Docs {
		if !strings.HasPrefix(d.Path, "docs/architecture/") && !strings.HasPrefix(d.Path, "docs/adr/") {
			continue
		}
		for _, e := range d.Entities {
			switch e.Kind {
			case "MOD", "IF", "FLOW", "EXT", "ADR":
			default:
				continue
			}
			if st := e.Fields["Status"]; st == "Superseded" || st == "Removed" {
				continue
			}
			for _, r := range append(e.List("Covers"), e.List("Relates")...) {
				cover[r] = append(cover[r], e.ID)
			}
		}
	}
	current := 0
	if idx := s.Doc(workspace.PRDIndex); idx != nil {
		current, _ = strconv.Atoi(idx.Header["Current-Release"])
	}
	rows := []TraceRow{}
	for _, d := range s.PRDDocs() {
		for _, e := range d.Entities {
			if (e.Kind != "FR" && e.Kind != "NFR") || e.Fields["Status"] != "Active" {
				continue
			}
			if rel, err := strconv.Atoi(e.Fields["Release"]); err == nil && current > 0 && rel > current {
				continue
			}
			els := append([]string{}, cover[e.ID]...)
			sort.Strings(els)
			rows = append(rows, TraceRow{Requirement: e.ID, Title: e.Title, Elements: els})
		}
	}
	return rows
}
