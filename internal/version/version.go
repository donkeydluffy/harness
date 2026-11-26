// Package version 给出当前二进制的版本（MOD-18）。
package version

import (
	"regexp"
	"runtime/debug"
	"strings"
)

// injected 由发布构建通过 -ldflags "-X github.com/donkeydluffy/harness/internal/version.injected=<版本>" 注入。
var injected string

// Current 返回当前二进制的版本（AC-049-1）。
//
// harness:IF-037
func Current() string {
	info, _ := debug.ReadBuildInfo()
	return fromBuild(injected, info)
}

// fromBuild 按优先级取版本：注入的版本号；构建信息中的模块版本；vcs.revision 的前 12 位；dev。
// 模块版本与提交哈希都会在工作区有未提交改动时以 -dirty 结尾。
func fromBuild(injected string, info *debug.BuildInfo) string {
	if injected != "" {
		return strings.TrimPrefix(injected, "v")
	}
	if info == nil {
		return "dev"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		v = strings.TrimPrefix(v, "v")
		if base, ok := strings.CutSuffix(v, "+dirty"); ok {
			v = base + "-dirty"
		}
		return v
	}
	var rev string
	var modified bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if rev == "" {
		return "dev"
	}
	if len(rev) > 12 {
		rev = rev[:12]
	}
	if modified {
		rev += "-dirty"
	}
	return rev
}

var releaseRe = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// Release 判断版本是否为发布版本号：形如 0.2.0，不是 dev、提交哈希、伪版本，也不含 -dirty（AC-051-2）。
func Release(v string) bool { return releaseRe.MatchString(v) }
