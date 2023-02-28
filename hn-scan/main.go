package main

import (
	"fmt"
	"os"

	"github.com/VladMinzatu/go-projects/hn-scan/cmd"
)

func main() {
	err := cmd.NewCmdApp().Run(os.Stderr, os.Args[1:])

	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
