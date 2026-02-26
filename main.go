package main

import (
	"fmt"
	"os"

	"github.com/jlandells/mm-channel-heatmap/cmd"
)

var Version = "development"

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" {
			fmt.Printf("mm-channel-heatmap version %s\n", Version)
			os.Exit(0)
		}
	}
	os.Exit(cmd.Execute())
}
