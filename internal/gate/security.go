package gate

import (
	"math"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"github.com/donkeydluffy/harness/internal/workspace"
)

// sensitive 判断路径命中敏感路径且不在豁免列表中（gate.sensitive）。不含 "/" 的模式匹配文件名。
func sensitive(s *workspace.Snapshot, p string) bool {
	match := func(patterns []string) bool {
		for _, pat := range patterns {
			if strings.Contains(pat, "/") {
				if workspace.Match(pat, p) {
					return true
				}
			} else if ok, _ := path.Match(pat, path.Base(p)); ok {
				return true
			}
		}
		return false
	}
	return match(s.Config.SensitivePaths) && !match(s.Config.SensitiveAllow)
}

func denySensitive(p string) Decision {
	return deny("gate.sensitive", "%s 是敏感路径（harness.json 的 sensitivePaths），代理不能读写", p)
}

// secretRule 是一条密钥规则。
type secretRule struct {
	id   string
	desc string
	re   *regexp.Regexp
}

var secretRules = []secretRule{
	{"secret.aws", "AWS access key", regexp.MustCompile(`\b(AKIA|ASIA)[0-9A-Z]{16}\b`)},
	{"secret.pem", "PEM 私钥头", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{"secret.gcp", "Google API key", regexp.MustCompile(`\bAIza[0-9A-Za-z_-]{35}\b`)},
	{"secret.github", "GitHub token", regexp.MustCompile(`\bgh[pousr]_[0-9A-Za-z]{36,}\b`)},
}

var tokenRe = regexp.MustCompile(`[A-Za-z0-9+/_=-]{32,}`)

// findSecret 返回文本命中的密钥规则编号与说明；没有命中时 id 为空。
func findSecret(text string) (id, desc string) {
	for _, r := range secretRules {
		if r.re.MatchString(text) {
			return r.id, r.desc
		}
	}
	for _, tok := range tokenRe.FindAllString(text, -1) {
		if !pathLike(tok) && !wordIdentifier(tok) && highEntropy(tok) {
			return "secret.entropy", "高熵 token"
		}
	}
	return "", ""
}

// pathLike 判断 token 是以 /、./、~/ 开头的文件路径。路径里的用户名、目录名使熵偏高；
// base64 密钥可能含 /（甚至多个），所以只按开头判断。
func pathLike(tok string) bool {
	return strings.HasPrefix(tok, "/") || strings.HasPrefix(tok, "./") || strings.HasPrefix(tok, "~/")
}

// wordIdentifier 判断 token 是由英文单词拼成的标识符（如较长的测试名、常量名），这类 token 不按熵判定。
// 条件：只含字母、数字、下划线；按 _、大小写边界、字母数字边界切开后，不足 3 个字母的段至多 2 个、
// 数字段至多 3 个、像单词的段（首字母外全小写、至少 3 个字母、含元音）至少 3 个，
// 且非全大写段中的元音占字母的 30% 以上（英文约 40%，随机字母约 19%）。
// 用随机的 base62 token 估算，约百万分之十几会被误认作标识符。
func wordIdentifier(tok string) bool {
	segs := identSegments(tok)
	short, digits, words := 0, 0, 0
	letters, vowels := 0, 0
	for _, g := range segs {
		low := strings.ToLower(g)
		switch {
		case g[0] >= '0' && g[0] <= '9':
			digits++
			continue
		case len(g) < 3:
			short++
		case low[1:] == g[1:] && strings.ContainsAny(low, "aeiou"):
			words++
		}
		if strings.ToUpper(g) != g { // 全大写的缩写（HTTP、KB）常无元音，不计入元音比例
			letters += len(g)
			for _, c := range low {
				if strings.ContainsRune("aeiou", c) {
					vowels++
				}
			}
		}
	}
	return short <= 2 && digits <= 3 && words >= 3 && vowels*10 >= letters*3
}

// identSegments 按 _、小写到大写、连续大写到"大写+小写"、字母与数字之间切分；含其他字符时返回 nil。
func identSegments(tok string) []string {
	kind := func(c byte) int {
		switch {
		case c >= 'a' && c <= 'z':
			return 1
		case c >= 'A' && c <= 'Z':
			return 2
		case c >= '0' && c <= '9':
			return 3
		case c == '_':
			return 0
		}
		return -1
	}
	var out []string
	start := 0
	flush := func(end int) {
		if end > start {
			out = append(out, tok[start:end])
		}
		start = end
	}
	for i := 0; i < len(tok); i++ {
		k := kind(tok[i])
		switch {
		case k < 0:
			return nil
		case k == 0:
			flush(i)
			start = i + 1
			continue
		case i == start:
			continue
		}
		prev := kind(tok[i-1])
		switch {
		case (prev == 3) != (k == 3): // 字母与数字之间
			flush(i)
		case prev == 1 && k == 2: // aB
			flush(i)
		case prev == 2 && k == 1 && i-1 > start: // ABc：c 前的大写字母属于下一段
			flush(i - 1)
		}
	}
	flush(len(tok))
	return out
}

// highEntropy 判断 token 像随机密钥：同时含大小写字母与数字，字符熵不低于 4.3 比特
// （十六进制的提交哈希至多 4 比特，且通常只有小写）。
func highEntropy(tok string) bool {
	if !strings.ContainsAny(tok, "0123456789") ||
		strings.ToLower(tok) == tok || strings.ToUpper(tok) == tok {
		return false
	}
	freq := map[rune]float64{}
	for _, c := range tok {
		freq[c]++
	}
	var h float64
	n := float64(len(tok))
	for _, f := range freq {
		h -= f / n * math.Log2(f/n)
	}
	return h >= 4.3
}

// addedText 返回写入后内容中原文件没有的行，只检查新增内容，避免已有内容反复触发。
func addedText(root, p, content string) string {
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(p)))
	if err != nil {
		return content
	}
	old := map[string]bool{}
	for _, l := range strings.Split(string(b), "\n") {
		old[l] = true
	}
	var out []string
	for _, l := range strings.Split(content, "\n") {
		if !old[l] {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func denySecret(where, id, desc string) Decision {
	return deny(id, "%s 含疑似密钥（%s，规则 %s）；密钥应放在环境变量或密钥管理服务中", where, desc, id)
}

// dangerous 返回命令命中的危险命令模式（gate.command）。模式中的词按顺序出现即命中，
// 词之间可以夹其他参数，"*" 匹配任意内容。
func dangerous(s *workspace.Snapshot, cmd string) string {
	norm := " " + strings.Join(strings.Fields(regexp.MustCompile(`\|+|;|&&`).ReplaceAllStringFunc(cmd, func(m string) string {
		return " " + m + " "
	})), " ") + " "
	for _, pat := range s.Config.DangerousCommands {
		var parts []string
		for _, tok := range strings.Fields(pat) {
			if tok == "*" {
				parts = append(parts, `\S+`)
			} else {
				parts = append(parts, regexp.QuoteMeta(tok))
			}
		}
		if len(parts) == 0 {
			continue
		}
		re, err := regexp.Compile(`\s` + strings.Join(parts, `\s(?:\S+\s)*?`) + `\s`)
		if err == nil && re.MatchString(norm) {
			return pat
		}
	}
	return ""
}

// recursiveRemoveOutside 判断 rm 递归删除仓库外路径或无法静态判定的路径。
func recursiveRemoveOutside(root string, wd cwd, call *syntax.CallExpr) (string, bool) {
	args := words(call.Args)
	if len(args) == 0 || filepath.Base(args[0]) != "rm" {
		return "", false
	}
	recursive := false
	for _, a := range args[1:] {
		if a == "--recursive" || (strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") && strings.ContainsAny(a, "rR")) {
			recursive = true
		}
	}
	if !recursive {
		return "", false
	}
	for i, a := range args[1:] {
		switch {
		case a == "":
			// 通配符取其所在目录判定；含变量或命令替换时无法判定
			lit := call.Args[i+1].Lit()
			if lit == "" {
				return "无法静态判定的路径", true
			}
			if !path.IsAbs(lit) && wd.unknown {
				return "无法静态判定的目录中的 " + lit, true
			}
			if r := Resolve(root, wd.dir, path.Dir(lit)); r == ".." || strings.HasPrefix(r, "../") {
				return lit, true
			}
		case strings.HasPrefix(a, "-"):
		case strings.HasPrefix(a, "~"):
			return a, true
		case !path.IsAbs(a) && wd.unknown:
			return "无法静态判定的目录中的 " + a, true
		default:
			if r := Resolve(root, wd.dir, a); r == ".." || strings.HasPrefix(r, "../") || r == "." {
				return a, true
			}
		}
	}
	return "", false
}
