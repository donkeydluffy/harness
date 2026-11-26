package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// releaseTargets 是发布的 6 个平台（AC-N002-1）。
var releaseTargets = []string{"darwin/arm64", "darwin/amd64", "linux/amd64", "linux/arm64", "windows/amd64", "windows/arm64"}

// crossBuild 为 target（GOOS/GOARCH）交叉编译 harness，返回产物路径；失败时返回编译输出。
func crossBuild(t *testing.T, target string) (string, string, error) {
	t.Helper()
	goos, goarch, _ := strings.Cut(target, "/")
	out := filepath.Join(t.TempDir(), "harness-"+goos+"-"+goarch)
	if goos == "windows" {
		out += ".exe"
	}
	cmd := exec.Command("go", "build", "-trimpath", "-o", out, ".")
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
	b, err := cmd.CombinedOutput()
	return out, string(b), err
}

func TestBuild_AC_N002_1_N1(t *testing.T) {
	if testing.Short() {
		t.Skip("交叉编译 6 个平台较慢")
	}
	for _, target := range releaseTargets {
		out, log, err := crossBuild(t, target)
		if err != nil {
			t.Fatalf("%s 交叉编译失败：%v\n%s", target, err, log)
		}
		if fi, err := os.Stat(out); err != nil || fi.Size() == 0 {
			t.Fatalf("%s 没有产出二进制", target)
		}
	}
}

func TestBuild_AC_N002_1_E1(t *testing.T) {
	// 不存在的平台编译失败，crossBuild 报告错误，发布流程据此失败。
	if _, log, err := crossBuild(t, "linux/nosuch"); err == nil || !strings.Contains(log, "nosuch") {
		t.Fatalf("不存在的平台应编译失败：%v %s", err, log)
	}
}
