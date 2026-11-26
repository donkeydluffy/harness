// Package vcs 通过调用系统 git 访问提交范围、提交改动、指定版本的文件内容（MOD-07）。
package vcs

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ErrNotRepo 表示目录不在 git 仓库中。
var ErrNotRepo = errors.New("不是 git 仓库")

// Repo 是 git 仓库的只读访问（IF-007）。
//
// harness:IF-007
type Repo interface {
	Root() string
	Commits(rangeSpec string) ([]Commit, error)
	Changes(c Commit) ([]FileChange, error)
	Staged() ([]FileChange, error)
	FileAt(rev, path string) ([]byte, error) // rev 为空表示暂存区
	CommitsTouching(paths []string, since string) (int, error)
}

// Commit 是一个提交。
type Commit struct {
	Hash, Message string
	Trailers      map[string][]string
}

// FileChange 是提交中一个文件的改动。Status 为 A、M、D、R 等。
type FileChange struct {
	Path           string
	Added, Deleted int
	Status         string
}

type repo struct{ root string }

// Open 打开 dir 所在的 git 仓库；非 git 仓库返回 ErrNotRepo（IF-007）。
func Open(dir string) (Repo, error) {
	out, err := run(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, ErrNotRepo
	}
	return &repo{root: strings.TrimSpace(string(out))}, nil
}

func run(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-c", "core.quotepath=off"}, args...)...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s：%v：%s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func (r *repo) Root() string { return r.root }

func (r *repo) Commits(rangeSpec string) ([]Commit, error) {
	out, err := run(r.root, "log", "--reverse", "--format=%H%x00%B%x1e", rangeSpec)
	if err != nil {
		return nil, err
	}
	var cs []Commit
	for _, rec := range strings.Split(string(out), "\x1e") {
		rec = strings.TrimLeft(rec, "\n")
		hash, msg, ok := strings.Cut(rec, "\x00")
		if !ok {
			continue
		}
		msg = strings.TrimRight(msg, "\n")
		cs = append(cs, Commit{Hash: hash, Message: msg, Trailers: ParseTrailers(msg)})
	}
	return cs, nil
}

var trailerRe = regexp.MustCompile(`^([A-Za-z][A-Za-z0-9-]*):\s*(.+)$`)

// ParseTrailers 解析提交信息最后一段中形如 "Key: value" 的尾注；最后一段必须全部是尾注行。
func ParseTrailers(msg string) map[string][]string {
	paras := strings.Split(strings.TrimSpace(strings.ReplaceAll(msg, "\r\n", "\n")), "\n\n")
	t := map[string][]string{}
	if len(paras) < 2 {
		return t
	}
	for _, l := range strings.Split(paras[len(paras)-1], "\n") {
		m := trailerRe.FindStringSubmatch(strings.TrimSpace(l))
		if m == nil {
			return map[string][]string{}
		}
		t[m[1]] = append(t[m[1]], strings.TrimSpace(m[2]))
	}
	return t
}

func (r *repo) Changes(c Commit) ([]FileChange, error) {
	return r.changes("diff-tree", "--root", "--no-commit-id", "-r", "-M", c.Hash)
}

func (r *repo) Staged() ([]FileChange, error) {
	return r.changes("diff", "--cached", "-M")
}

func (r *repo) changes(args ...string) ([]FileChange, error) {
	num, err := run(r.root, append(append([]string{}, args...), "--numstat")...)
	if err != nil {
		return nil, err
	}
	st, err := run(r.root, append(append([]string{}, args...), "--name-status")...)
	if err != nil {
		return nil, err
	}
	status := map[string]string{}
	for _, l := range strings.Split(strings.TrimSpace(string(st)), "\n") {
		f := strings.Split(l, "\t")
		if len(f) >= 2 {
			status[f[len(f)-1]] = f[0][:1]
		}
	}
	var out []FileChange
	for _, l := range strings.Split(strings.TrimSpace(string(num)), "\n") {
		f := strings.Split(l, "\t")
		if len(f) < 3 {
			continue
		}
		path := f[len(f)-1]
		if i := strings.Index(path, " => "); i >= 0 { // 重命名 a => b 或 dir/{a => b}
			path = renamed(path)
		}
		a, _ := strconv.Atoi(f[0]) // 二进制文件为 "-"，计 0
		d, _ := strconv.Atoi(f[1])
		out = append(out, FileChange{Path: path, Added: a, Deleted: d, Status: status[path]})
	}
	return out, nil
}

func renamed(p string) string {
	if i, j := strings.Index(p, "{"), strings.Index(p, "}"); i >= 0 && j > i {
		_, to, _ := strings.Cut(p[i+1:j], " => ")
		return strings.ReplaceAll(p[:i]+to+p[j+1:], "//", "/")
	}
	_, to, _ := strings.Cut(p, " => ")
	return to
}

func (r *repo) FileAt(rev, path string) ([]byte, error) {
	return run(r.root, "show", rev+":"+path)
}

func (r *repo) CommitsTouching(paths []string, since string) (int, error) {
	out, err := run(r.root, append([]string{"rev-list", "--count", since + "..HEAD", "--"}, paths...)...)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}
