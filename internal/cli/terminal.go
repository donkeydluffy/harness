package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// stdTerminal 是 app.Terminal 的标准输入输出实现。
type stdTerminal struct {
	in  *bufio.Reader
	out io.Writer
}

func newTerminal(out io.Writer) *stdTerminal {
	return &stdTerminal{in: bufio.NewReader(os.Stdin), out: out}
}

// IsInteractive 判断标准输入是否为终端；管道、文件、/dev/null、代理执行的命令都不是。
func (t *stdTerminal) IsInteractive() bool {
	return isTerminal(os.Stdin.Fd())
}

func (t *stdTerminal) Confirm(prompt string) bool {
	fmt.Fprint(t.out, prompt)
	line, _ := t.in.ReadString('\n')
	a := strings.ToLower(strings.TrimSpace(line))
	return a == "y" || a == "yes"
}

// Identity 返回 git 配置的 "姓名 <邮箱>"；未配置时用操作系统用户名。
func (t *stdTerminal) Identity() string {
	name := gitConfig("user.name")
	email := gitConfig("user.email")
	switch {
	case name != "" && email != "":
		return name + " <" + email + ">"
	case name != "":
		return name
	}
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	return "unknown"
}

func gitConfig(key string) string {
	out, err := exec.Command("git", "config", "--get", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
