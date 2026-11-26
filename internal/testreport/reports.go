package testreport

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// LintIssue 是格式化检查或静态检查发现的一个问题；Line 为 0 表示不针对某一行（IF-033）。
type LintIssue struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Tool string `json:"tool"`
	Msg  string `json:"msg"`
}

// LintReport 是 harness check lint 的报告（AC-014-4）。
type LintReport struct {
	SourceHash string      `json:"sourceHash"`
	Issues     []LintIssue `json:"issues"`
}

// IFCoverage 是一个接口的单元测试结果。
type IFCoverage struct {
	Tests    []string `json:"tests"`
	Passed   bool     `json:"passed"`
	EntryHit bool     `json:"entryHit"`
}

// UnitReport 是 harness check unit 的报告（AC-021-2、AC-021-3）。
type UnitReport struct {
	SourceHash string                `json:"sourceHash"`
	Interfaces map[string]IFCoverage `json:"interfaces"`
	Modules    map[string]float64    `json:"modules"`
	Skipped    []string              `json:"skipped"`
}

// 报告的默认路径。
const (
	LintFile = ".harness/reports/lint.json"
	UnitFile = ".harness/reports/unit.json"
)

// ReadLint 读取 lint 报告；文件不存在时返回 nil 与 nil 错误（IF-033）。
//
// harness:IF-033
func ReadLint(path string) (*LintReport, error) {
	var r LintReport
	if ok, err := readJSON(path, &r); !ok {
		return nil, err
	}
	return &r, nil
}

// ReadUnit 读取 unit 报告；文件不存在时返回 nil 与 nil 错误。
func ReadUnit(path string) (*UnitReport, error) {
	var r UnitReport
	if ok, err := readJSON(path, &r); !ok {
		return nil, err
	}
	return &r, nil
}

// Write 写入报告，问题按文件、行号排序。
func (r *LintReport) Write(path string) error {
	if r.Issues == nil {
		r.Issues = []LintIssue{}
	}
	sort.SliceStable(r.Issues, func(i, j int) bool {
		a, b := r.Issues[i], r.Issues[j]
		if a.File != b.File {
			return a.File < b.File
		}
		return a.Line < b.Line
	})
	return writeJSON(path, r)
}

// Write 写入报告。
func (r *UnitReport) Write(path string) error {
	if r.Skipped == nil {
		r.Skipped = []string{}
	}
	return writeJSON(path, r)
}

func readJSON(path string, v any) (bool, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return false, err
	}
	return true, nil
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
