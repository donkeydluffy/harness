package lock

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeHash_IF_005(t *testing.T) {
	a := "### FR-001 a  \r\n- Status: Active\n\n\n\nx\t\n\n"
	b := "### FR-001 a\n- Status: Active\n\nx"
	if Normalize(a) != b {
		t.Fatalf("normalize: %q", Normalize(a))
	}
	if Hash(a) != Hash(b) || !strings.HasPrefix(Hash(a), "sha256:") || Hash(a) == Hash(b+"y") {
		t.Fatal("hash")
	}
}

func TestReadWrite_IF_006(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, File)
	l, err := Read(p)
	if err != nil || l.Version != 1 || len(l.Items) != 0 || Exists(p) {
		t.Fatalf("missing file: %+v %v", l, err)
	}
	l.Items["FR-002"] = "sha256:b"
	l.Items["FR-001"] = "sha256:a"
	l.Legacy = append(l.Legacy, Approval{Artifact: "docs/PRD.md", Hash: "sha256:x", By: "dev", At: time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC)})
	if err := l.Write(p); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(p)
	got, err := Read(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := got.Write(p); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(p)
	if !bytes.Equal(first, second) {
		t.Fatalf("not byte-stable:\n%s\n%s", first, second)
	}
	s := string(first)
	order := []string{`"approvals"`, `"artifact"`, `"at"`, `"by"`, `"hash"`, `"audit"`, `"items"`, `"FR-001"`, `"FR-002"`, `"version"`}
	last := -1
	for _, k := range order {
		i := strings.Index(s, k)
		if i <= last {
			t.Fatalf("key %s out of order:\n%s", k, s)
		}
		last = i
	}
	if !strings.HasSuffix(s, "}\n") {
		t.Fatal("trailing newline")
	}
}

func TestApproved_IF_006(t *testing.T) {
	l := New()
	l.Legacy = []Approval{{Artifact: "a", Hash: "h1"}, {Artifact: "a", Hash: "h2"}}
	if !l.Approved("a", "h2") || l.Approved("a", "h1") || l.Approved("b", "h2") {
		t.Fatal("approved")
	}
	if a := l.LastApproval("a"); a == nil || a.Hash != "h2" {
		t.Fatalf("最近一次批准：%+v", a)
	}
	if l.LastApproval("b") != nil {
		t.Fatal("没有批准记录时返回 nil")
	}
}

// 字段缺失的锁文件读出后各集合不为 nil；路径不可读或不可写时返回错误。
func TestReadEmptyFields_IF_006(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, File)
	os.WriteFile(p, []byte(`{"items": null, "legacy": null, "audit": null}`), 0o644)
	l, err := Read(p)
	if err != nil || l.Items == nil || l.Legacy == nil || l.Audit == nil {
		t.Fatalf("%v %+v", err, l)
	}
	if _, err := Read(dir); err == nil {
		t.Fatal("路径是目录时应返回错误")
	}
	if err := New().Write(filepath.Join(dir, "missing", File)); err == nil {
		t.Fatal("目录不存在时写入应返回错误")
	}
}

func TestReadMalformed_IF_006(t *testing.T) {
	p := filepath.Join(t.TempDir(), File)
	os.WriteFile(p, []byte("{"), 0o644)
	if _, err := Read(p); err == nil {
		t.Fatal("want error")
	}
}

// 第 2 版锁文件按单位记录批准：读写往返不变，第 1 版不写出 units。
func TestUnits_IF_006(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, File)
	if err := os.WriteFile(p, []byte(`{"version": 2}`), 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := Read(p)
	if err != nil || !l.ByUnit() || l.Units == nil {
		t.Fatalf("第 2 版应按单位记录：%v %+v", err, l)
	}
	at := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	l.Units["MOD-02"] = UnitApproval{Hash: "sha256:a", Text: "### MOD-02 app", By: "x", At: at, Upstream: map[string]string{"FR-002": "sha256:b"}}
	if err := l.Write(p); err != nil {
		t.Fatal(err)
	}
	again, err := Read(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := again.Units["MOD-02"]; got.Upstream["FR-002"] != "sha256:b" || !got.At.Equal(at) || got.Text != "### MOD-02 app" {
		t.Errorf("往返后不一致：%+v", got)
	}
	b, _ := New().Marshal()
	if !strings.Contains(string(b), `"version": 2`) || !New().ByUnit() || strings.Contains(string(b), "approvals") {
		t.Errorf("新建的锁文件应为第 2 版，且不写出按文件批准：%s", b)
	}
}

// Release 1 的 phase 字段读到时忽略，写出后消失；没有 version 的文件按第 1 版读取（AC-048-1）。
func TestPhaseDropped_IF_006(t *testing.T) {
	p := filepath.Join(t.TempDir(), File)
	os.WriteFile(p, []byte(`{"phase": "implementation", "items": {}}`), 0o644)
	l, err := Read(p)
	if err != nil || l.ByUnit() {
		t.Fatalf("应按第 1 版读取：%+v %v", l, err)
	}
	if err := l.Write(p); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(p); strings.Contains(string(b), "phase") || strings.Contains(string(b), "approvals") {
		t.Fatalf("写出后不应再有 phase 与空的按文件批准：%s", b)
	}
}
