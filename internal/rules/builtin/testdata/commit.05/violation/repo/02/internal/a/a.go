package a

import (
	"fmt"

	"github.com/x/harness/internal/b"
)

func A() { fmt.Println(b.B(), 1) }
