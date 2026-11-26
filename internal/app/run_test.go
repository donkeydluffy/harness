package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_IF_038(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(dir, "fuzz"); err == nil {
		t.Error("不支持的检查应返回错误")
	}
	os.WriteFile(filepath.Join(dir, "harness.json"), []byte(`{"smoke": "echo boom; exit 4"}`), 0o644)
	rep, err := Run(dir, "smoke")
	if err != nil || len(rep.Violations) != 1 || rep.Violations[0].Rule != "smoke" || !strings.Contains(rep.Violations[0].Msg, "boom") {
		t.Fatalf("冒烟测试失败时应报告：%v %+v", err, rep)
	}
	if rep, err := Run(dir, "unit"); err != nil || len(rep.Violations) != 0 {
		t.Fatalf("unit：%v %+v", err, rep)
	}
}
