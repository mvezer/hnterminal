package main

import (
	"hnterminal/internal/api"
	config "hnterminal/internal/config"
	"hnterminal/internal/ui"
	"hnterminal/tui"
)

func main() {
	apiClient := api.NewApiClient(nil)
	currentConfig := config.New()
	repository := api.NewRepository(apiClient, currentConfig)
	tui := tui.New(currentConfig)
	ui := ui.New(tui, repository, apiClient)
	ui.Init()
	ui.Run()
}
