package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/donkeydluffy/harness/internal/lang"
)

// InterfacesFile 是提交到仓库的接口快照文件（formats.md 第 8 节）。
const InterfacesFile = "docs/architecture/interfaces.txt"

// Interfaces 读取接口快照：接口编号 -> 签名。文件不存在时 ok 为 false。
func (s *Snapshot) Interfaces() (sigs map[string]string, ok bool) {
	b, err := os.ReadFile(filepath.Join(s.Root, InterfacesFile))
	if err != nil {
		return nil, false
	}
	return ParseInterfaces(b), true
}

// ParseInterfaces 解析接口快照：每段第一行为接口编号，第二行为签名，段之间空一行。
func ParseInterfaces(b []byte) map[string]string {
	out := map[string]string{}
	for _, block := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n\n") {
		lines := strings.SplitN(strings.TrimSpace(block), "\n", 2)
		if len(lines) == 2 && strings.HasPrefix(lines[0], "IF-") {
			out[strings.TrimSpace(lines[0])] = strings.TrimSpace(lines[1])
		}
	}
	return out
}

// FormatInterfaces 按接口编号排序输出接口快照；同一输入的输出逐字节一致（AC-052-2/B1）。
func FormatInterfaces(sigs map[string]string) []byte {
	ids := make([]string, 0, len(sigs))
	for id := range sigs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var b strings.Builder
	for i, id := range ids {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(id + "\n" + sigs[id] + "\n")
	}
	return []byte(b.String())
}

// Anchors 返回 sourcePaths 内非测试源文件中的全部接口锚点，按文件与行排序。快照须以 WithSources 加载。
func (s *Snapshot) Anchors() []lang.Decl {
	var out []lang.Decl
	for _, f := range s.Sources {
		a, ok := lang.AnalyzerFor(f.Path)
		if !ok || f.Kind != lang.NotTest {
			continue
		}
		src, err := f.Content()
		if err != nil {
			continue
		}
		out = append(out, a.Anchors(f.Path, src)...)
	}
	return out
}

// ExtractInterfaces 从锚点提取接口快照：每个接口取第一个带签名的锚点（多个锚点由规则 if.01 报告）。
func (s *Snapshot) ExtractInterfaces() map[string]string {
	sigs := map[string]string{}
	for _, d := range s.Anchors() {
		if _, dup := sigs[d.ID]; !dup && d.Signature != "" {
			sigs[d.ID] = d.Signature
		}
	}
	return sigs
}
