package main

import (
	"fmt"
	"os"

	"github.com/danzori/maxgram/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(".env"); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "maxgram:", err)

		os.Exit(1)
	}
}
