// Package toolrun 运行外部工具（格式化与静态检查、单元测试与覆盖率、冒烟测试）并写报告（MOD-16）。
// 全量检查不调用本模块，只读取报告（ADR-011）。
package toolrun

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/workspace"
)

// MissingTool 是工具未安装时的错误；Hint 为安装命令，项目覆盖的命令没有内置提示时给出命令名（AC-014-2）。
type MissingTool struct{ Language, Command, Hint string }

func (e *MissingTool) Error() string {
	return fmt.Sprintf("%s 的工具 %s 未安装：%s", e.Language, e.Command, e.Hint)
}

// Lint 对项目声明且存在源文件的每种语言，运行格式化检查与静态检查（项目配置的 tools 优先，否则用语言配置包的默认值），
// 把问题与源码哈希写入 .harness/reports/lint.json（AC-014-3、AC-014-4）。快照须以 WithSources 加载。
//
// harness:IF-030
func Lint(s *workspace.Snapshot) (*testreport.LintReport, error) {
	rep := &testreport.LintReport{SourceHash: s.SourceHash(), Issues: []testreport.LintIssue{}}
	for _, name := range s.Config.Languages {
		pack, ok := lang.PackFor(name)
		if !ok || !hasSources(s, name) {
			continue // 没有该语言的源文件时不运行它的工具（AC-014-4/B1）
		}
		for _, which := range []string{"format", "lint"} {
			t := pack.Format
			if which == "lint" {
				t = pack.Lint
			}
			if cmd := s.Config.Tools[name][which]; cmd != "" {
				t = lang.Tool{Cmd: cmd, Output: "lines", Install: "安装 " + firstWord(cmd) + " 后重试"}
			}
			if _, err := exec.LookPath(firstWord(t.Cmd)); err != nil {
				return nil, &MissingTool{Language: name, Command: firstWord(t.Cmd), Hint: t.Install}
			}
			out, code := run(s.Root, t.Cmd)
			rep.Issues = append(rep.Issues, parse(t, out, code)...)
		}
	}
	return rep, rep.Write(filepath.Join(s.Root, testreport.LintFile))
}

// Unit 写入单元测试报告（IF-031）。Release 2 的覆盖率采集由 T-049 接入 Go；在此之前全部语言记为 Skipped。
//
// harness:IF-031
func Unit(s *workspace.Snapshot) (*testreport.UnitReport, error) {
	rep := &testreport.UnitReport{SourceHash: s.SourceHash(), Interfaces: map[string]testreport.IFCoverage{},
		Modules: map[string]float64{}, Skipped: append([]string{}, s.Config.Languages...)}
	return rep, rep.Write(filepath.Join(s.Root, testreport.UnitFile))
}

// Smoke 运行项目配置的冒烟测试命令，返回退出码与输出的最后 50 行（AC-055-2）。没有声明命令时返回错误。
//
// harness:IF-032
func Smoke(s *workspace.Snapshot) (exit int, tail []string, err error) {
	cmd := strings.TrimSpace(s.Config.Smoke)
	if cmd == "" {
		return 0, nil, fmt.Errorf("项目配置没有声明冒烟测试命令（harness.json 的 smoke）")
	}
	out, code := run(s.Root, cmd)
	return code, lastLines(out, 50), nil
}

func hasSources(s *workspace.Snapshot, name string) bool {
	for _, f := range s.Sources {
		if f.Lang != nil && f.Lang.Name() == name {
			return true
		}
	}
	return false
}

// run 在 root 以 sh -c 运行命令，返回合并的标准输出与标准错误、退出码。
func run(root, cmd string) (string, int) {
	c := exec.Command("sh", "-c", cmd)
	c.Dir = root
	var buf bytes.Buffer
	c.Stdout, c.Stderr = &buf, &buf
	err := c.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		code = -1
		buf.WriteString(err.Error())
	}
	return buf.String(), code
}

func firstWord(cmd string) string {
	if f := strings.Fields(cmd); len(f) > 0 {
		return f[0]
	}
	return ""
}

// toolName 取命令开头直到第一个参数（以 - . / $ 开头或含 /）的词，如 "go vet"、"gofmt"、"ruff check"。
func toolName(cmd string) string {
	var words []string
	for _, w := range strings.Fields(cmd) {
		if strings.ContainsAny(w[:1], "-./$") || strings.Contains(w, "/") {
			break
		}
		words = append(words, w)
	}
	if len(words) == 0 {
		return firstWord(cmd)
	}
	return strings.Join(words, " ")
}

var lineRe = regexp.MustCompile(`^(?:vet: )?(\S+?):(\d+)(?::\d+)?:\s*(.*)$`)

// parse 按工具的输出格式解析问题；退出码非 0 而没有解析出问题时，记一条不带文件的问题，消息为输出的最后 20 行。
func parse(t lang.Tool, out string, code int) []testreport.LintIssue {
	name := toolName(t.Cmd)
	var issues []testreport.LintIssue
	for _, l := range strings.Split(out, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		switch t.Output {
		case "files":
			f := strings.TrimPrefix(l, "Would reformat: ")
			if strings.ContainsAny(f, " \t") {
				continue // 汇总行，不是文件名
			}
			issues = append(issues, testreport.LintIssue{File: clean(f), Tool: name, Msg: "未按 " + name + " 的要求格式化"})
		default:
			if m := lineRe.FindStringSubmatch(l); m != nil {
				n, _ := strconv.Atoi(m[2])
				issues = append(issues, testreport.LintIssue{File: clean(m[1]), Line: n, Tool: name, Msg: m[3]})
			}
		}
	}
	if code != 0 && len(issues) == 0 {
		issues = append(issues, testreport.LintIssue{Tool: name, Msg: fmt.Sprintf("退出码 %d：%s", code, strings.Join(lastLines(out, 20), "\n"))})
	}
	return issues
}

func clean(p string) string { return filepath.ToSlash(filepath.Clean(strings.TrimPrefix(p, "./"))) }

func lastLines(out string, n int) []string {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}
