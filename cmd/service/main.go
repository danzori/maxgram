package main

import (
	"fmt"
	"os"

	"github.com/danzori/maxgram/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(".env"); err != nil {
		fmt.Fprintln(os.Stderr, "maxgram: ", err)
		os.Exit(1)
	}
}
