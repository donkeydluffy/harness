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
	l.Phase = "planning"
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
	order := []string{`"approvals"`, `"artifact"`, `"at"`, `"by"`, `"hash"`, `"audit"`, `"items"`, `"FR-001"`, `"FR-002"`, `"phase"`, `"version"`}
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
	if strings.Contains(string(b), "units") || New().ByUnit() {
		t.Errorf("第 1 版锁文件不应按单位记录：%s", b)
	}
}
