// Package scaffold 生成项目骨架：产物模板、语言配置包、项目配置、代理 hook 配置（MOD-13）。
// 模板通过 go:embed 打包进二进制（AC-N002-2）。
package scaffold

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/adapter"
	"github.com/donkeydluffy/harness/internal/lang"
	"github.com/donkeydluffy/harness/internal/version"
	"github.com/donkeydluffy/harness/internal/workspace"
)

//go:embed all:templates
var templates embed.FS

// InitOptions 是 init 的选项（IF-020）。
type InitOptions struct {
	Languages []string
	Agents    []string // 为空时为 claudecode
	CI        string
	Binary    string // hook 配置中调用的 harness 命令，为空时为 harness
	// SelfHosted 表示目标项目就是 harness 自身（go.mod 的模块路径为 modulePath）：
	// CI 从当前提交的源码构建 harness，生成文件与同一提交的模板一致。
	SelfHosted bool
	Version    string // 当前 harness 版本（IF-037）
	CIVersion  string // --ci-version；为空时用 Version
}

// pin 返回 CI 中固定的 harness 版本。
func (o InitOptions) pin() string {
	if o.CIVersion != "" {
		return strings.TrimPrefix(o.CIVersion, "v")
	}
	return o.Version
}

// VersionError 表示 CI 要固定的版本不是发布版本号（AC-051-2）。
type VersionError struct{ Version string }

func (e *VersionError) Error() string {
	return fmt.Sprintf("当前 harness 版本 %q 不是发布版本号，生成的 CI 无法固定 harness 版本；改用发布版的 harness，或用 --ci-version <版本> 指定（如 --ci-version 0.2.0）", e.Version)
}

// gateIgnoreFile 是忽略门禁运行时目录的 .gitignore。
const gateIgnoreFile = ".harness/.gitignore"

// selfHosted 判断 root 下 go.mod 的模块路径是否为 harness 自身。
func selfHosted(root string) bool {
	b, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(b), "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[0] == "module" {
			return f[1] == modulePath
		}
	}
	return false
}

// Result 是 init 的结果（IF-020）：Written 为新写入的文件，Skipped 为已存在且内容不同而未覆盖的文件。
type Result struct{ Written, Skipped []string }

// UnsupportedError 表示选项中有不支持的语言或代理。
type UnsupportedError struct {
	What, Name string
	Supported  []string
}

func (e *UnsupportedError) Error() string {
	return fmt.Sprintf("不支持的%s %q，支持：%s", e.What, e.Name, strings.Join(e.Supported, ", "))
}

// Languages 返回有语言配置包（IF-034）的语言。
func Languages() []string {
	var out []string
	for _, l := range lang.All() {
		if _, ok := lang.PackFor(l.Name()); ok {
			out = append(out, l.Name())
		}
	}
	return out
}

// modulePath 是 CI 模板中安装 harness 用的模块路径，与 go.mod 的模块路径一致。
const modulePath = "github.com/donkeydluffy/harness"

// ciFiles 是 --ci 的取值与生成的 CI 配置文件。
var ciFiles = map[string]string{"github": ".github/workflows/harness.yml", "gitlab": ".gitlab-ci.yml"}

// gitHooks 是安装到 .git/hooks 的 hook。
var gitHooks = []string{"commit-msg", "pre-commit"}

// GitHooksDir 返回项目的 git hook 目录；不是 git 仓库（或 .git 不是目录，如工作树）时 ok 为 false。
func GitHooksDir(root string) (dir string, ok bool) {
	fi, err := os.Stat(filepath.Join(root, ".git"))
	if err != nil || !fi.IsDir() {
		return "", false
	}
	return filepath.Join(root, ".git", "hooks"), true
}

// Generate 写入模板；已存在且内容相同的文件不动，内容不同的跳过并列入 Skipped（AC-001-2）。
// 选项不合法时不写任何文件（AC-001-1/E1）。git 仓库中同时安装 commit-msg、pre-commit hook（AC-013-3）。
//
// harness:IF-020
func Generate(root string, opt InitOptions) (Result, error) {
	opt.SelfHosted = selfHosted(root)
	if !opt.SelfHosted && !version.Release(opt.pin()) {
		return Result{}, &VersionError{Version: opt.pin()}
	}
	files, err := Files(opt)
	if err != nil {
		return Result{}, err
	}
	exec := map[string]bool{}
	if _, ok := GitHooksDir(root); ok {
		for _, h := range gitHooks {
			b, err := templates.ReadFile("templates/githooks/" + h)
			if err != nil {
				return Result{}, err
			}
			p := ".git/hooks/" + h
			files[p], exec[p] = render(b, opt), true
		}
	}
	var paths []string
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	var res Result
	for _, p := range paths {
		abs := filepath.Join(root, filepath.FromSlash(p))
		old, err := os.ReadFile(abs)
		switch {
		case err == nil && bytes.Equal(old, files[p]):
			continue
		case err == nil:
			res.Skipped = append(res.Skipped, p)
			continue
		case !errors.Is(err, fs.ErrNotExist):
			return res, err
		}
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return res, fmt.Errorf("无法创建目录 %s：%w", filepath.Dir(abs), err)
		}
		mode := os.FileMode(0o644)
		if exec[p] {
			mode = 0o755
		}
		if err := os.WriteFile(abs, files[p], mode); err != nil {
			return res, fmt.Errorf("无法写入 %s：%w", abs, err)
		}
		res.Written = append(res.Written, p)
	}
	return res, nil
}

// Files 返回 init 生成的全部文件（仓库相对路径 → 内容）。
func Files(opt InitOptions) (map[string][]byte, error) {
	supported := Languages()
	if len(opt.Languages) == 0 {
		return nil, &UnsupportedError{"语言", "", supported}
	}
	agents := opt.Agents
	if len(agents) == 0 {
		agents = []string{"claudecode"}
	}
	binary := opt.Binary
	if binary == "" {
		binary = "harness"
	}
	out := map[string][]byte{}
	// 门禁的快照与会话计数是运行时状态：用 .harness/ 下的 .gitignore 忽略它们（连同它自身），
	// 不改项目根目录的 .gitignore，也不需要提交。
	out[gateIgnoreFile] = []byte("snap/\nsession/\n.gitignore\n")
	if err := copyTree(out, "templates/artifacts"); err != nil {
		return nil, err
	}
	if err := copyTreeTo(out, "templates/skills", ".claude/skills"); err != nil {
		return nil, err
	}
	cfg := struct {
		Languages   []string                     `json:"languages"`
		SourcePaths []string                     `json:"sourcePaths"`
		Tools       map[string]map[string]string `json:"tools"`
	}{Languages: []string{}, SourcePaths: []string{}, Tools: map[string]map[string]string{}}
	var tests []langTest
	for _, name := range opt.Languages {
		if !contains(supported, name) {
			return nil, &UnsupportedError{"语言", name, supported}
		}
		if contains(cfg.Languages, name) {
			continue
		}
		cfg.Languages = append(cfg.Languages, name)
		pack, _ := lang.PackFor(name)
		tests = append(tests, langTest{name, pack.Test})
		// 写明各语言的格式化与静态检查命令，项目可直接修改（AC-014-1、AC-014-3）。
		cfg.Tools[name] = map[string]string{"format": pack.Format.Cmd, "lint": pack.Lint.Cmd}
		for _, p := range pack.SourcePaths {
			if !contains(cfg.SourcePaths, p) {
				cfg.SourcePaths = append(cfg.SourcePaths, p)
			}
		}
		if err := copyTree(out, "templates/langs/"+name+"/files"); err != nil {
			return nil, err
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false) // 命令中的 >、& 原样写出
	enc.SetIndent("", "  ")
	enc.Encode(cfg)
	out[workspace.ConfigFile] = buf.Bytes()
	ci := opt.CI
	if ci == "" {
		ci = "github"
	}
	ciPath, ok := ciFiles[ci]
	if !ok {
		return nil, &UnsupportedError{"CI", ci, []string{"github", "gitlab"}}
	}
	b, err := ciConfig(ci, cfg.Languages, tests, opt)
	if err != nil {
		return nil, err
	}
	out[ciPath] = b
	for _, name := range agents {
		a, ok := adapter.Get(name)
		if !ok {
			return nil, &UnsupportedError{"代理", name, adapter.Names()}
		}
		p, content := a.HookConfig(binary)
		out[p] = content
	}
	return out, nil
}

// CIConfig 返回 harness init 为语言列表 langs 生成、固定 harness 版本 pin 的 CI 配置（ci 为 github 或 gitlab）。
func CIConfig(ci string, langs []string, pin string) ([]byte, error) {
	return ciConfig(ci, langs, langTests(langs), InitOptions{CIVersion: pin})
}

// ciConfig 渲染 CI 配置：安装 harness 的步骤、各语言的测试步骤与模板中的 {{binary}}、{{module}}。
func ciConfig(ci string, langs []string, tests []langTest, opt InitOptions) ([]byte, error) {
	tmpl, err := templates.ReadFile("templates/ci/" + ci + ".yml")
	if err != nil {
		return nil, err
	}
	out := strings.ReplaceAll(string(render(tmpl, opt)), "{{testSteps}}", testSteps(ci, tests))
	out = strings.ReplaceAll(out, "{{install}}", installSteps(ci, langs, opt))
	image := "buildpack-deps:stable"
	if opt.SelfHosted || contains(langs, "go") {
		image = "golang:latest"
	}
	return []byte(strings.ReplaceAll(out, "{{image}}", image)), nil
}

// releaseURL 是发布资产的下载地址前缀。
const releaseURL = "https://" + modulePath + "/releases/download/v"

// installSteps 渲染 CI 中安装 harness 的步骤：harness 自身仓库从当前提交的源码构建；用户项目下载固定版本的
// Linux amd64 发布归档并用 checksums.txt 校验，校验失败时步骤失败、后续检查不运行（AC-051-1）。
// 需要 Go 工具链的只有 harness 自身仓库与 Go 项目（测试与 go vet）。
func installSteps(ci string, langs []string, opt InitOptions) string {
	archive := "harness_" + opt.pin() + "_linux_amd64.tar.gz"
	download := []string{
		"curl -fsSLO " + releaseURL + opt.pin() + "/" + archive,
		"curl -fsSLO " + releaseURL + opt.pin() + "/checksums.txt",
		"grep ' " + archive + "$' checksums.txt | sha256sum -c -",
		"tar -xzf " + archive + " harness",
	}
	var b strings.Builder
	if ci == "github" {
		if opt.SelfHosted || contains(langs, "go") {
			b.WriteString("      - uses: actions/setup-go@v5\n        with:\n          go-version: stable\n")
		}
		if opt.SelfHosted {
			b.WriteString("      - name: 安装 harness\n        run: go install ./cmd/harness\n")
			return b.String()
		}
		fmt.Fprintf(&b, "      - name: 安装 harness %s\n        run: |\n", opt.pin())
		for _, l := range append(download, "sudo install harness /usr/local/bin/harness") {
			b.WriteString("          " + l + "\n")
		}
		return b.String()
	}
	if opt.SelfHosted {
		return "    - go install ./cmd/harness\n"
	}
	for _, l := range append(download, "install harness /usr/local/bin/harness") {
		b.WriteString("    - " + l + "\n")
	}
	return b.String()
}

// langTests 返回各语言配置包的 CI 测试命令；没有配置包的语言忽略。
func langTests(langs []string) []langTest {
	var out []langTest
	for _, name := range langs {
		if pack, ok := lang.PackFor(name); ok {
			out = append(out, langTest{name, pack.Test})
		}
	}
	return out
}

// langTest 是一种语言在 CI 中生成测试报告的命令。
type langTest struct{ lang, cmd string }

// testSteps 渲染 CI 配置中运行测试、把报告写到 .harness/reports/ 的步骤；
// 没有通用测试命令的语言留一行注释，指向该语言的说明。
func testSteps(ci string, tests []langTest) string {
	var b strings.Builder
	for _, t := range tests {
		switch {
		case t.cmd == "" && ci == "github":
			fmt.Fprintf(&b, "      # %s：没有通用的测试命令，测试报告的生成方式见 docs/lang/%s.md。\n", t.lang, t.lang)
		case t.cmd == "":
			fmt.Fprintf(&b, "    # %s：没有通用的测试命令，测试报告的生成方式见 docs/lang/%s.md。\n", t.lang, t.lang)
		case ci == "github":
			fmt.Fprintf(&b, "      - name: 测试（%s）\n        run: %s\n", t.lang, t.cmd)
		default:
			fmt.Fprintf(&b, "    - %s\n", t.cmd)
		}
	}
	return b.String()
}

// render 替换模板中的 {{binary}} 与 {{module}}。
func render(b []byte, opt InitOptions) []byte {
	binary := opt.Binary
	if binary == "" {
		binary = "harness"
	}
	s := strings.ReplaceAll(string(b), "{{binary}}", binary)
	return []byte(strings.ReplaceAll(s, "{{module}}", modulePath))
}

// copyTree 把内嵌目录 dir 下的文件按相对路径放入 out。
func copyTree(out map[string][]byte, dir string) error { return copyTreeTo(out, dir, "") }

// copyTreeTo 把内嵌目录 dir 下的文件放到项目的 to 目录下（to 为空时为项目根目录）。
func copyTreeTo(out map[string][]byte, dir, to string) error {
	return fs.WalkDir(templates, dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := templates.ReadFile(p)
		if err != nil {
			return err
		}
		out[path.Join(to, strings.TrimPrefix(p, dir+"/"))] = b
		return nil
	})
}

func contains(list []string, x string) bool {
	for _, v := range list {
		if v == x {
			return true
		}
	}
	return false
}
