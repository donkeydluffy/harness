package internal

import "os"

func F() {
	os.Remove("x") //nolint:errcheck // DEBT-001
}
