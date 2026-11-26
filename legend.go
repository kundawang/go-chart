package chart

import (
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// legendSeries collects the information needed to render one legend row.
type legendSeries struct {
	label string
	style Style
	ohlc  OHLCLegendProvider
}

// collectLegendSeries returns the visible, non-annotation series rendered
// in a legend along with their effective style.
func collectLegendSeries(c *Chart) []legendSeries {
	var output []legendSeries
	for index, s := range c.Series {
		if s.GetStyle().Hidden {
			continue
		}
		if _, isAnnotationSeries := s.(AnnotationSeries); isAnnotationSeries {
			continue
		}
		entry := legendSeries{
			label: s.GetName(),
			style: s.GetStyle().InheritFrom(c.styleDefaultsSeries(index)),
		}
		if lp, isOHLC := s.(OHLCLegendProvider); isOHLC {
			entry.ohlc = lp
		}
		output = append(output, entry)
	}
	return output
}

// drawLegendSwatch draws either a candle swatch or a plain line for a
// legend row between (lx, ly) and (lx+width, ly).
func drawLegendSwatch(r Renderer, entry legendSeries, lx, ly, width int) {
	if entry.ohlc != nil {
		Draw.OHLCLegendSwatch(r, entry.ohlc.GetUpColor(), entry.ohlc.GetDownColor(), lx, ly, width)
		return
	}
	r.SetStrokeColor(entry.style.GetStrokeColor())
	r.SetStrokeWidth(entry.style.GetStrokeWidth())
	r.SetStrokeDashArray(entry.style.GetStrokeDashArray())
	r.MoveTo(lx, ly)
	r.LineTo(lx+width, ly)
	r.Stroke()
}

// Legend returns a legend renderable function.
func Legend(c *Chart, userDefaults ...Style) Renderable {
	return func(r Renderer, cb Box, chartDefaults Style) {
		legendDefaults := Style{
			FillColor:   drawing.ColorWhite,
			FontColor:   DefaultTextColor,
			FontSize:    8.0,
			StrokeColor: DefaultAxisColor,
			StrokeWidth: DefaultAxisLineWidth,
		}

		var legendStyle Style
		if len(userDefaults) > 0 {
			legendStyle = userDefaults[0].InheritFrom(chartDefaults.InheritFrom(legendDefaults))
		} else {
			legendStyle = chartDefaults.InheritFrom(legendDefaults)
		}

		// DEFAULTS
		legendPadding := Box{
			Top:    5,
			Left:   5,
			Right:  5,
			Bottom: 5,
		}
		lineTextGap := 5
		lineLengthMinimum := 25

		entries := collectLegendSeries(c)

		legend := Box{
			Top:  cb.Top,
			Left: cb.Left,
			// bottom and right will be sized by the legend content + relevant padding.
		}

		legendContent := Box{
			Top:    legend.Top + legendPadding.Top,
			Left:   legend.Left + legendPadding.Left,
			Right:  legend.Left + legendPadding.Left,
			Bottom: legend.Top + legendPadding.Top,
		}

		legendStyle.GetTextOptions().WriteToRenderer(r)

		// measure
		labelCount := 0
		for x := 0; x < len(entries); x++ {
			if len(entries[x].label) > 0 {
				tb := r.MeasureText(entries[x].label)
				if labelCount > 0 {
					legendContent.Bottom += DefaultMinimumTickVerticalSpacing
				}
				legendContent.Bottom += tb.Height()
				right := legendContent.Left + tb.Width() + lineTextGap + lineLengthMinimum
				legendContent.Right = MaxInt(legendContent.Right, right)
				labelCount++
			}
		}

		legend = legend.Grow(legendContent)
		legend.Right = legendContent.Right + legendPadding.Right
		legend.Bottom = legendContent.Bottom + legendPadding.Bottom

		Draw.Box(r, legend, legendStyle)

		legendStyle.GetTextOptions().WriteToRenderer(r)

		ycursor := legendContent.Top
		tx := legendContent.Left
		legendCount := 0
		var label string
		for x := 0; x < len(entries); x++ {
			label = entries[x].label
			if len(label) > 0 {
				if legendCount > 0 {
					ycursor += DefaultMinimumTickVerticalSpacing
				}

				tb := r.MeasureText(label)

				ty := ycursor + tb.Height()
				r.Text(label, tx, ty)

				th2 := tb.Height() >> 1

				lx := tx + tb.Width() + lineTextGap
				ly := ty - th2
				lx2 := legendContent.Right - legendPadding.Right

				drawLegendSwatch(r, entries[x], lx, ly, lx2-lx)

				ycursor += tb.Height()
				legendCount++
			}
		}
	}
}

// LegendThin is a legend that doesn't obscure the chart area.
func LegendThin(c *Chart, userDefaults ...Style) Renderable {
	return func(r Renderer, cb Box, chartDefaults Style) {
		legendDefaults := Style{
			FillColor:   drawing.ColorWhite,
			FontColor:   DefaultTextColor,
			FontSize:    8.0,
			StrokeColor: DefaultAxisColor,
			StrokeWidth: DefaultAxisLineWidth,
			Padding: Box{
				Top:    2,
				Left:   7,
				Right:  7,
				Bottom: 5,
			},
		}

		var legendStyle Style
		if len(userDefaults) > 0 {
			legendStyle = userDefaults[0].InheritFrom(chartDefaults.InheritFrom(legendDefaults))
		} else {
			legendStyle = chartDefaults.InheritFrom(legendDefaults)
		}

		r.SetFont(legendStyle.GetFont())
		r.SetFontColor(legendStyle.GetFontColor())
		r.SetFontSize(legendStyle.GetFontSize())

		entries := collectLegendSeries(c)

		var textHeight int
		var textWidth int
		var textBox Box
		for x := 0; x < len(entries); x++ {
			if len(entries[x].label) > 0 {
				textBox = r.MeasureText(entries[x].label)
				textHeight = MaxInt(textBox.Height(), textHeight)
				textWidth = MaxInt(textBox.Width(), textWidth)
			}
		}

		legendBoxHeight := textHeight + legendStyle.Padding.Top + legendStyle.Padding.Bottom
		chartPadding := cb.Top
		legendYMargin := (chartPadding - legendBoxHeight) >> 1

		legendBox := Box{
			Left:   cb.Left,
			Right:  cb.Right,
			Top:    legendYMargin,
			Bottom: legendYMargin + legendBoxHeight,
		}

		Draw.Box(r, legendBox, legendDefaults)

		r.SetFont(legendStyle.GetFont())
		r.SetFontColor(legendStyle.GetFontColor())
		r.SetFontSize(legendStyle.GetFontSize())

		lineTextGap := 5
		lineLengthMinimum := 25

		tx := legendBox.Left + legendStyle.Padding.Left
		ty := legendYMargin + legendStyle.Padding.Top + textHeight
		var label string
		var lx, ly int
		th2 := textHeight >> 1
		for index := range entries {
			label = entries[index].label
			if len(label) > 0 {
				textBox = r.MeasureText(label)
				r.Text(label, tx, ty)

				lx = tx + textBox.Width() + lineTextGap
				ly = ty - th2

				drawLegendSwatch(r, entries[index], lx, ly, lineLengthMinimum)

				tx += textBox.Width() + DefaultMinimumTickHorizontalSpacing + lineTextGap + lineLengthMinimum
			}
		}
	}
}

// LegendLeft is a legend that is designed for longer series lists.
func LegendLeft(c *Chart, userDefaults ...Style) Renderable {
	return func(r Renderer, cb Box, chartDefaults Style) {
		legendDefaults := Style{
			FillColor:   drawing.ColorWhite,
			FontColor:   DefaultTextColor,
			FontSize:    8.0,
			StrokeColor: DefaultAxisColor,
			StrokeWidth: DefaultAxisLineWidth,
		}

		var legendStyle Style
		if len(userDefaults) > 0 {
			legendStyle = userDefaults[0].InheritFrom(chartDefaults.InheritFrom(legendDefaults))
		} else {
			legendStyle = chartDefaults.InheritFrom(legendDefaults)
		}

		// DEFAULTS
		legendPadding := Box{
			Top:    5,
			Left:   5,
			Right:  5,
			Bottom: 5,
		}
		lineTextGap := 5
		lineLengthMinimum := 25

		entries := collectLegendSeries(c)

		legend := Box{
			Top:  5,
			Left: 5,
			// bottom and right will be sized by the legend content + relevant padding.
		}

		legendContent := Box{
			Top:    legend.Top + legendPadding.Top,
			Left:   legend.Left + legendPadding.Left,
			Right:  legend.Left + legendPadding.Left,
			Bottom: legend.Top + legendPadding.Top,
		}

		legendStyle.GetTextOptions().WriteToRenderer(r)

		// measure
		labelCount := 0
		for x := 0; x < len(entries); x++ {
			if len(entries[x].label) > 0 {
				tb := r.MeasureText(entries[x].label)
				if labelCount > 0 {
					legendContent.Bottom += DefaultMinimumTickVerticalSpacing
				}
				legendContent.Bottom += tb.Height()
				right := legendContent.Left + tb.Width() + lineTextGap + lineLengthMinimum
				legendContent.Right = MaxInt(legendContent.Right, right)
				labelCount++
			}
		}

		legend = legend.Grow(legendContent)
		legend.Right = legendContent.Right + legendPadding.Right
		legend.Bottom = legendContent.Bottom + legendPadding.Bottom

		Draw.Box(r, legend, legendStyle)

		legendStyle.GetTextOptions().WriteToRenderer(r)

		ycursor := legendContent.Top
		tx := legendContent.Left
		legendCount := 0
		var label string
		for x := 0; x < len(entries); x++ {
			label = entries[x].label
			if len(label) > 0 {
				if legendCount > 0 {
					ycursor += DefaultMinimumTickVerticalSpacing
				}

				tb := r.MeasureText(label)

				ty := ycursor + tb.Height()
				r.Text(label, tx, ty)

				th2 := tb.Height() >> 1

				lx := tx + tb.Width() + lineTextGap
				ly := ty - th2
				lx2 := legendContent.Right - legendPadding.Right

				drawLegendSwatch(r, entries[x], lx, ly, lx2-lx)

				ycursor += tb.Height()
				legendCount++
			}
		}
	}
}
