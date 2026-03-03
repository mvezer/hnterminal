package tui

import (
	"slices"
	"testing"
)

func testCalculateGrid(percentages []int, targetSize int, expected []int, t *testing.T) {
	actual := calculateGrid(percentages, targetSize)
	if slices.Compare(actual, expected) != 0 {
		t.Logf("Wrong grid size calcultaion result: expected %v but got %v", expected, actual)
	}
}

// func Test_TestCalculateGrid(t *testing.T) {
// 	testCalculateGrid([]int{50, 50}, 10, []int{5, 5}, t)
// 	testCalculateGrid([]int{10, 40, 0, 0}, 200, []int{20, 80, 50, 50}, t)
// 	testCalculateGrid([]int{0, 0, 0, 0}, 100, []int{25, 25, 25, 25}, t)
// 	testCalculateGrid([]int{50, 0, 0, 0}, 100, []int{50, 17, 17, 16}, t)
// }
