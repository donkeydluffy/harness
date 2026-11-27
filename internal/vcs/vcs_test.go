package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestRepo_IF_007(t *testing.T) {
	dir := t.TempDir()
	if _, err := Open(dir); err != ErrNotRepo {
		t.Fatalf("want ErrNotRepo, got %v", err)
	}
	git(t, dir, "init", "-q", "-b", "main")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("1\n2\n"), 0o644)
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "first\n\nTask: T-001")
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("1\n3\n4\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("x\n"), 0o644)
	git(t, dir, "add", ".")

	r, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	staged, err := r.Staged()
	if err != nil || len(staged) != 2 || staged[0].Path != "a.txt" || staged[0].Added != 2 || staged[0].Deleted != 1 || staged[1].Status != "A" {
		t.Fatalf("staged: %+v %v", staged, err)
	}
	if b, err := r.FileAt("", "b.txt"); err != nil || string(b) != "x\n" {
		t.Fatalf("index file: %q %v", b, err)
	}
	git(t, dir, "commit", "-q", "-m", "second\n\nbody\n\nTask: T-002\nRefs: FR-001")
	cs, err := r.Commits("HEAD~1..HEAD")
	if err != nil || len(cs) != 1 || cs[0].Trailers["Task"][0] != "T-002" || cs[0].Trailers["Refs"][0] != "FR-001" {
		t.Fatalf("commits: %+v %v", cs, err)
	}
	all, _ := r.Commits("HEAD")
	if len(all) != 2 || all[0].Trailers["Task"][0] != "T-001" {
		t.Fatalf("all: %+v", all)
	}
	ch, err := r.Changes(all[0])
	if err != nil || len(ch) != 1 || ch[0].Added != 2 || ch[0].Status != "A" {
		t.Fatalf("root changes: %+v %v", ch, err)
	}
	if b, err := r.FileAt("HEAD~1", "a.txt"); err != nil || string(b) != "1\n2\n" {
		t.Fatalf("FileAt: %q %v", b, err)
	}
	if _, err := r.FileAt("HEAD~1", "b.txt"); err == nil {
		t.Fatal("want error for missing file")
	}
	if n, err := r.CommitsTouching([]string{"b.txt"}, "HEAD~1"); err != nil || n != 1 {
		t.Fatalf("touching: %d %v", n, err)
	}
	if h, err := r.LastCommit([]string{"b.txt"}); err != nil || h != all[1].Hash {
		t.Fatalf("last commit: %q %v", h, err)
	}
	if h, err := r.LastCommit([]string{"none.txt"}); err != nil || h != "" {
		t.Fatalf("no commit: %q %v", h, err)
	}
}

func TestParseTrailers_IF_007(t *testing.T) {
	if got := ParseTrailers("subject\n\nnot: a trailer block\nplain text"); len(got) != 0 {
		t.Fatalf("mixed paragraph is not trailers: %v", got)
	}
	if got := ParseTrailers("subject only"); len(got) != 0 {
		t.Fatal("subject only")
	}
	if got := ParseTrailers("s\n\nTask: T-1\nTask: T-2"); len(got["Task"]) != 2 {
		t.Fatalf("multi: %v", got)
	}
}
