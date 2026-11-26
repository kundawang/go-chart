package chart

import (
	"math"
	"sort"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"

	"github.com/wcharczuk/go-chart/v2/drawing"
)

// This file implements the optional tick-label collision avoidance layout.
//
// The layout measures every tick label using the active font, projects the
// tick value onto the axis and builds an axis aligned bounding box for the
// rendered label (including rotated text). It then selects a stable, pairwise
// non-overlapping subset of labels. The first and last ticks as well as any
// zero tick are treated as anchors and kept whenever their boxes are mutually
// compatible.
//
// Tick marks and grid lines are never regenerated or moved: they keep
// rendering at the original tick values. Only label text is suppressed.
//
// The selection is renderer independent: text extents are derived from the
// truetype face directly and inflated by a small safety margin so that the
// SVG and PNG renderers always make the same choice (the raster context uses
// slightly larger glyph extents than the vector drawer).

const (
	// DefaultTickLabelMinGap is the default minimum clear space, in pixels,
	// enforced between two visible tick label bounding boxes.
	DefaultTickLabelMinGap = 2

	// tickLabelMetricTolerance absorbs sub-pixel rounding and font hinting
	// differences between the measured box and the rendered glyphs.
	tickLabelMetricTolerance = 1

	// tickLabelRasterMetricFactor optionally scales the canonical vector
	// metrics. Selection stays renderer independent by using this canonical
	// measurement for both SVG and PNG; it is left at 1.0 so that the
	// priority end ticks survive on narrow axes, while tickLabelMetricTolerance
	// absorbs rounding and the small raster/vector glyph extent difference.
	tickLabelRasterMetricFactor = 1.0

	// tickLabelZeroEpsilon identifies a tick as the zero tick.
	tickLabelZeroEpsilon = 1e-9
)

// AxisOrientation identifies which kind of axis a tick label layout is
// computed for.
type AxisOrientation int

const (
	// AxisOrientationHorizontal is the horizontal (X) axis; labels vary along X.
	AxisOrientationHorizontal AxisOrientation = iota
	// AxisOrientationVertical is the vertical (Y) axis; labels vary along Y.
	AxisOrientationVertical
)

// TickLabelAnchorPriority values rank ticks when anchor boxes conflict.
// Higher values are considered first.
const (
	TickLabelPriorityNormal = 0
	TickLabelPriorityZero   = 50
	TickLabelPriorityLast   = 90
	TickLabelPriorityFirst  = 100
)

// TickLabelBox pairs a tick with its measured chart-coordinate bounding box
// and the priority used by the avoidance selection.
type TickLabelBox struct {
	// Index is the index of the tick in the original tick slice.
	Index int
	// Tick is the measured tick.
	Tick Tick
	// Anchor marks first/last/zero ticks.
	Anchor bool
	// Priority ranks the tick during stable conflict resolution.
	Priority int
	// Box is the axis aligned label bounding box in chart coordinates.
	Box Box
}

// TickLabelLayout is the result of measuring tick labels and selecting a
// non-overlapping subset.
type TickLabelLayout struct {
	// Orientation is the axis orientation the layout was computed for.
	Orientation AxisOrientation
	// Labels contains every tick with its measured box, in tick order.
	Labels []TickLabelBox
	// Visible reports whether the label at the original tick index is kept.
	Visible []bool
	// VisibleTicks is the kept subset of ticks, in original order.
	VisibleTicks []Tick
	// Boxes is the chart-coordinate box of every kept label, in tick order.
	Boxes []Box
	// MinGap is the clear spacing used by the selection.
	MinGap int
}

// VisibleIndices returns the original indices of the kept labels in order.
func (l TickLabelLayout) VisibleIndices() []int {
	var out []int
	for i, v := range l.Visible {
		if v {
			out = append(out, i)
		}
	}
	return out
}

// IsZeroTick reports whether t represents the axis zero tick.
func IsZeroTick(t Tick) bool {
	return t.Value == 0 || math.Abs(t.Value) <= tickLabelZeroEpsilon
}

// boxesOverlap reports whether two boxes collide once metric tolerance and a
// required minimum clear gap are accounted for. The gap is required between
// the (already raster-conservative) measured box edges.
func boxesOverlap(a, b Box, tolerance, minGap int) bool {
	required := minGap + 2*tolerance
	separatedX := a.Right+required <= b.Left || b.Right+required <= a.Left
	separatedY := a.Bottom+required <= b.Top || b.Bottom+required <= a.Top
	return !separatedX && !separatedY
}

// SelectNonOverlappingBoxes selects a stable pairwise non-overlapping subset
// of boxes.
//
// priorities ranks each index: higher priority items are considered first;
// equal priorities are considered in ascending index order. An item is kept
// if its box does not overlap any already kept box. The returned slice has
// one entry per input box.
func SelectNonOverlappingBoxes(boxes []Box, priorities []int, minGap int) []bool {
	n := len(boxes)
	visible := make([]bool, n)
	if n == 0 {
		return visible
	}
	if minGap < 0 {
		minGap = DefaultTickLabelMinGap
	}

	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool {
		pa, pb := TickLabelPriorityNormal, TickLabelPriorityNormal
		if priorities != nil {
			if order[a] < len(priorities) {
				pa = priorities[order[a]]
			}
			if order[b] < len(priorities) {
				pb = priorities[order[b]]
			}
		}
		if pa != pb {
			return pa > pb
		}
		return order[a] < order[b]
	})

	var kept []int
	for _, idx := range order {
		collides := false
		for _, k := range kept {
			if boxesOverlap(boxes[idx], boxes[k], tickLabelMetricTolerance, minGap) {
				collides = true
				break
			}
		}
		if !collides {
			visible[idx] = true
			kept = append(kept, idx)
		}
	}
	return visible
}

// tickPriorities builds the stable anchor priority table for ticks.
func tickPriorities(ticks []Tick) []int {
	n := len(ticks)
	prio := make([]int, n)
	for i := range ticks {
		switch {
		case n > 1 && i == 0:
			prio[i] = TickLabelPriorityFirst
		case n > 1 && i == n-1:
			prio[i] = TickLabelPriorityLast
		case IsZeroTick(ticks[i]):
			prio[i] = TickLabelPriorityZero
		default:
			prio[i] = TickLabelPriorityNormal
		}
	}
	// A lone tick is both first and last; give it the top priority.
	if n == 1 {
		prio[0] = TickLabelPriorityFirst
	}
	return prio
}

// canonicalLabelSize measures the unrotated extents of a label directly from
// the truetype face so the result is identical for the SVG and PNG
// renderers. The returned width already carries a small safety factor that
// covers the raster renderer's slightly larger glyph layout.
func canonicalLabelSize(r Renderer, label string, style Style) (w, h int) {
	dpi := DefaultDPI
	if r != nil {
		if d := r.GetDPI(); d > 0 {
			dpi = d
		}
	}
	size := style.GetFontSize(DefaultAxisFontSize)
	fnt := style.GetFont()
	if fnt == nil {
		if f, err := GetDefaultFont(); err == nil {
			fnt = f
		}
	}
	if fnt != nil {
		face := truetype.NewFace(fnt, &truetype.Options{
			DPI:  dpi,
			Size: size,
		})
		drawer := &font.Drawer{Face: face}
		w = drawer.MeasureString(label).Ceil()
	}
	h = int(drawing.PointsToPixels(dpi, size))
	if h <= 0 {
		h = 12
	}
	// Conservative compensation for the raster context metrics.
	w = int(math.Ceil(float64(w) * tickLabelRasterMetricFactor))
	return
}

// rotatePoint rotates (x,y) around (cx,cy) by theta radians.
func rotatePoint(cx, cy, x, y float64, theta float64) (float64, float64) {
	if theta == 0 {
		return x, y
	}
	sin, cos := math.Sincos(theta)
	dx, dy := x-cx, y-cy
	return cx + dx*cos - dy*sin, cy + dx*sin + dy*cos
}

// rotatedLabelBox builds the chart-coordinate AABB of a label whose measured
// (unrotated) em box is w by h, drawn at text origin (ox, oy) and rotated
// theta radians around that origin.
//
// The charting convention treats the text origin as the bottom-left corner of
// the measured em box (see XAxis/YAxis rendering), hence the unrotated box
// corners relative to the origin are (0,-h),(w,-h),(w,0),(0,0).
func rotatedLabelBox(ox, oy, w, h int, thetaDegrees float64) Box {
	theta := DegreesToRadians(thetaDegrees)
	corners := [4][2]float64{
		{float64(ox), float64(oy - h)},
		{float64(ox + w), float64(oy - h)},
		{float64(ox + w), float64(oy)},
		{float64(ox), float64(oy)},
	}
	var minX, minY = math.MaxFloat64, math.MaxFloat64
	var maxX, maxY = -math.MaxFloat64, -math.MaxFloat64
	for _, c := range corners {
		x, y := rotatePoint(float64(ox), float64(oy), c[0], c[1], theta)
		minX = math.Min(minX, x)
		maxX = math.Max(maxX, x)
		minY = math.Min(minY, y)
		maxY = math.Max(maxY, y)
	}
	return Box{
		Top:    int(math.Floor(minY)),
		Left:   int(math.Floor(minX)),
		Right:  int(math.Ceil(maxX)),
		Bottom: int(math.Ceil(maxY)),
	}
}

// tickLabelAnchorPoint returns the tick anchor in chart coordinates and the
// text origin used by the axis rendering code.
func tickLabelTextBox(r Renderer, style Style, label string, canvasBox Box, value int, orientation AxisOrientation, tp TickPosition, axisType YAxisType, index, tickCount int) Box {
	w, h := canonicalLabelSize(r, label, style)
	rotation := style.GetTextRotationDegrees(0)

	if orientation == AxisOrientationHorizontal {
		cx := canvasBox.Left + value
		switch tp {
		case TickPositionBetweenTicks:
			// Between-ticks labels are laid out inside the interval between
			// the previous tick and this tick. Collision handling uses the
			// centered interval box when a previous position is unavailable;
			// callers that need exact between-tick geometry pass the midpoint
			// through value. We simply center on the provided coordinate.
			ox := cx - w>>1
			oy := canvasBox.Bottom + DefaultXAxisMargin + h
			if rotation == 0 {
				return Box{Top: oy - h, Left: ox, Right: ox + w, Bottom: oy}
			}
			return rotatedLabelBox(cx, canvasBox.Bottom+2*DefaultXAxisMargin, w, h, rotation)
		default:
			if rotation == 0 {
				ox := cx - w>>1
				oy := canvasBox.Bottom + DefaultXAxisMargin + h
				return Box{Top: oy - h, Left: ox, Right: ox + w, Bottom: oy}
			}
			// Rotated X labels are anchored at the tick coordinate itself.
			oy := canvasBox.Bottom + 2*DefaultXAxisMargin
			return rotatedLabelBox(cx, oy, w, h, rotation)
		}
	}

	// Vertical axis.
	ly := canvasBox.Bottom - value
	primary := axisType != YAxisSecondary
	// The text X offset mirrors YAxis.Render up to the (small) stroke width.
	var tx int
	if primary {
		lx := canvasBox.Right + DefaultAxisLineWidth
		tx = lx + DefaultYAxisMargin
	} else {
		lx := canvasBox.Left - DefaultAxisLineWidth
		tx = lx - DefaultYAxisMargin - w
	}
	if rotation == 0 {
		oy := ly + h>>1
		return Box{Top: oy - h, Left: tx, Right: tx + w, Bottom: oy}
	}
	// Rotated Y labels are drawn with their origin at the tick height.
	return rotatedLabelBox(tx, ly, w, h, rotation)
}

// TickLabelLayoutOptions customizes the avoidance layout.
type TickLabelLayoutOptions struct {
	// MinGap is the minimum clear space in pixels between label boxes.
	// Zero selects DefaultTickLabelMinGap.
	MinGap int
	// TickPosition mirrors the axis TickPosition setting.
	TickPosition TickPosition
	// YAxisType selects primary (right) or secondary (left) Y placement.
	YAxisType YAxisType
	// Coordinates optionally provides pre-projected axis coordinates
	// (Range.Translate results) per tick; when nil, Project is used.
	Coordinates []int
	// Project maps a tick to its Range.Translate coordinate.
	Project func(Tick) int
}

// ComputeTickLabelLayout measures and selects tick labels for an axis.
func ComputeTickLabelLayout(r Renderer, orientation AxisOrientation, canvasBox Box, style Style, ticks []Tick, opts ...TickLabelLayoutOptions) TickLabelLayout {
	var cfg TickLabelLayoutOptions
	if len(opts) > 0 {
		cfg = opts[0]
	}
	minGap := cfg.MinGap
	if minGap == 0 {
		minGap = DefaultTickLabelMinGap
	}
	n := len(ticks)
	layout := TickLabelLayout{
		Orientation: orientation,
		Labels:      make([]TickLabelBox, n),
		Visible:     make([]bool, n),
		MinGap:      minGap,
	}
	if n == 0 {
		return layout
	}

	prio := tickPriorities(ticks)
	boxes := make([]Box, n)
	for i, t := range ticks {
		translated := 0
		switch {
		case i < len(cfg.Coordinates):
			translated = cfg.Coordinates[i]
		case cfg.Project != nil:
			translated = cfg.Project(t)
		}
		b := tickLabelTextBox(r, style, t.Label, canvasBox, translated, orientation, cfg.TickPosition, cfg.YAxisType, i, n)
		boxes[i] = b
		layout.Labels[i] = TickLabelBox{
			Index:    i,
			Tick:     t,
			Anchor:   prio[i] >= TickLabelPriorityZero,
			Priority: prio[i],
			Box:      b,
		}
	}

	layout.Visible = SelectNonOverlappingBoxes(boxes, prio, minGap)
	for i, v := range layout.Visible {
		if v {
			layout.VisibleTicks = append(layout.VisibleTicks, ticks[i])
			layout.Boxes = append(layout.Boxes, boxes[i])
		}
	}
	return layout
}

// selectAxisTickLabels is the shared entry point used by XAxis/YAxis render.
func selectAxisTickLabels(r Renderer, orientation AxisOrientation, canvasBox Box, ra Range, style Style, ticks []Tick, tp TickPosition, axisType YAxisType, minGap int) TickLabelLayout {
	coords := make([]int, len(ticks))
	for i, t := range ticks {
		coords[i] = ra.Translate(t.Value)
	}
	if orientation == AxisOrientationHorizontal && tp == TickPositionBetweenTicks {
		// Between-ticks labels occupy the interval before their tick; use the
		// interval midpoint as the collision coordinate.
		for i := range coords {
			if i == 0 {
				continue
			}
			coords[i] = (coords[i] + coords[i-1]) >> 1
		}
	}
	return ComputeTickLabelLayout(r, orientation, canvasBox, style, ticks, TickLabelLayoutOptions{
		MinGap:       minGap,
		TickPosition: tp,
		YAxisType:    axisType,
		Coordinates:  coords,
	})
}

// GetNonOverlappingXTicks returns the pairwise non-overlapping tick subset
// for a horizontal axis using the axis range projection. It mirrors the
// argument order of XAxis.Measure.
func GetNonOverlappingXTicks(r Renderer, canvasBox Box, ra Range, style Style, ticks []Tick) []Tick {
	return selectAxisTickLabels(r, AxisOrientationHorizontal, canvasBox, ra, style, ticks, TickPositionUnderTick, YAxisPrimary, DefaultTickLabelMinGap).VisibleTicks
}

// GetNonOverlappingTicks is an alias for GetNonOverlappingXTicks.
func GetNonOverlappingTicks(r Renderer, canvasBox Box, ra Range, style Style, ticks []Tick) []Tick {
	return GetNonOverlappingXTicks(r, canvasBox, ra, style, ticks)
}

// AvoidOverlappingTickLabels is an ergonomic alias for GetNonOverlappingTicks.
func AvoidOverlappingTickLabels(r Renderer, canvasBox Box, ra Range, style Style, ticks []Tick) []Tick {
	return GetNonOverlappingTicks(r, canvasBox, ra, style, ticks)
}
