package gate

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// maxCopy 以上的文件只记录元数据，命令后发生变化时只报告不恢复。
const maxCopy = 1 << 20

// 不纳入快照的目录：版本库、门禁状态、依赖缓存。
var snapSkip = map[string]bool{".git": true, ".harness": true, "node_modules": true}

type fileState struct {
	Size  int64     `json:"size"`
	MTime time.Time `json:"mtime"`
	Copy  bool      `json:"copy"` // 有内容副本
}

// Change 是命令前后的一处文件变化。
type Change struct {
	Path     string
	Action   string // created / modified / deleted
	Restored bool   // 已恢复为命令前状态
	Reason   string // 未恢复的原因（被允许或文件过大）
	Saved    string // 恢复前另存的位置（仓库相对路径），没有另存时为空（AC-056-2）
}

// restoredDir 存放命令后恢复前另存的文件内容：门禁运行时目录，已被忽略，代理不能写入（ADR-018）。
const restoredDir = ".harness/session/restored"

// stamp 便于测试固定另存目录的时间部分。
var stamp = func() string { return time.Now().UTC().Format("20060102T150405.000000000Z") }

// save 把仓库文件 rel 的当前内容另存到 restoredDir/<at>/<rel>，返回另存位置；失败时为空。
func save(root, at, rel string) string {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return ""
	}
	dst := restoredDir + "/" + at + "/" + rel
	abs := filepath.Join(root, filepath.FromSlash(dst))
	if os.MkdirAll(filepath.Dir(abs), 0o755) != nil || os.WriteFile(abs, b, 0o644) != nil {
		return ""
	}
	return dst
}

func snapDir(root, session string) string {
	return filepath.Join(root, ".harness", "snap", filepath.Base(session))
}

// Before 在命令执行前记录仓库文件的 (路径, 大小, mtime) 与小文件内容副本，存于 .harness/snap/<会话>（IF-018）。
//
// harness:IF-018
func Before(root, sessionID string) error {
	dir := snapDir(root, sessionID)
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	states, err := scan(root)
	if err != nil {
		return err
	}
	for rel, st := range states {
		if st.Size > maxCopy {
			continue
		}
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, "files", rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return err
		}
		st.Copy = true
		states[rel] = st
	}
	b, _ := json.Marshal(states)
	return os.WriteFile(filepath.Join(dir, "state.json"), b, 0o644)
}

// After 比对命令前后状态；不被 allowed 允许的改动恢复为命令前内容，新建文件删除（IF-018）。
// 恢复或删除前把文件当前内容另存到 .harness/session/restored/<时间>/<路径>（AC-056-2）。
// 没有命令前快照时返回 nil。
func After(root, sessionID string, allowed func(path string) bool) ([]Change, error) {
	dir := snapDir(root, sessionID)
	b, err := os.ReadFile(filepath.Join(dir, "state.json"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var before map[string]fileState
	if err := json.Unmarshal(b, &before); err != nil {
		return nil, err
	}
	now, err := scan(root)
	if err != nil {
		return nil, err
	}
	var changes []Change
	for rel, st := range now {
		old, ok := before[rel]
		switch {
		case !ok:
			changes = append(changes, Change{Path: rel, Action: "created"})
		case old.Size != st.Size || !old.MTime.Equal(st.MTime):
			changes = append(changes, Change{Path: rel, Action: "modified"})
		}
	}
	for rel := range before {
		if _, ok := now[rel]; !ok {
			changes = append(changes, Change{Path: rel, Action: "deleted"})
		}
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	at := stamp()
	for i := range changes {
		c := &changes[i]
		if allowed(c.Path) {
			c.Reason = "允许"
			continue
		}
		abs := filepath.Join(root, filepath.FromSlash(c.Path))
		if c.Action == "created" {
			c.Saved = save(root, at, c.Path)
			c.Restored = os.Remove(abs) == nil
			continue
		}
		if !before[c.Path].Copy {
			c.Reason = "文件超过 1 MiB，未保存副本，无法恢复"
			continue
		}
		orig, err := os.ReadFile(filepath.Join(dir, "files", filepath.FromSlash(c.Path)))
		if err == nil {
			c.Saved = save(root, at, c.Path)
			os.MkdirAll(filepath.Dir(abs), 0o755)
			err = os.WriteFile(abs, orig, 0o644)
		}
		c.Restored = err == nil
	}
	return changes, os.RemoveAll(dir)
}

func scan(root string) (map[string]fileState, error) {
	out := map[string]fileState{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != root && snapSkip[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		out[filepath.ToSlash(rel)] = fileState{Size: fi.Size(), MTime: fi.ModTime()}
		return nil
	})
	return out, err
}
