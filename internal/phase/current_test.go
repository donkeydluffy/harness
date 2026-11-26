package phase

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCurrent_IF_016(t *testing.T) {
	dir := t.TempDir()
	s := snapshot(t, dir)
	phases, err := Load(s)
	if err != nil {
		t.Fatal(err)
	}
	if cur, err := Current(s, phases); err != nil || cur.Name != phases[0].Name {
		t.Fatalf("未记录阶段时应为第一个阶段：%v %+v", err, cur)
	}
	os.WriteFile(filepath.Join(dir, "harness.lock"), []byte(`{"phase": "architecture"}`), 0o644)
	if cur, err := Current(snapshot(t, dir), phases); err != nil || cur.Name != "architecture" {
		t.Fatalf("应读锁文件中的阶段：%v %+v", err, cur)
	}
	os.WriteFile(filepath.Join(dir, "harness.lock"), []byte(`{"phase": "nope"}`), 0o644)
	if _, err := Current(snapshot(t, dir), phases); err == nil {
		t.Fatal("未知阶段应返回错误")
	}
}
