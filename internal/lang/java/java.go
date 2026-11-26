// Package java 是 Java/Kotlin 语言包。
package java

import (
	"path"
	"regexp"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang/spec"
	"github.com/donkeydluffy/harness/internal/lang/textscan"
)

// Lang 实现 spec.Language。
type Lang struct{}

var syntax = textscan.Syntax{Line: []string{"//"}, Block: [][2]string{{"/*", "*/"}}, Quotes: []string{`"""`, `"`, "'"}}

func (Lang) Name() string                       { return "java" }
func (Lang) Extensions() []string               { return []string{".java", ".kt", ".kts"} }
func (Lang) Comments(src []byte) []spec.Comment { return syntax.Comments(src) }

var testFile = regexp.MustCompile(`(Test|Tests|IT|TestCase)\.(java|kt)$`)

// TestKind：*Test、*Tests、*IT 为测试文件；*IT（Maven failsafe 约定）、位于 integrationTest 等目录的为集成测试。
func (Lang) TestKind(p string, _ []byte) spec.TestKind {
	b := path.Base(p)
	it := strings.HasSuffix(strings.TrimSuffix(strings.TrimSuffix(b, ".java"), ".kt"), "IT")
	return textscan.Kind(p, testFile.MatchString(b), it)
}

var testMethod = regexp.MustCompile("@(?:Test|ParameterizedTest)\\b[^\\n]*\\n(?:\\s*@[^\\n]*\\n)*\\s*(?:public\\s+|internal\\s+)?(?:void\\s+|fun\\s+)(`[^`]+`|\\w+)\\s*\\(")

// TestNames 返回带 @Test 注解的方法名；Kotlin 反引号名称去掉反引号。
func (Lang) TestNames(src []byte) []string {
	var out []string
	for _, m := range testMethod.FindAllStringSubmatch(string(src), -1) {
		out = append(out, strings.Trim(m[1], "`"))
	}
	return out
}

// Pack 返回语言配置包（IF-034）。
func (Lang) Pack() spec.Pack {
	return spec.Pack{
		SourcePaths: []string{"src/**"},
		Format:      spec.Tool{Cmd: "google-java-format --dry-run --set-exit-if-changed $(find src -name '*.java')", Output: "files", Install: "brew install google-java-format（或从 https://github.com/google/google-java-format/releases 下载）"},
		Lint:        spec.Tool{Cmd: "pmd check -d src -R rulesets/java/quickstart.xml -f emacs --no-progress", Output: "lines", Install: "brew install pmd（或从 https://pmd.github.io 下载）"},
	}
}
