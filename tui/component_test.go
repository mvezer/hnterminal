package tui

import (
	"slices"
	"testing"
)

func Test_AddChild(t *testing.T) {
	parent := NewBox(HorizontalGrid)
	child := NewText("I'm the child", HorizontalGrid)
	child.SetGeometry(0, 0, 40, 2)
	parent.AddChild(&child)
	if len(parent.Children()) != 1 {
		t.Errorf("Expected to have 1 child and go %d instead", len(parent.Children()))
	}
	if parent.Children()[0].Id() != child.Id() {
		t.Errorf("Expected child to have id %d and got %d", child.Id(), parent.Children()[0].Id())
	}
}

func buildTestTree() []*BaseComponent {
	lastComponentId = -1
	root := NewBox(HorizontalGrid)
	child1 := NewBox(HorizontalGrid)
	child2 := NewBox(HorizontalGrid)
	root.AddChild(&child1)
	root.AddChild(&child2)
	child3 := NewBox(HorizontalGrid)
	child4 := NewBox(HorizontalGrid)
	child5 := NewBox(HorizontalGrid)
	child1.AddChild(&child3)
	child1.AddChild(&child4)
	child1.AddChild(&child5)
	child6 := NewBox(HorizontalGrid)
	child7 := NewBox(HorizontalGrid)
	child2.AddChild(&child6)
	child2.AddChild(&child7)

	return []*BaseComponent{&root, &child1, &child2, &child3, &child4, &child5, &child6, &child7}
}

func Test_Traverse(t *testing.T) {
	components := buildTestTree()
	root := components[0]
	rootTraverseIds := make([]int, 0)
	for c := range root.Traverse() {
		rootTraverseIds = append(rootTraverseIds, c.Id())
	}
	if len(rootTraverseIds) != 8 {
		t.Errorf("Expected to have 8 components and got %d", len(rootTraverseIds))
	}
	if slices.Compare(rootTraverseIds, []int{0, 1, 2, 3, 4, 5, 6, 7}) != 0 {
		t.Errorf("The order of ids is wrong, expected [0, 1, 2, 3, 4, 5, 6, 7] and got %v", rootTraverseIds)
	}

	child1 := components[1]
	child1TraverseIds := make([]int, 0)
	for c := range child1.Traverse() {
		child1TraverseIds = append(child1TraverseIds, c.Id())
	}
	if len(child1TraverseIds) != 4 {
		t.Errorf("Expected to have 4 components and got %d", len(child1TraverseIds))
	}
	if slices.Compare(child1TraverseIds, []int{1, 3, 4, 5}) != 0 {
		t.Errorf("The order of ids is wrong, expected [1, 3, 4, 5] and got %v", child1TraverseIds)
	}

	child2 := components[2]
	child2TraverseIds := make([]int, 0)
	for c := range child2.Traverse() {
		child2TraverseIds = append(child2TraverseIds, c.Id())
	}
	if len(child2TraverseIds) != 3 {
		t.Errorf("Expected to have 3 components and got %d", len(child2TraverseIds))
	}
	if slices.Compare(child2TraverseIds, []int{2, 6, 7}) != 0 {
		t.Errorf("The order of ids is wrong, expected [2, 6, 7] and got %v", child1TraverseIds)
	}

	child7 := components[7]
	child7TraverseIds := make([]int, 0)
	for c := range child7.Traverse() {
		child7TraverseIds = append(child7TraverseIds, c.Id())
	}
	if len(child7TraverseIds) != 1 {
		t.Errorf("Expected to have 1 components and got %d", len(child7TraverseIds))
	}

	if slices.Compare(child7TraverseIds, []int{7}) != 0 {
		t.Errorf("The order of ids is wrong, expected [2, 6, 7] and got %v", child1TraverseIds)
	}
}

func Test_Find(t *testing.T) {
	components := buildTestTree()
	components[4].SetDirty(true)
	expecedId := components[4].Id()
	actualId := components[4].Find(func(c *BaseComponent) bool {
		return c.dirty
	}).Id()
	if expecedId != actualId {
		t.Errorf("Expected to find component with id %d and got %d", expecedId, actualId)
	}
}
