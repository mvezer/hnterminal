package main

import (
	config "hnterminal/internal/config"
	"hnterminal/internal/tui"
)

func main() {
	currentConfig := config.New()
	tui := tui.New(currentConfig)
	tui.Init()
	tui.Run()
	// if config.IsTUI() {
	// 	tui := ui.NewTui(currentConfig)
	// 	tui.Run()
	// } else {
	// 	cli := ui.NewCli(currentConfig)
	// 	cli.Run()
	// 	defer cli.Close()
	// }
}
