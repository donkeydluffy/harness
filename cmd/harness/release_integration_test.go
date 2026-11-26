package main

import (
	"crypto/sha256"
	"fmt"
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

// verifyScript 取出 release.yml 中 verify-begin 与 verify-end 之间的资产校验命令。
func verifyScript(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../../.github/workflows/release.yml")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	i, j := strings.Index(s, "# verify-begin"), strings.Index(s, "# verify-end")
	if i < 0 || j < i {
		t.Fatal("release.yml 缺少 verify-begin/verify-end 标记")
	}
	return s[i:j]
}

// releaseDir 生成 tag 对应的 6 个平台归档、checksums.txt 与 formula，返回资产目录与 formula 路径。
func releaseDir(t *testing.T, tag, formulaVersion string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	var sums strings.Builder
	for _, target := range releaseTargets {
		goos, goarch, _ := strings.Cut(target, "/")
		ext := "tar.gz"
		if goos == "windows" {
			ext = "zip"
		}
		name := "harness_" + strings.TrimPrefix(tag, "v") + "_" + goos + "_" + goarch + "." + ext
		data := []byte("binary " + target)
		os.WriteFile(filepath.Join(dir, name), data, 0o644)
		sums.WriteString(fmt.Sprintf("%x  %s\n", sha256.Sum256(data), name))
	}
	os.WriteFile(filepath.Join(dir, "checksums.txt"), []byte(sums.String()), 0o644)
	formula := filepath.Join(t.TempDir(), "harness.rb")
	os.WriteFile(formula, []byte("class Harness < Formula\n  version \""+formulaVersion+"\"\nend\n"), 0o644)
	return dir, formula
}

// runVerify 以 bash 运行校验命令，返回输出与是否成功。
func runVerify(t *testing.T, dir, tag, formula string) (string, bool) {
	t.Helper()
	out, err := exec.Command("bash", "-c", verifyScript(t), "verify", dir, tag, formula).CombinedOutput()
	return string(out), err == nil
}

func TestRelease_AC_N002_3_N1(t *testing.T) {
	cfg, err := os.ReadFile("../../.goreleaser.yaml")
	if err != nil || !strings.Contains(string(cfg), "goos: [darwin, linux, windows]") || !strings.Contains(string(cfg), "goarch: [amd64, arm64]") {
		t.Fatalf(".goreleaser.yaml 应构建 6 个平台：%v", err)
	}
	dir, formula := releaseDir(t, "v0.1.0", "0.1.0")
	if out, ok := runVerify(t, dir, "v0.1.0", formula); !ok || !strings.Contains(out, "6 个平台的归档与校验和一致") {
		t.Fatalf("完整的发布资产应通过校验：\n%s", out)
	}
	os.Remove(filepath.Join(dir, "harness_0.1.0_windows_arm64.zip"))
	if out, ok := runVerify(t, dir, "v0.1.0", formula); ok || !strings.Contains(out, "缺少 windows/arm64") {
		t.Fatalf("缺少平台时应失败：\n%s", out)
	}
}

func TestRelease_AC_N002_3_N2(t *testing.T) {
	dir, formula := releaseDir(t, "v1.2.3", "1.2.3")
	if out, ok := runVerify(t, dir, "v1.2.3", formula); !ok || !strings.Contains(out, "formula 版本 1.2.3") {
		t.Fatalf("formula 版本与 tag 一致时应通过：\n%s", out)
	}
	dir, formula = releaseDir(t, "v1.2.3", "1.2.2")
	if out, ok := runVerify(t, dir, "v1.2.3", formula); ok || !strings.Contains(out, "formula 版本与 tag v1.2.3 不一致") {
		t.Fatalf("formula 版本落后时应失败：\n%s", out)
	}
}

func TestRelease_AC_N002_3_E1(t *testing.T) {
	dir, formula := releaseDir(t, "v0.1.0", "0.1.0")
	os.WriteFile(filepath.Join(dir, "harness_0.1.0_linux_amd64.tar.gz"), []byte("被替换的文件"), 0o644)
	if out, ok := runVerify(t, dir, "v0.1.0", formula); ok || !strings.Contains(out, "校验和与文件不一致") {
		t.Fatalf("校验和不一致时应失败：\n%s", out)
	}
}
