package main

import (
	"fmt"
	"os"

	"github.com/ngq/gorp/cmd/gorp/cmd"
)

// @title           Gorp Admin API
// @version         0.1.0
// @description     Gin-based framework demo.
// @BasePath        /
func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
