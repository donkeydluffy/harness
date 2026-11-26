// Command harness 是 harness 的命令行入口。
package main

import (
	"os"

	"github.com/donkeydluffy/harness/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdout, os.Stderr))
}
