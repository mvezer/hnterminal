package ui

import "hnterminal/tui"

type StoryList struct {
	root    *tui.BaseComponent
	stories []*Story
}

func NewStoryList() StoryList {
	root := tui.NewBox(tui.FixedWidth)
	stories := make([]*Story, 0)
	storyList := StoryList{&root, stories}
	return storyList
}

func (sl *StoryList) HasStory(storyToFind *Story) bool {
	for _, s := range sl.stories {
		if s.root.Id() == storyToFind.root.Id() {
			return true
		}
	}
	return false
}

func (sl *StoryList) Add(s *Story) {
	if sl.HasStory(s) {
		return
	}
	sl.root.AddChild(s.Root())
	sl.stories = append(sl.stories, s)
}

func (sl *StoryList) Root() *tui.BaseComponent {
	return sl.root
}
