package main

import (
	cli "hnterminal/internal/cli"
	config "hnterminal/internal/config"
	tui "hnterminal/internal/tui"
)

func main() {
	currentConfig := config.New()
	if config.IsTUI() {
		tui := tui.New(currentConfig)
		tui.Run()
	} else {
		cli := cli.New(currentConfig)
		cli.Run()
		defer cli.Close()
	}
}
