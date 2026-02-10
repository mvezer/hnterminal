package main

import (
	"fmt"
	cli "hnterminal/internal/cli"
	config "hnterminal/internal/config"
)

func main() {
	config.ParseArgs()
	if config.IsTUI() {
		fmt.Println("TUI mode not implemented yet")
	} else {
		cli := cli.New(config.GetConfig())
		cli.Run()
		defer cli.Close()
	}
}
