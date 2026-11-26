package chart

import (
	"math"

	"github.com/wcharczuk/go-chart/v2/drawing"
)

var (
	// Draw contains helpers for drawing common objects.
	Draw = &draw{}
)

type draw struct{}

// LineSeries draws a line series with a renderer.
func (d draw) LineSeries(r Renderer, canvasBox Box, xrange, yrange Range, style Style, vs ValuesProvider) {
	if vs.Len() == 0 {
		return
	}

	cb := canvasBox.Bottom
	cl := canvasBox.Left

	v0x, v0y := vs.GetValues(0)
	x0 := cl + xrange.Translate(v0x)
	y0 := cb - yrange.Translate(v0y)

	yv0 := yrange.Translate(0)

	var vx, vy float64
	var x, y int

	if style.ShouldDrawStroke() && style.ShouldDrawFill() {
		style.GetFillOptions().WriteDrawingOptionsToRenderer(r)
		r.MoveTo(x0, y0)
		for i := 1; i < vs.Len(); i++ {
			vx, vy = vs.GetValues(i)
			x = cl + xrange.Translate(vx)
			y = cb - yrange.Translate(vy)
			r.LineTo(x, y)
		}
		r.LineTo(x, MinInt(cb, cb-yv0))
		r.LineTo(x0, MinInt(cb, cb-yv0))
		r.LineTo(x0, y0)
		r.Fill()
	}

	if style.ShouldDrawStroke() {
		style.GetStrokeOptions().WriteDrawingOptionsToRenderer(r)

		r.MoveTo(x0, y0)
		for i := 1; i < vs.Len(); i++ {
			vx, vy = vs.GetValues(i)
			x = cl + xrange.Translate(vx)
			y = cb - yrange.Translate(vy)
			r.LineTo(x, y)
		}
		r.Stroke()
	}

	if style.ShouldDrawDot() {
		defaultDotWidth := style.GetDotWidth()

		style.GetDotOptions().WriteDrawingOptionsToRenderer(r)
		for i := 0; i < vs.Len(); i++ {
			vx, vy = vs.GetValues(i)
			x = cl + xrange.Translate(vx)
			y = cb - yrange.Translate(vy)

			dotWidth := defaultDotWidth
			if style.DotWidthProvider != nil {
				dotWidth = style.DotWidthProvider(xrange, yrange, i, vx, vy)
			}

			if style.DotColorProvider != nil {
				dotColor := style.DotColorProvider(xrange, yrange, i, vx, vy)

				r.SetFillColor(dotColor)
				r.SetStrokeColor(dotColor)
			}

			r.Circle(dotWidth, x, y)
			r.FillStroke()
		}
	}
}

// BoundedSeries draws a series that implements BoundedValuesProvider.
func (d draw) BoundedSeries(r Renderer, canvasBox Box, xrange, yrange Range, style Style, bbs BoundedValuesProvider, drawOffsetIndexes ...int) {
	drawOffsetIndex := 0
	if len(drawOffsetIndexes) > 0 {
		drawOffsetIndex = drawOffsetIndexes[0]
	}

	cb := canvasBox.Bottom
	cl := canvasBox.Left

	v0x, v0y1, v0y2 := bbs.GetBoundedValues(0)
	x0 := cl + xrange.Translate(v0x)
	y0 := cb - yrange.Translate(v0y1)

	var vx, vy1, vy2 float64
	var x, y int

	xvalues := make([]float64, bbs.Len())
	xvalues[0] = v0x
	y2values := make([]float64, bbs.Len())
	y2values[0] = v0y2

	style.GetFillAndStrokeOptions().WriteToRenderer(r)
	r.MoveTo(x0, y0)
	for i := 1; i < bbs.Len(); i++ {
		vx, vy1, vy2 = bbs.GetBoundedValues(i)

		xvalues[i] = vx
		y2values[i] = vy2

		x = cl + xrange.Translate(vx)
		y = cb - yrange.Translate(vy1)
		if i > drawOffsetIndex {
			r.LineTo(x, y)
		} else {
			r.MoveTo(x, y)
		}
	}
	y = cb - yrange.Translate(vy2)
	r.LineTo(x, y)
	for i := bbs.Len() - 1; i >= drawOffsetIndex; i-- {
		vx, vy2 = xvalues[i], y2values[i]
		x = cl + xrange.Translate(vx)
		y = cb - yrange.Translate(vy2)
		r.LineTo(x, y)
	}
	r.Close()
	r.FillStroke()
}

// ohlcCandleColorer is implemented by OHLC series that color their
// candles according to the direction of the period.
type ohlcCandleColorer interface {
	CandleColor(index int) drawing.Color
}

// OHLCSeries draws an open/high/low/close series as candlesticks.
//
// Each candle consists of a vertical wick spanning the high-low range and
// a body spanning the open-close range. Candles where the open equals the
// close (doji) render as a horizontal line at that value.
//
// When candleWidth is zero the body width is derived from the pixel
// distance to the nearest neighboring point, so non-uniform spacing on the
// x axis yields proportionally sized bodies. A single, neighborless candle
// falls back to a fraction of the canvas width.
func (d draw) OHLCSeries(r Renderer, canvasBox Box, xrange, yrange Range, style Style, vs OHLCValuesProvider, candleWidths ...int) {
	if vs.Len() == 0 {
		return
	}

	explicitWidth := 0
	if len(candleWidths) > 0 {
		explicitWidth = candleWidths[0]
	}
	autoWidth := explicitWidth <= 0

	colorer, hasColors := vs.(ohlcCandleColorer)

	cb := canvasBox.Bottom
	cl := canvasBox.Left

	// A single point produces a zero-delta range, which Translate maps to
	// NaN. Center such degenerate candles in the canvas instead so a
	// one-point series still renders a recognizable candle.
	degenerateX := xrange.GetDelta() == 0
	degenerateY := yrange.GetDelta() == 0
	midX := canvasBox.Left + (canvasBox.Width() >> 1)
	midY := canvasBox.Top + (canvasBox.Height() >> 1)

	// precompute the pixel x coordinate of every point so neighbor
	// distances are measured in the same space the bodies are drawn in.
	px := make([]int, vs.Len())
	for i := 0; i < vs.Len(); i++ {
		vx, _, _, _, _ := vs.GetOHLCValues(i)
		if degenerateX {
			px[i] = midX
		} else {
			px[i] = cl + xrange.Translate(vx)
		}
	}

	singleWidth := MaxInt(MinimumCandleWidth, int(math.Floor(float64(canvasBox.Width())*DefaultSingleCandleWidthRatio)))

	for index := 0; index < vs.Len(); index++ {
		_, open, high, low, close := vs.GetOHLCValues(index)

		width := explicitWidth
		if autoWidth {
			width = singleWidth
			if vs.Len() > 1 {
				gap := math.MaxFloat64
				if index > 0 {
					gap = math.Min(gap, math.Abs(float64(px[index]-px[index-1])))
				}
				if index < vs.Len()-1 {
					gap = math.Min(gap, math.Abs(float64(px[index+1]-px[index])))
				}
				width = MaxInt(MinimumCandleWidth, int(math.Floor(gap*DefaultCandleWidthRatio)))
			}
		}
		halfWidth := width >> 1

		x := px[index]
		var highY, lowY, openY, closeY int
		if degenerateY {
			highY = midY
			lowY = midY
			openY = midY
			closeY = midY
		} else {
			highY = cb - yrange.Translate(high)
			lowY = cb - yrange.Translate(low)
			openY = cb - yrange.Translate(open)
			closeY = cb - yrange.Translate(close)
		}

		// edge candles are centered on the boundary data point, so clamp
		// every pixel coordinate to the canvas to keep wicks, bodies and
		// ticks from spilling into the axis area.
		highY = MinInt(canvasBox.Bottom, MaxInt(canvasBox.Top, highY))
		lowY = MinInt(canvasBox.Bottom, MaxInt(canvasBox.Top, lowY))
		openY = MinInt(canvasBox.Bottom, MaxInt(canvasBox.Top, openY))
		closeY = MinInt(canvasBox.Bottom, MaxInt(canvasBox.Top, closeY))
		bodyLeft := MaxInt(canvasBox.Left, x-halfWidth)
		bodyRight := MinInt(canvasBox.Right, x+halfWidth)

		color := style.GetStrokeColor()
		if hasColors {
			color = colorer.CandleColor(index)
		}

		// wick: vertical line from the low to the high.
		r.SetStrokeColor(color)
		r.SetStrokeWidth(style.GetStrokeWidth(DefaultSeriesLineWidth))
		r.SetStrokeDashArray(nil)
		r.MoveTo(x, highY)
		r.LineTo(x, lowY)
		r.Stroke()

		if open == close {
			// doji: no body, draw a horizontal tick instead.
			r.MoveTo(bodyLeft, openY)
			r.LineTo(bodyRight, openY)
			r.Stroke()
			continue
		}

		body := Box{
			Top:    MinInt(openY, closeY),
			Left:   bodyLeft,
			Right:  bodyRight,
			Bottom: MaxInt(openY, closeY),
		}
		r.SetFillColor(color)
		d.Box(r, body, Style{
			StrokeColor: color,
			StrokeWidth: style.GetStrokeWidth(DefaultSeriesLineWidth),
			FillColor:   color,
		})
	}
}

// OHLCLegendSwatch draws a small candlestick swatch for use in legends
// between (lx, ly) and (lx+swatchWidth, ly), centered vertically on ly.
func (d draw) OHLCLegendSwatch(r Renderer, upColor, downColor drawing.Color, lx, ly, swatchWidth int) {
	halfHeight := 4
	quarterWidth := swatchWidth >> 2
	midX := lx + (swatchWidth >> 1)

	// up candle on the left half.
	r.SetStrokeColor(upColor)
	r.SetStrokeWidth(DefaultSeriesLineWidth)
	r.MoveTo(lx+quarterWidth, ly-halfHeight-1)
	r.LineTo(lx+quarterWidth, ly+halfHeight+1)
	r.Stroke()

	r.SetFillColor(upColor)
	d.Box(r, Box{
		Top:    ly - halfHeight,
		Left:   lx,
		Right:  lx + (swatchWidth >> 1),
		Bottom: ly,
	}, Style{
		StrokeColor: upColor,
		StrokeWidth: DefaultSeriesLineWidth,
		FillColor:   upColor,
	})

	// down candle on the right half.
	r.SetStrokeColor(downColor)
	r.MoveTo(midX+quarterWidth, ly-halfHeight-1)
	r.LineTo(midX+quarterWidth, ly+halfHeight+1)
	r.Stroke()

	r.SetFillColor(downColor)
	d.Box(r, Box{
		Top:    ly,
		Left:   midX,
		Right:  lx + swatchWidth,
		Bottom: ly + halfHeight,
	}, Style{
		StrokeColor: downColor,
		StrokeWidth: DefaultSeriesLineWidth,
		FillColor:   downColor,
	})
}

// HistogramSeries draws a value provider as boxes from 0.
func (d draw) HistogramSeries(r Renderer, canvasBox Box, xrange, yrange Range, style Style, vs ValuesProvider, barWidths ...int) {
	if vs.Len() == 0 {
		return
	}

	//calculate bar width?
	seriesLength := vs.Len()
	barWidth := int(math.Floor(float64(xrange.GetDomain()) / float64(seriesLength)))
	if len(barWidths) > 0 {
		barWidth = barWidths[0]
	}

	cb := canvasBox.Bottom
	cl := canvasBox.Left

	//foreach datapoint, draw a box.
	for index := 0; index < seriesLength; index++ {
		vx, vy := vs.GetValues(index)
		y0 := yrange.Translate(0)
		x := cl + xrange.Translate(vx)
		y := yrange.Translate(vy)

		d.Box(r, Box{
			Top:    cb - y0,
			Left:   x - (barWidth >> 1),
			Right:  x + (barWidth >> 1),
			Bottom: cb - y,
		}, style)
	}
}

// MeasureAnnotation measures how big an annotation would be.
func (d draw) MeasureAnnotation(r Renderer, canvasBox Box, style Style, lx, ly int, label string) Box {
	style.WriteToRenderer(r)
	defer r.ResetStyle()

	textBox := r.MeasureText(label)
	textWidth := textBox.Width()
	textHeight := textBox.Height()
	halfTextHeight := textHeight >> 1

	pt := style.Padding.GetTop(DefaultAnnotationPadding.Top)
	pl := style.Padding.GetLeft(DefaultAnnotationPadding.Left)
	pr := style.Padding.GetRight(DefaultAnnotationPadding.Right)
	pb := style.Padding.GetBottom(DefaultAnnotationPadding.Bottom)

	strokeWidth := style.GetStrokeWidth()

	top := ly - (pt + halfTextHeight)
	right := lx + pl + pr + textWidth + DefaultAnnotationDeltaWidth + int(strokeWidth)
	bottom := ly + (pb + halfTextHeight)

	return Box{
		Top:    top,
		Left:   lx,
		Right:  right,
		Bottom: bottom,
	}
}

// Annotation draws an anotation with a renderer.
func (d draw) Annotation(r Renderer, canvasBox Box, style Style, lx, ly int, label string) {
	style.GetTextOptions().WriteToRenderer(r)
	defer r.ResetStyle()

	textBox := r.MeasureText(label)
	textWidth := textBox.Width()
	halfTextHeight := textBox.Height() >> 1

	style.GetFillAndStrokeOptions().WriteToRenderer(r)

	pt := style.Padding.GetTop(DefaultAnnotationPadding.Top)
	pl := style.Padding.GetLeft(DefaultAnnotationPadding.Left)
	pr := style.Padding.GetRight(DefaultAnnotationPadding.Right)
	pb := style.Padding.GetBottom(DefaultAnnotationPadding.Bottom)

	textX := lx + pl + DefaultAnnotationDeltaWidth
	textY := ly + halfTextHeight

	ltx := lx + DefaultAnnotationDeltaWidth
	lty := ly - (pt + halfTextHeight)

	rtx := lx + pl + pr + textWidth + DefaultAnnotationDeltaWidth
	rty := ly - (pt + halfTextHeight)

	rbx := lx + pl + pr + textWidth + DefaultAnnotationDeltaWidth
	rby := ly + (pb + halfTextHeight)

	lbx := lx + DefaultAnnotationDeltaWidth
	lby := ly + (pb + halfTextHeight)

	r.MoveTo(lx, ly)
	r.LineTo(ltx, lty)
	r.LineTo(rtx, rty)
	r.LineTo(rbx, rby)
	r.LineTo(lbx, lby)
	r.LineTo(lx, ly)
	r.Close()
	r.FillStroke()

	style.GetTextOptions().WriteToRenderer(r)
	r.Text(label, textX, textY)
}

// Box draws a box with a given style.
func (d draw) Box(r Renderer, b Box, s Style) {
	s.GetFillAndStrokeOptions().WriteToRenderer(r)
	defer r.ResetStyle()

	r.MoveTo(b.Left, b.Top)
	r.LineTo(b.Right, b.Top)
	r.LineTo(b.Right, b.Bottom)
	r.LineTo(b.Left, b.Bottom)
	r.LineTo(b.Left, b.Top)
	r.FillStroke()
}

func (d draw) BoxRotated(r Renderer, b Box, thetaDegrees float64, s Style) {
	d.BoxCorners(r, b.Corners().Rotate(thetaDegrees), s)
}

func (d draw) BoxCorners(r Renderer, bc BoxCorners, s Style) {
	s.GetFillAndStrokeOptions().WriteToRenderer(r)
	defer r.ResetStyle()

	r.MoveTo(bc.TopLeft.X, bc.TopLeft.Y)
	r.LineTo(bc.TopRight.X, bc.TopRight.Y)
	r.LineTo(bc.BottomRight.X, bc.BottomRight.Y)
	r.LineTo(bc.BottomLeft.X, bc.BottomLeft.Y)
	r.Close()
	r.FillStroke()
}

// DrawText draws text with a given style.
func (d draw) Text(r Renderer, text string, x, y int, style Style) {
	style.GetTextOptions().WriteToRenderer(r)
	defer r.ResetStyle()

	r.Text(text, x, y)
}

func (d draw) MeasureText(r Renderer, text string, style Style) Box {
	style.GetTextOptions().WriteToRenderer(r)
	defer r.ResetStyle()

	return r.MeasureText(text)
}

// TextWithin draws the text within a given box.
func (d draw) TextWithin(r Renderer, text string, box Box, style Style) {
	style.GetTextOptions().WriteToRenderer(r)
	defer r.ResetStyle()

	lines := Text.WrapFit(r, text, box.Width(), style)
	linesBox := Text.MeasureLines(r, lines, style)

	y := box.Top

	switch style.GetTextVerticalAlign() {
	case TextVerticalAlignBottom, TextVerticalAlignBaseline: // i have to build better baseline handling into measure text
		y = y - linesBox.Height()
	case TextVerticalAlignMiddle:
		y = y + (box.Height() >> 1) - (linesBox.Height() >> 1)
	case TextVerticalAlignMiddleBaseline:
		y = y + (box.Height() >> 1) - linesBox.Height()
	}

	var tx, ty int
	for _, line := range lines {
		lineBox := r.MeasureText(line)
		switch style.GetTextHorizontalAlign() {
		case TextHorizontalAlignCenter:
			tx = box.Left + ((box.Width() - lineBox.Width()) >> 1)
		case TextHorizontalAlignRight:
			tx = box.Right - lineBox.Width()
		default:
			tx = box.Left
		}
		if style.TextRotationDegrees == 0 {
			ty = y + lineBox.Height()
		} else {
			ty = y
		}

		r.Text(line, tx, ty)
		y += lineBox.Height() + style.GetTextLineSpacing()
	}
}
