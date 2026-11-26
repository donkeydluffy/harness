// Package workspace 一次性加载规则和门禁需要的全部输入，形成只读快照（MOD-09）。
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/lock"
	"github.com/donkeydluffy/harness/internal/model"
	"github.com/donkeydluffy/harness/internal/testreport"
	"github.com/donkeydluffy/harness/internal/vcs"
)

// Snapshot 是一次检查的只读输入。规则只读快照，不自行读文件。
type Snapshot struct {
	Root    string
	Config  *Config
	Project *model.Project
	Lock    *lock.Lock // 锁文件不存在时为空 Lock（空基线）
	Repo    vcs.Repo   // 非 git 仓库时为 nil
	// CommitRange 是要检查的提交范围；为空时依赖 git 的规则检查暂存区（即将形成的提交）。
	CommitRange string
	Message     string                 // 见 LoadOptions.Message
	Sources     []*SourceFile          // LoadOptions.WithSources 时加载
	Tests       []testreport.TestCase  // LoadOptions.WithTests 时加载
	Reports     []string               // 读到的测试报告文件；为空表示没有报告
	Lint        *testreport.LintReport // LoadOptions.WithTests 时读取 lint 报告；不存在时为 nil
	LintErr     error                  // lint 报告无法读取或解析
	Diags       []model.Diagnostic
}

// LoadOptions 控制加载哪些可选输入；门禁路径使用零值以控制耗时。
type LoadOptions struct {
	WithTests, WithSources bool
	CommitRange            string
	// Message 是即将形成的提交的信息（commit-msg hook 传入）；检查暂存区时作为该提交的信息。
	Message string
}

// ArtifactGlobs 是作为产物解析的 Markdown 文件，路径相对项目根目录。
var ArtifactGlobs = []string{
	"docs/PRD.md",
	"docs/prd/*.md",
	"docs/architecture/*.md",
	"docs/architecture/modules/*.md",
	"docs/adr/*.md",
	"docs/tasks.md",
	"docs/debt.md",
	"docs/progress.md",
}

// Load 读取配置和全部产物并组装项目模型（IF-009）。文件存在但不可读时返回错误。
func Load(root string, opt LoadOptions) (*Snapshot, error) {
	cfg, err := LoadConfig(filepath.Join(root, ConfigFile))
	if err != nil {
		return nil, err
	}
	lk, err := lock.Read(filepath.Join(root, lock.File))
	if err != nil {
		return nil, err
	}
	s := &Snapshot{Root: root, Config: cfg, Lock: lk, CommitRange: opt.CommitRange, Message: opt.Message}
	if r, err := vcs.Open(root); err == nil {
		s.Repo = r
	}
	if opt.WithSources {
		if s.Sources, err = loadSources(root, cfg.SourcePaths); err != nil {
			return nil, err
		}
	}
	if opt.WithTests {
		if s.Reports, s.Tests, err = loadReports(root, cfg.TestReports); err != nil {
			return nil, err
		}
		s.Lint, s.LintErr = testreport.ReadLint(filepath.Join(root, testreport.LintFile))
	}
	var docs []*model.Doc
	for _, rel := range artifactFiles(root) {
		src, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return nil, fmt.Errorf("读取产物 %s：%w", rel, err)
		}
		d, diags := model.ParseMarkdown(rel, src)
		docs = append(docs, d)
		s.Diags = append(s.Diags, diags...)
	}
	p, diags := model.Build(docs)
	s.Project = p
	s.Diags = append(s.Diags, diags...)
	return s, nil
}

func artifactFiles(root string) []string {
	seen := map[string]bool{}
	var out []string
	for _, g := range ArtifactGlobs {
		matches, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(g)))
		sort.Strings(matches)
		for _, m := range matches {
			rel, _ := filepath.Rel(root, m)
			rel = filepath.ToSlash(rel)
			if !seen[rel] {
				seen[rel] = true
				out = append(out, rel)
			}
		}
	}
	return out
}

// Doc 返回指定相对路径的产物，不存在时返回 nil。
func (s *Snapshot) Doc(rel string) *model.Doc {
	for _, d := range s.Project.Docs {
		if d.Path == rel {
			return d
		}
	}
	return nil
}

// PRDDocs 返回 PRD 索引及按领域拆分的 PRD 文件，索引在前。
func (s *Snapshot) PRDDocs() []*model.Doc {
	var out []*model.Doc
	for _, d := range s.Project.Docs {
		if IsPRDPath(d.Path) {
			out = append(out, d)
		}
	}
	return out
}

// PRDIndex 是 PRD 索引文件的路径。
const PRDIndex = "docs/PRD.md"

// IsPRDPath 判断路径是否为 PRD 索引或拆分文件。
func IsPRDPath(rel string) bool {
	return rel == PRDIndex || (strings.HasPrefix(rel, "docs/prd/") && strings.HasSuffix(rel, ".md") && !strings.Contains(rel[len("docs/prd/"):], "/"))
}

// ConfigFile 是项目配置文件名。
const ConfigFile = "harness.json"

// Config 是 harness.json 的内容，字段见 docs/architecture/formats.md 第 6 节。
type Config struct {
	Languages         []string          `json:"languages"`
	SourcePaths       []string          `json:"sourcePaths"`
	TestReports       []string          `json:"testReports"`
	Start             string            `json:"start"` // 启动命令；没有可启动应用时写 none: <原因>（AC-055-1）
	Smoke             string            `json:"smoke"`
	SensitivePaths    []string          `json:"sensitivePaths"`
	SensitiveAllow    []string          `json:"sensitiveAllow"`
	DangerousCommands []string          `json:"dangerousCommands"`
	VagueWords        []string          `json:"vagueWords"`
	CommitMaxLines    int               `json:"commitMaxLines"`
	StopMaxBlocks     int               `json:"stopMaxBlocks"`
	StaleCommits      int               `json:"staleCommits"`
	Phases            []json.RawMessage `json:"phases"`
	Rules             []json.RawMessage `json:"rules"`
	// Tools 按语言覆盖格式化检查（format）与静态检查（lint）命令（AC-014-3）。
	Tools map[string]map[string]string `json:"tools"`
}

// DefaultConfig 返回全部字段取默认值的配置。
func DefaultConfig() *Config {
	return &Config{
		Languages:         []string{},
		SourcePaths:       []string{},
		TestReports:       []string{".harness/reports/*.xml", ".harness/reports/*.jsonl"},
		SensitivePaths:    []string{".env*", "*.pem", "*.key", "id_rsa*"},
		SensitiveAllow:    []string{".env.example"},
		DangerousCommands: []string{"git push --force", "git push -f", "curl * | sh", "wget * | sh"},
		VagueWords:        []string{"快速", "友好", "适当", "尽量", "等等", "之类", "如有必要", "和/或", "大约", "灵活", "高效", "包括但不限于"},
		CommitMaxLines:    400,
		StopMaxBlocks:     3,
		StaleCommits:      10,
	}
}

// LoadConfig 读取配置，缺省字段填默认值；文件不存在时返回默认配置（IF-010）。
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	src, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取配置 %s：%w", path, err)
	}
	// 只覆盖文件中出现的字段，其余保持默认值。
	if err := json.Unmarshal(src, cfg); err != nil {
		return nil, fmt.Errorf("解析配置 %s：%w", path, err)
	}
	return cfg, nil
}
