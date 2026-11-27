package toolrun

import "testing"

// harness init 把默认命令写入 tools：与配置包默认命令相同的命令沿用配置包的输出格式（AC-014-3、AC-014-4）。
func TestLint_AC_014_4_N1_DefaultInTools(t *testing.T) {
	s := snapshot(t, map[string]string{
		"harness.json":    `{"languages": ["go"], "sourcePaths": ["internal/**"], "tools": {"go": {"format": "gofmt  -l .", "lint": "go vet ./..."}}}`,
		"go.mod":          "module demo\n\ngo 1.21\n",
		"internal/a/b.go": "package a\nfunc  F() {}\n",
	})
	rep, err := Lint(s)
	if err != nil {
		t.Fatal(err)
	}
	for _, is := range rep.Issues {
		if is.File == "internal/a/b.go" && is.Tool == "gofmt" {
			return
		}
	}
	t.Fatalf("tools 中写的是默认的 gofmt -l，未格式化的文件应列入报告：%+v", rep.Issues)
}
