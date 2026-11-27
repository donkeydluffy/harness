package internal

import "os"

func F() {
	os.Remove("x") //nolint:errcheck
}
