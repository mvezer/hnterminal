package cli

import (
	"fmt"
	apiclient "hnterminal/internal/apiclient"
	config "hnterminal/internal/config"
	repository "hnterminal/internal/repository"
	"strings"
	"time"
)

type Cli struct {
	config *config.Config
	api    *apiclient.ApiClient
	repo   *repository.Repository
}

func New(config *config.Config) *Cli {
	return &Cli{config, nil, nil}
}

func (c *Cli) Init() {
	c.api = apiclient.New(nil)
	c.repo = repository.New(c.api)
}

func (c *Cli) Close() {
	c.repo.Close()
}

func (c *Cli) RenderStory(index int, story repository.Item) string {
	var rendered strings.Builder
	fmt.Fprintf(&rendered, "%d. %s\n", index, story.Title)
	fmt.Fprintf(&rendered, "  url: %s \n", story.Url)
	fmt.Fprintf(&rendered, "  date: %s | score: %d | comments: %d", time.Unix(int64(story.Time), 0).Format("2006-01-02 15:04:05"), story.Score, story.CommentsCount)
	return rendered.String()
}

func (c *Cli) Run() {
	switch c.config.Command {
	case "top":
		c.Init()
		topStories, err := c.api.GetTopStoryIds()
		if err != nil {
			fmt.Println(err)
		}
		storiesCount := len(topStories)
		if c.config.Count > 0 {
			storiesCount = c.config.Count
		}
		for idx := range storiesCount {
			story, err := c.repo.GetItem(topStories[idx])
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("--------------------------------\n%s\n", c.RenderStory(idx+1, *story))
		}
	default:

	}
}
