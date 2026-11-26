//go:build darwin || freebsd || netbsd || openbsd

package cli

import (
	"syscall"
	"unsafe"
)

func isTerminal(fd uintptr) bool {
	var t syscall.Termios
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCGETA, uintptr(unsafe.Pointer(&t)))
	return e == 0
}
