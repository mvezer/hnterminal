package tui

import (
	"hnterminal/utils"
	"log"
)

type layoutFunc func(Component)

type Layout int

const (
	HorizontalGrid Layout = iota
	VerticalGrid
	FixedWidth
	FixedHeight
)

var layoutFuncs = map[Layout]layoutFunc{
	HorizontalGrid: horizontalGridLayoutFunc,
	VerticalGrid:   verticalGridLayoutFunc,
	FixedWidth:     fixedWidthLayoutFunc,
	FixedHeight:    fixedHeightLayoutFunc,
}

// Filters and updates floating components
// then returns the non-floating components
func filterAndUpdateFloating(components []Component) []Component {
	nonfloating := make([]Component, 0)
	floating := make([]Component, 0)
	for _, c := range components {
		if c.Floating() {
			floating = append(floating, c)
		} else {
			nonfloating = append(nonfloating, c)
		}
	}
	for _, c := range floating {
		w := c.Width()
		if c.WidthPercent() > 0 {
			w = int(float64(c.Parent().Width()) / 100.0 * float64(c.WidthPercent()))
		}
		h := c.Height()
		if c.HeightPercent() > 0 {
			w = int(float64(c.Parent().Height()) / 100.0 * float64(c.HeightPercent()))
		}
		x := (c.Parent().Width() - w) / 2
		y := (c.Parent().Height() - h) / 2
		c.SetGeometry(x, y, w, h)
	}
	return nonfloating
}

func updateFloating(c Component) bool {
	w := c.FixedWidth()
	h := c.FixedHeight()
	x := 0
	y := 0
	if c.WidthPercent() > 0 {
		w = int(float64(c.Parent().Width()) / 100.0 * float64(c.WidthPercent()))
	}
	if c.HeightPercent() > 0 {
		w = int(float64(c.Parent().Height()) / 100.0 * float64(c.HeightPercent()))
	}
	if !c.IsRoot() {
		x = (c.Parent().Width() - w) / 2
		y = (c.Parent().Height() - h) / 2
	}
	c.SetGeometry(x, y, w, h)
	return w == -1 || h == -1
}

// retuns the calculated grid sizes
func calculateGrid(percentages []int, targetSize int) []int {
	sizes := make([]int, len(percentages))
	fixedCount := 0
	percentSum := 0.0
	fpercentages := make([]float64, len(percentages))
	for i, p := range percentages {
		if p != 0 {
			fixedCount++
			if (percentSum + float64(p)) > 100.0 {
				fpercentages[i] = (100.0 - float64(percentSum)) / 100
			} else {
				fpercentages[i] = float64(p) / 100.0
				percentSum += fpercentages[i]
			}
		}
	}

	fillerPercent := (1.0 - percentSum) / float64((len(percentages) - fixedCount))
	currentOffset := 0
	for i, p := range fpercentages {
		s := int(p * float64(targetSize))
		if p == 0 {
			s = int(fillerPercent * float64(targetSize))
		}
		sizes[i] = s
		currentOffset += s
	}
	sizeDifference := currentOffset - targetSize
	i := 0
	for sizeDifference != 0 {
		if i == len(percentages) {
			i = 0
		}
		delta := utils.Abs(sizeDifference) / sizeDifference
		if fixedCount == len(percentages) || percentages[i] == 0 {
			sizes[i] -= delta
			sizeDifference -= delta
		}
		i++
	}
	return sizes
}

// type horizontalGridLayoutFunc struct{}
// the height is given by the parent or the biggest height of the children
func horizontalGridLayoutFunc(component Component) {
	children := component.Children()
	childCount := 0
	for _, c := range children {
		if !c.Floating() {
			children[childCount] = c
			childCount++
		}
	}
	children = children[:childCount]
	percentages := make([]int, len(children))
	for i, c := range children {
		percentages[i] = c.WidthPercent()
	}
	if component.Height() == -1 { // we try to figure out the height
		if component.FixedHeight() != -1 { // if the fixed height is set, we use that
			component.SetHeight(component.FixedHeight() + component.Padding().Top + component.Padding().Bottom)
		} else { // ...otherwise we use the biggest child height
			biggestHeight := 0
			for _, c := range children {
				if c.Height() == -1 { // if any of the children has no height set, we bail and use -1 for the component height
					biggestHeight = -1
					break
				}
				if c.Height() > biggestHeight {
					biggestHeight = c.Height()
				}
			}
			if biggestHeight == -1 {
				component.SetHeight(-1)
			} else {
				component.SetHeight(biggestHeight + component.Padding().Top + component.Padding().Bottom)
			}
		}
	}
	xOffset := 0
	height := component.Height()
	for i, w := range calculateGrid(percentages, component.Width()-component.Padding().Left-component.Padding().Right) {
		children[i].SetGeometry(xOffset+component.Padding().Left, 0, w, height)
		xOffset += w
	}
}

func verticalGridLayoutFunc(component Component) {
}

// returns true if the component needs follow up
func fixedWidthLayoutFunc(component Component) {
	children := component.Children()
	childCount := 0
	for _, c := range children {
		if !c.Floating() {
			children[childCount] = c
			childCount++
		}
	}
	children = children[:childCount]
	for _, c := range children {
		c.SetWidth(component.Width() - component.Padding().Left - component.Padding().Right)
		c.AfterUpdate()
	}
	if component.FixedHeight() != -1 {
		component.SetHeight(component.FixedHeight() + component.Padding().Top + component.Padding().Bottom)
	} else {
		// add up the heights of the children
		yOffset := 0
		for _, c := range children {
			if c.Height() == -1 {
				yOffset = -1
				break
			}
			log.Printf("%d - %d", component.Id(), yOffset)
			c.SetX(component.Padding().Left)
			c.SetY(yOffset + component.Padding().Top)
			yOffset += c.Height()
			c.AfterUpdate()
		}
		if yOffset == -1 {
			component.SetHeight(-1)
		} else {
			component.SetHeight(yOffset + component.Padding().Top + component.Padding().Bottom)
		}
	}
}

func fixedHeightLayoutFunc(c Component) {
}

func ApplyLayout(component Component) bool {
	if component.Floating() {
		updateFloating(component)
	}
	layoutFuncs[component.Layout()](component)
	return component.Height() == -1 || component.Width() == -1
}
