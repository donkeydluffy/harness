//go:build !(darwin || freebsd || netbsd || openbsd || linux || windows)

package cli

// 无法判定是否为终端的平台一律视为非交互，批准只能在支持的平台上进行。
func isTerminal(fd uintptr) bool { return false }
