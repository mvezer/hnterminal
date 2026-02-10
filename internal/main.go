package main

import (
	"fmt"
	// apiclient "hnterminal/internal/apiclient"
	config "hnterminal/internal/config"
	// repository "hnterminal/internal/repository"
	cli "hnterminal/internal/cli"
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
	// api := apiclient.New(nil)
	// repo := repository.New(nil)
	// defer repo.Close()
	// itemIds, err := api.GetTopStoryIds()
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// storiesCount := 5
	// for id := range storiesCount {
	// 	item, err := repo.GetItem(itemIds[id])
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		continue
	// 	}
	// 	fmt.Printf("%s\n", "----------")
	// 	fmt.Printf("%s\n", item.Title)
	// 	fmt.Printf("%d\n", item.Score)
	// 	// fmt.Printf("%v\n", item.Kids)
	// }
}
