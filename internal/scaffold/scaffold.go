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

	"harness/internal/adapter"
	"harness/internal/lang"
	"harness/internal/workspace"
)

//go:embed all:templates
var templates embed.FS

// InitOptions 是 init 的选项（IF-020）。
type InitOptions struct {
	Languages []string
	Agents    []string // 为空时为 claudecode
	CI        string
	Binary    string // hook 配置中调用的 harness 命令，为空时为 harness
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

// Languages 返回有语言配置包且已注册的语言。
func Languages() []string {
	var out []string
	for _, l := range lang.All() {
		if _, err := fs.Stat(templates, "templates/langs/"+l.Name()+"/pack.json"); err == nil {
			out = append(out, l.Name())
		}
	}
	return out
}

// modulePath 是 CI 模板中安装 harness 用的模块路径（DEBT-001：发布前改为仓库地址）。
const modulePath = "harness"

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
func Generate(root string, opt InitOptions) (Result, error) {
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
	if err := copyTree(out, "templates/artifacts"); err != nil {
		return nil, err
	}
	if err := copyTreeTo(out, "templates/skills", ".claude/skills"); err != nil {
		return nil, err
	}
	cfg := struct {
		Languages   []string `json:"languages"`
		SourcePaths []string `json:"sourcePaths"`
	}{Languages: []string{}, SourcePaths: []string{}}
	for _, name := range opt.Languages {
		if !contains(supported, name) {
			return nil, &UnsupportedError{"语言", name, supported}
		}
		if contains(cfg.Languages, name) {
			continue
		}
		cfg.Languages = append(cfg.Languages, name)
		b, err := templates.ReadFile("templates/langs/" + name + "/pack.json")
		if err != nil {
			return nil, err
		}
		var pack struct{ SourcePaths []string }
		if err := json.Unmarshal(b, &pack); err != nil {
			return nil, fmt.Errorf("语言配置包 %s 的 pack.json 有误：%w", name, err)
		}
		for _, p := range pack.SourcePaths {
			if !contains(cfg.SourcePaths, p) {
				cfg.SourcePaths = append(cfg.SourcePaths, p)
			}
		}
		if err := copyTree(out, "templates/langs/"+name+"/files"); err != nil {
			return nil, err
		}
	}
	b, _ := json.MarshalIndent(cfg, "", "  ")
	out[workspace.ConfigFile] = append(b, '\n')
	ci := opt.CI
	if ci == "" {
		ci = "github"
	}
	ciPath, ok := ciFiles[ci]
	if !ok {
		return nil, &UnsupportedError{"CI", ci, []string{"github", "gitlab"}}
	}
	tmpl, err := templates.ReadFile("templates/ci/" + ci + ".yml")
	if err != nil {
		return nil, err
	}
	out[ciPath] = render(tmpl, opt)
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
