package tui

import (
	"hnterminal/internal/utils"
)

type layoutFunc func(*BaseComponent)

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
func filterAndUpdateFloating(components []*BaseComponent) []*BaseComponent {
	nonfloating := make([]*BaseComponent, 0)
	floating := make([]*BaseComponent, 0)
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

func updateFloating(c *BaseComponent) bool {
	w := c.fixedWidth
	h := c.fixedHeight
	x := 0
	y := 0
	if c.widthPercent > 0 {
		w = int(float64(c.parent.width) / 100.0 * float64(c.widthPercent))
	}
	if c.HeightPercent() > 0 {
		w = int(float64(c.parent.height) / 100.0 * float64(c.heightPercent))
	}
	if !c.IsRoot() {
		x = (c.parent.width - w) / 2
		y = (c.parent.height - h) / 2
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
func horizontalGridLayoutFunc(component *BaseComponent) {
	children := component.FilteredChildren(func(c *BaseComponent) bool { // filter out floating components
		return !c.floating
	})
	percentages := make([]int, len(children))
	for i, c := range children {
		percentages[i] = c.widthPercent
	}
	if component.height == -1 { // we try to figure out the height
		if component.fixedHeight != -1 { // if the fixed height is set, we use that
			component.height = component.fixedHeight + component.Padding().Top + component.Padding().Bottom
		} else { // ...otherwise we use the biggest child height
			biggestHeight := 0
			for _, c := range children {
				if c.height == -1 { // if any of the children has no height set, we bail and use -1 for the component height
					biggestHeight = -1
					break
				}
				if c.height > biggestHeight {
					biggestHeight = c.height
				}
			}
			if biggestHeight == -1 {
				component.height = -1
			} else {
				component.height = biggestHeight + component.Padding().Top + component.Padding().Bottom
			}
		}
	}
	xOffset := 0
	height := component.height
	for i, w := range calculateGrid(percentages, component.width-component.padding.Left-component.padding.Right) {
		children[i].x = xOffset + component.padding.Left
		children[i].y = 0
		children[i].width = w
		children[i].height = height
		xOffset += w
	}
}

func verticalGridLayoutFunc(component *BaseComponent) {
}

// returns true if the component needs follow up
func fixedWidthLayoutFunc(component *BaseComponent) {
	children := component.FilteredChildren(func(c *BaseComponent) bool { // filter out floating components
		return !c.floating
	})
	for _, c := range children {
		c.width = component.width - component.padding.Left - component.padding.Right
		c.kind.OnUpdate(c)
	}
	if component.height == -1 { // if not set, we try to use the fixedHeight
		if component.fixedHeight != -1 {
			component.height = component.fixedHeight + component.padding.Top + component.padding.Bottom
		} else {
			// add up the heights of the children
			yOffset := 0
			for _, c := range children {
				if c.height == -1 {
					yOffset = -1
					break
				}
				c.x = component.padding.Left
				c.y = yOffset + component.padding.Top
				yOffset += c.height
				c.kind.OnUpdate(c)
			}
			if yOffset == -1 {
				component.height = -1
			} else {
				component.height = yOffset + component.padding.Top + component.padding.Bottom
			}
		}
	}
}

func fixedHeightLayoutFunc(c *BaseComponent) {
}

func ApplyLayout(component *BaseComponent) bool {
	if component.floating {
		updateFloating(component)
	}
	layoutFuncs[component.layout](component)
	return component.height == -1 || component.width == -1
}
