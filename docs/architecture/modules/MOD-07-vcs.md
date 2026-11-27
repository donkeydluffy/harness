# MOD-07 vcs 详细设计

- Status: Approved

## 接口

### IF-007 Repo
- Module: MOD-07
- Covers: FR-010, FR-013, FR-030, FR-034

```go
// IF-007
type Repo interface {
	Root() string
	Commits(rangeSpec string) ([]Commit, error)       // 如 "origin/main..HEAD"
	Changes(c Commit) ([]FileChange, error)           // 路径、增删行数
	Staged() ([]FileChange, error)                    // pre-commit 使用
	FileAt(rev, path string) ([]byte, error)
	CommitsTouching(paths []string, since string) (int, error)
	LastCommit(paths []string) (string, error)        // 最后一次改动 paths 的提交；没有时为空串（AC-030-4）
}
type Commit struct{ Hash, Message string; Trailers map[string][]string }
type FileChange struct{ Path string; Added, Deleted int; Status string }

func Open(dir string) (Repo, error) // 非 git 仓库返回 ErrNotRepo
```

## 内部设计

调用系统 `git`（`git log --format`、`git diff --numstat`、`git interpret-trailers --parse`），不引入 git 库。
