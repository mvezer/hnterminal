package ui

import (
	"hnterminal/internal/api"
	"hnterminal/tui"
	"hnterminal/utils"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

type UI struct {
	t          *tui.TUI
	repository *api.Repository
	apiClient  *api.ApiClient
	storyList  *StoryList
	comments   *tui.BaseComponent
}

func New(t *tui.TUI, repository *api.Repository, apiClient *api.ApiClient) *UI {
	storyList := NewStoryList()
	comments := tui.NewBox(tui.FixedWidth)
	return &UI{t, repository, apiClient, &storyList, &comments}
}

func (u *UI) StoryList() *StoryList {
	return u.storyList
}

func (u *UI) Comments() *tui.BaseComponent {
	return u.comments
}

func (u *UI) Init() {
	ids, err := u.apiClient.GetTopStoryIds()
	if err != nil {
		utils.HandleError(err, utils.ErrorSeverityFatal)
	}
	u.t.Root().SetStyle(tcell.StyleDefault.Foreground(color.White).Background(color.DarkGray))
	u.t.Root().SetLayout(tui.HorizontalGrid)
	u.storyList.Root().SetWidthPercent(50)
	for i := range min(5, len(ids)) {
		story := NewStory(ids[i], u.repository)
		u.storyList.Add(&story)
		story.Load()
	}
	u.t.Root().AddChild(u.storyList.Root())
	u.t.Root().AddChild(u.comments)

}

func (u *UI) Run() {
	u.t.Screen().Clear()
	u.t.Screen().Show()
	defer u.t.Quit()
	for {
		u.t.Draw()
		ev := <-u.t.Screen().EventQ()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			u.t.Draw()
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyCtrlC, tcell.KeyEscape:
				u.t.Quit()
				return
			case tcell.KeyRight:
				u.storyList.Root().SetWidthPercent(min(u.storyList.Root().WidthPercent()+1, 100))
			case tcell.KeyLeft:
				u.storyList.Root().SetWidthPercent(max(u.storyList.Root().WidthPercent()-1, 5))
				// case tcell.KeyLeft:
				// 	box1.SetMinWidth(box1.MinWidth() - 1)
				// 	box1.SetMaxWidth(box1.MaxWidth() - 1)
				// 	t.UpdateGeometry(nil)
				// case tcell.KeyEnter:
				// 	toggleFloatBox(t)
				// case tcell.KeyEnter:
				// case tcell.KeyRune:
				// 	fmt.Printf("Rune: %c\n", ev.())
				// default:
				// 	log.Printf("Key: %s\n", ev.Name())
			}
		}
	}
}
