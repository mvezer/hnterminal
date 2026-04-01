package ui

import (
	"hnterminal/internal/api"
	"hnterminal/tui"
	"hnterminal/utils"
)

type Story struct {
	id         int
	root       *tui.BaseComponent
	title      *TextBox
	url        *TextBox
	infoBox    *tui.BaseComponent
	score      *tui.BaseComponent
	time       *tui.BaseComponent
	comments   *tui.BaseComponent
	item       *api.Item
	repository *api.Repository
	t          *tui.TUI
}

func NewStory(storyId int, repository *api.Repository) Story {
	root := tui.NewBox(tui.FixedWidth)
	root.Kind().(*tui.Box).SetBorder(tui.Border{Bottom: true, Top: false, Left: false, Right: false})
	root.Kind().(*tui.Box).SetBorderStyle(tui.BorderStyleSingle)
	root.SetPadding(tui.Padding{Top: 0, Left: 2, Right: 2, Bottom: 1})
	title := NewTextBox("Title")
	// title.text.OnUpdate()
	url := NewTextBox("URL")
	// url.text.OnUpdate()
	root.AddChild(title.Root())
	root.AddChild(url.Root())
	story := Story{root: &root, title: &title, url: &url, id: storyId, repository: repository}
	story.Update()
	return story
}

func (s *Story) Root() *tui.BaseComponent {
	return s.root
}

func (s *Story) Load() {
	item, err := s.repository.GetItem(s.id)
	if err != nil {
		utils.HandleError(err, utils.ErrorSeverityError)
	}
	s.item = item
	s.Update()
}

func (s *Story) Update() {
	if s.item == nil { // not loaded yet
		s.title.SetText("Loading...")
		s.url.SetText("...")
	} else {
		s.title.SetText(s.item.Title)
		s.title.text.OnUpdate()
		s.url.SetText(s.item.Url)
		s.title.text.OnUpdate()
	}
}
