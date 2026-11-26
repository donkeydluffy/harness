// Package testreport 解析 JUnit XML 与 go test -json，输出统一的测试结果（MOD-08）。
package testreport

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Status 是测试结果。
type Status string

const (
	Passed  Status = "passed"
	Failed  Status = "failed"
	Skipped Status = "skipped"
)

// TestCase 是一个测试的结果（IF-008）。
type TestCase struct {
	Name, Suite, File string
	Status            Status
}

type junitCase struct {
	Name      string    `xml:"name,attr"`
	Classname string    `xml:"classname,attr"`
	File      string    `xml:"file,attr"`
	Failure   *struct{} `xml:"failure"`
	Error     *struct{} `xml:"error"`
	Skipped   *struct{} `xml:"skipped"`
}

type junitSuite struct {
	Name   string       `xml:"name,attr"`
	Cases  []junitCase  `xml:"testcase"`
	Suites []junitSuite `xml:"testsuite"`
}

// ParseJUnit 解析 JUnit XML（根元素为 testsuites 或 testsuite）（IF-008）。
func ParseJUnit(r io.Reader) ([]TestCase, error) {
	var root struct {
		XMLName xml.Name
		junitSuite
	}
	if err := xml.NewDecoder(r).Decode(&root); err != nil {
		return nil, fmt.Errorf("解析 JUnit XML：%w", err)
	}
	if root.XMLName.Local != "testsuites" && root.XMLName.Local != "testsuite" {
		return nil, fmt.Errorf("解析 JUnit XML：根元素 %s 不是 testsuites 或 testsuite", root.XMLName.Local)
	}
	var out []TestCase
	var walk func(s junitSuite)
	walk = func(s junitSuite) {
		for _, c := range s.Cases {
			st := Passed
			switch {
			case c.Failure != nil || c.Error != nil:
				st = Failed
			case c.Skipped != nil:
				st = Skipped
			}
			suite := c.Classname
			if suite == "" {
				suite = s.Name
			}
			out = append(out, TestCase{Name: c.Name, Suite: suite, File: c.File, Status: st})
		}
		for _, sub := range s.Suites {
			walk(sub)
		}
	}
	walk(root.junitSuite)
	return out, nil
}

// ParseGoJSON 解析 go test -json 输出；子测试名含父测试前缀（如 TestA/sub）（IF-008）。
func ParseGoJSON(r io.Reader) ([]TestCase, error) {
	type event struct{ Action, Package, Test string }
	idx := map[string]int{}
	var out []TestCase
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] != '{' {
			continue // go test 在构建失败时会输出非 JSON 行
		}
		var e event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			return nil, fmt.Errorf("解析 go test -json 第 %d 行：%w", n, err)
		}
		var st Status
		switch e.Action {
		case "pass":
			st = Passed
		case "fail":
			st = Failed
		case "skip":
			st = Skipped
		default:
			continue
		}
		if e.Test == "" {
			continue // 包级结果
		}
		key := e.Package + "\x00" + e.Test
		if i, ok := idx[key]; ok {
			out[i].Status = st
			continue
		}
		idx[key] = len(out)
		out = append(out, TestCase{Name: e.Test, Suite: e.Package, Status: st})
	}
	return out, sc.Err()
}

// ParseFile 按扩展名分派：.xml 为 JUnit，.jsonl / .json 为 go test -json（IF-008）。
func ParseFile(path string) ([]TestCase, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	switch strings.ToLower(filepath.Ext(path)) {
	case ".xml":
		return ParseJUnit(f)
	case ".jsonl", ".json":
		return ParseGoJSON(f)
	default:
		return nil, fmt.Errorf("不支持的测试报告格式：%s（支持 .xml、.jsonl、.json）", path)
	}
}
