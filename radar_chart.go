package chart

import (
	"errors"
	"fmt"
	"io"
	"math"

	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

const (
	// DefaultRadarTickCount is the default number of concentric rings drawn on a radar chart.
	DefaultRadarTickCount = 5
	// DefaultRadarLabelGap is the distance in pixels between an axis vertex and its label.
	DefaultRadarLabelGap = 8
	// DefaultRadarLabelPadding is the space reserved around the radar for axis labels
	// when the radius is computed automatically.
	DefaultRadarLabelPadding = 12
	// DefaultRadarLineWidth is the default stroke width of a radar series outline.
	DefaultRadarLineWidth = 2.0
	// DefaultRadarDotWidth is the default radius of the points drawn on a radar series outline.
	DefaultRadarDotWidth = 3.0
	// DefaultRadarFillAlpha is the default alpha channel applied to radar series fills.
	DefaultRadarFillAlpha = 64
)

// RadarDimension is one ordered axis of a RadarChart.
//
// Max is the upper bound used to normalize that dimension's values onto the
// shared radial scale. It should be greater than zero. A zero (or invalid)
// Max means the dimension has no usable scale; values on that dimension are
// drawn at the center of the chart.
type RadarDimension struct {
	Name string
	Max  float64
}

// RadarSeries is one data outline on a RadarChart.
//
// Values are read in the same order as RadarChart.Dimensions. Missing values
// are treated as zero (drawn at the center) and extra values are ignored.
type RadarSeries struct {
	Name   string
	Style  Style
	Values []float64
}

// RadarChart is a chart that draws one or more closed data outlines against a
// shared set of ordered dimensions, with radial grid lines, axis labels and a
// legend.
//
// Dimensions start at the top of the chart and proceed clockwise.
type RadarChart struct {
	Title      string
	TitleStyle Style

	ColorPalette ColorPalette

	Width  int
	Height int
	DPI    float64

	Background Style
	Canvas     Style

	// GridStyle styles the concentric rings of the shared radial scale.
	GridStyle Style
	// AxisStyle styles the spokes that run from the center to each dimension.
	AxisStyle Style
	// LabelStyle styles the dimension labels that sit outside each vertex.
	LabelStyle Style
	// TickLabelStyle styles the labels attached to the concentric rings.
	TickLabelStyle Style
	// SeriesStyle is inherited by every series before the series' own style is applied.
	SeriesStyle Style
	// LegendStyle styles the legend. Use Hidden() to suppress the legend.
	LegendStyle Style

	Font        *truetype.Font
	defaultFont *truetype.Font

	// Radius is the radius in pixels of the outermost ring. When unset (or
	// non-positive) a radius is computed to fit the canvas and its labels.
	Radius int

	// TickCount is the number of concentric rings of the shared radial scale.
	TickCount int

	Dimensions []RadarDimension
	Series     []RadarSeries
	Elements   []Renderable
}

// NormalizeRadarValue maps a raw dimension value onto the shared [0,1] radial
// scale using that dimension's upper bound.
//
// The result is always finite and within [0,1]. Zero and negative values,
// non-finite values (NaN or infinity) and a non-positive or non-finite
// upper bound all normalize to 0, i.e. the center of the chart. Values above
// the upper bound are clamped to 1, i.e. the outermost ring.
func NormalizeRadarValue(value, axisMax float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	if math.IsNaN(axisMax) || math.IsInf(axisMax, 0) || axisMax <= 0 {
		return 0
	}
	if value <= 0 {
		return 0
	}
	fraction := value / axisMax
	if fraction > 1 {
		return 1
	}
	if fraction < 0 {
		return 0
	}
	return fraction
}

// RadarPoint returns the pixel coordinate for a value on a radar dimension.
//
// The dimension identified by dimensionIndex (of dimensionCount) is placed
// clockwise starting from the top of the chart: index 0 is straight up,
// index 1 is to the right, and so on. fraction is the normalized value in
// [0,1]; 0 is the center and 1 is the outermost ring. A negative or
// non-finite fraction is treated as 0 and fractions above 1 are clamped to 1.
func RadarPoint(cx, cy int, radius float64, dimensionIndex, dimensionCount int, fraction float64) (x, y int) {
	if dimensionCount <= 0 {
		return cx, cy
	}

	index := dimensionIndex % dimensionCount
	if index < 0 {
		index += dimensionCount
	}

	if math.IsNaN(fraction) || math.IsInf(fraction, 0) || fraction <= 0 {
		return cx, cy
	}
	if fraction > 1 {
		fraction = 1
	}

	angle := float64(index) * _2pi / float64(dimensionCount)
	return CirclePoint(cx, cy, radius*fraction, angle)
}

// GetDPI returns the dpi for the chart.
func (rc RadarChart) GetDPI(defaults ...float64) float64 {
	if rc.DPI == 0 {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return DefaultDPI
	}
	return rc.DPI
}

// GetFont returns the text font.
func (rc RadarChart) GetFont() *truetype.Font {
	if rc.Font == nil {
		return rc.defaultFont
	}
	return rc.Font
}

// GetWidth returns the chart width or the default value.
func (rc RadarChart) GetWidth() int {
	if rc.Width == 0 {
		return DefaultChartWidth
	}
	return rc.Width
}

// GetHeight returns the chart height or the default value.
func (rc RadarChart) GetHeight() int {
	if rc.Height == 0 {
		return DefaultChartHeight
	}
	return rc.Height
}

// GetTickCount returns the number of concentric rings of the radial scale.
func (rc RadarChart) GetTickCount() int {
	if rc.TickCount <= 0 {
		return DefaultRadarTickCount
	}
	return rc.TickCount
}

// GetColorPalette returns the color palette for the chart.
func (rc RadarChart) GetColorPalette() ColorPalette {
	if rc.ColorPalette != nil {
		return rc.ColorPalette
	}
	return AlternateColorPalette
}

// NormalizedValues returns the per-dimension normalized values for a series,
// in the same order as RadarChart.Dimensions.
//
// Missing values are returned as 0 and extra values are dropped. An invalid
// series index returns an empty slice.
func (rc RadarChart) NormalizedValues(seriesIndex int) []float64 {
	if seriesIndex < 0 || seriesIndex >= len(rc.Series) {
		return nil
	}
	return rc.normalizedValues(rc.Series[seriesIndex])
}

// SeriesPoints returns the closed outline of a series as pixel coordinates.
//
// The outline is placed around (cx, cy) with the given outer radius. A
// degenerate series (one with no usable dimensions) returns no points. An
// invalid series index returns an empty slice.
func (rc RadarChart) SeriesPoints(seriesIndex, cx, cy int, radius float64) []Point {
	if seriesIndex < 0 || seriesIndex >= len(rc.Series) {
		return nil
	}
	values := rc.normalizedValues(rc.Series[seriesIndex])
	points := make([]Point, len(values))
	for index, fraction := range values {
		x, y := RadarPoint(cx, cy, radius, index, len(values), fraction)
		points[index] = Point{x, y}
	}
	return points
}

// Render renders the chart with the given renderer to the given io.Writer.
func (rc RadarChart) Render(rp RendererProvider, w io.Writer) error {
	if len(rc.Dimensions) == 0 {
		return errors.New("please provide at least one dimension")
	}
	if len(rc.Series) == 0 {
		return errors.New("please provide at least one series")
	}

	r, err := rp(rc.GetWidth(), rc.GetHeight())
	if err != nil {
		return err
	}

	if rc.Font == nil {
		defaultFont, err := GetDefaultFont()
		if err != nil {
			return err
		}
		rc.defaultFont = defaultFont
	}
	r.SetDPI(rc.GetDPI(DefaultDPI))

	canvasBox := rc.getDefaultCanvasBox()

	rc.drawBackground(r)
	rc.drawCanvas(r, canvasBox)

	cx, cy := canvasBox.Center()
	radius := rc.getRadius(r, canvasBox)

	rc.drawGrid(r, cx, cy, radius)
	rc.drawAxes(r, cx, cy, radius)
	rc.drawDimensionLabels(r, cx, cy, radius)
	rc.drawTickLabels(r, cx, cy, radius)
	rc.drawSeries(r, cx, cy, radius)
	rc.drawLegend(r, canvasBox)
	rc.drawTitle(r)

	for _, a := range rc.Elements {
		a(r, canvasBox, rc.styleDefaultsElements())
	}

	return r.Save(w)
}

func (rc RadarChart) normalizedValues(series RadarSeries) []float64 {
	values := make([]float64, len(rc.Dimensions))
	for index, dimension := range rc.Dimensions {
		if index < len(series.Values) {
			values[index] = NormalizeRadarValue(series.Values[index], dimension.Max)
		}
	}
	return values
}

func (rc RadarChart) getRadius(r Renderer, canvasBox Box) float64 {
	if rc.Radius > 0 {
		return float64(rc.Radius)
	}

	// Reserve a deterministic band for the dimension labels so that the
	// computed radius (and therefore the whole radar geometry) is identical
	// for the SVG and PNG renderers, whose text measurement can differ.
	fontSize := rc.getLabelStyle().GetFontSize(DefaultAxisFontSize)
	labelBand := (fontSize * 3.0) + DefaultRadarLabelGap + DefaultRadarLabelPadding

	titleBand := 0.0
	if len(rc.Title) > 0 && !rc.TitleStyle.Hidden {
		titleBand = rc.styleDefaultsTitle().GetFontSize(DefaultTitleFontSize) + float64(DefaultTitleTop)
	}

	available := float64(MinInt(canvasBox.Width(), canvasBox.Height())) / 2.0
	radius := available - labelBand - titleBand
	if radius < 10 {
		radius = 10
	}
	return radius
}

func (rc RadarChart) drawBackground(r Renderer) {
	Draw.Box(r, Box{
		Right:  rc.GetWidth(),
		Bottom: rc.GetHeight(),
	}, rc.getBackgroundStyle())
}

func (rc RadarChart) drawCanvas(r Renderer, canvasBox Box) {
	Draw.Box(r, canvasBox, rc.getCanvasStyle())
}

func (rc RadarChart) drawGrid(r Renderer, cx, cy int, radius float64) {
	if rc.GridStyle.Hidden {
		return
	}

	style := rc.getGridStyle()
	style.GetFillAndStrokeOptions().WriteToRenderer(r)

	dimensionCount := len(rc.Dimensions)
	tickCount := rc.GetTickCount()
	for tick := 1; tick <= tickCount; tick++ {
		fraction := float64(tick) / float64(tickCount)
		rc.drawClosedPolygon(r, cx, cy, radius, dimensionCount, fraction)
		r.FillStroke()
	}
}

func (rc RadarChart) drawAxes(r Renderer, cx, cy int, radius float64) {
	if rc.AxisStyle.Hidden {
		return
	}

	style := rc.getAxisStyle()
	style.GetStrokeOptions().WriteToRenderer(r)

	dimensionCount := len(rc.Dimensions)
	for index := 0; index < dimensionCount; index++ {
		x, y := RadarPoint(cx, cy, radius, index, dimensionCount, 1)
		r.MoveTo(cx, cy)
		r.LineTo(x, y)
		r.Stroke()
	}
}

func (rc RadarChart) drawDimensionLabels(r Renderer, cx, cy int, radius float64) {
	if rc.LabelStyle.Hidden {
		return
	}

	style := rc.getLabelStyle()
	style.GetTextOptions().WriteToRenderer(r)

	dimensionCount := len(rc.Dimensions)
	for index, dimension := range rc.Dimensions {
		if dimension.Name == "" {
			continue
		}

		x, y := RadarPoint(cx, cy, radius, index, dimensionCount, 1)
		angle := float64(index) * _2pi / float64(dimensionCount)
		directionX := math.Sin(angle)
		directionY := -math.Cos(angle)

		textBox := r.MeasureText(dimension.Name)

		var tx, ty int
		switch {
		case directionX > 0.001:
			tx = x + DefaultRadarLabelGap
		case directionX < -0.001:
			tx = x - DefaultRadarLabelGap - textBox.Width()
		default:
			tx = x - (textBox.Width() >> 1)
		}

		switch {
		case directionY > 0.001:
			ty = y + DefaultRadarLabelGap + textBox.Height()
		case directionY < -0.001:
			ty = y - DefaultRadarLabelGap
		default:
			ty = y + (textBox.Height() >> 1)
		}

		Draw.Text(r, dimension.Name, tx, ty, style)
	}
}

func (rc RadarChart) drawTickLabels(r Renderer, cx, cy int, radius float64) {
	if rc.TickLabelStyle.Hidden {
		return
	}

	style := rc.getTickLabelStyle()
	style.GetTextOptions().WriteToRenderer(r)

	tickCount := rc.GetTickCount()
	if tickCount <= 1 {
		return
	}

	// Draw the radial tick labels along the top spoke, skipping the
	// outermost ring so they don't collide with the dimension label.
	for tick := 1; tick < tickCount; tick++ {
		fraction := float64(tick) / float64(tickCount)
		x, y := RadarPoint(cx, cy, radius, 0, len(rc.Dimensions), fraction)
		label := fmt.Sprintf("%d%%", int(math.Round(fraction*100.0)))
		textBox := r.MeasureText(label)
		Draw.Text(r, label, x+4, y+(textBox.Height()>>1), style)
	}
}

func (rc RadarChart) drawSeries(r Renderer, cx, cy int, radius float64) {
	dimensionCount := len(rc.Dimensions)
	for index, series := range rc.Series {
		if series.Style.Hidden {
			continue
		}

		style := series.Style.InheritFrom(rc.styleDefaultsSeries(index))
		values := rc.normalizedValues(series)

		style.GetFillAndStrokeOptions().WriteToRenderer(r)
		rc.drawClosedPolygon(r, cx, cy, radius, dimensionCount, 0, values...)
		r.FillStroke()

		if style.ShouldDrawDot() {
			dotWidth := style.GetDotWidth()
			style.GetDotOptions().WriteToRenderer(r)
			for dimensionIndex := 0; dimensionIndex < dimensionCount; dimensionIndex++ {
				x, y := RadarPoint(cx, cy, radius, dimensionIndex, dimensionCount, values[dimensionIndex])
				r.Circle(dotWidth, x, y)
				r.FillStroke()
			}
		}
	}
}

// drawClosedPolygon traces a closed polygon at a uniform fraction of the
// radius, or, when values are supplied, at each dimension's own fraction.
func (rc RadarChart) drawClosedPolygon(r Renderer, cx, cy int, radius float64, dimensionCount int, fraction float64, values ...float64) {
	for index := 0; index < dimensionCount; index++ {
		pointFraction := fraction
		if index < len(values) {
			pointFraction = values[index]
		}
		x, y := RadarPoint(cx, cy, radius, index, dimensionCount, pointFraction)
		if index == 0 {
			r.MoveTo(x, y)
		} else {
			r.LineTo(x, y)
		}
	}
	r.Close()
}

func (rc RadarChart) drawLegend(r Renderer, canvasBox Box) {
	if rc.LegendStyle.Hidden {
		return
	}

	type legendEntry struct {
		name  string
		style Style
	}

	var entries []legendEntry
	for index, series := range rc.Series {
		if series.Style.Hidden || series.Name == "" {
			continue
		}
		entries = append(entries, legendEntry{
			name:  series.Name,
			style: series.Style.InheritFrom(rc.styleDefaultsSeries(index)),
		})
	}
	if len(entries) == 0 {
		return
	}

	style := rc.getLegendStyle()
	style.WriteToRenderer(r)

	padding := 5
	lineTextGap := 5
	lineLength := 25
	rowSpacing := 4

	var maxTextWidth, rowHeight int
	for _, entry := range entries {
		style.WriteToRenderer(r)
		textBox := r.MeasureText(entry.name)
		maxTextWidth = MaxInt(maxTextWidth, textBox.Width())
		rowHeight = MaxInt(rowHeight, textBox.Height())
	}

	contentWidth := maxTextWidth + lineTextGap + lineLength
	contentHeight := (rowHeight * len(entries)) + (rowSpacing * (len(entries) - 1))

	legendBox := Box{
		Left:   canvasBox.Left + padding,
		Top:    canvasBox.Top + padding,
		Right:  canvasBox.Left + padding + (padding * 2) + contentWidth,
		Bottom: canvasBox.Top + padding + (padding * 2) + contentHeight,
	}
	Draw.Box(r, legendBox, style)

	style.WriteToRenderer(r)

	y := legendBox.Top + padding
	for _, entry := range entries {
		textBox := r.MeasureText(entry.name)
		tx := legendBox.Left + padding
		ty := y + textBox.Height()
		r.Text(entry.name, tx, ty)

		lx := tx + maxTextWidth + lineTextGap
		ly := ty - (textBox.Height() >> 1)

		r.SetStrokeColor(entry.style.GetStrokeColor())
		r.SetStrokeWidth(entry.style.GetStrokeWidth())
		r.SetStrokeDashArray(entry.style.GetStrokeDashArray())

		r.MoveTo(lx, ly)
		r.LineTo(lx+lineLength, ly)
		r.Stroke()

		y += textBox.Height() + rowSpacing
	}
}

func (rc RadarChart) drawTitle(r Renderer) {
	if len(rc.Title) > 0 && !rc.TitleStyle.Hidden {
		Draw.TextWithin(r, rc.Title, rc.Box(), rc.styleDefaultsTitle())
	}
}

func (rc RadarChart) getDefaultCanvasBox() Box {
	return rc.Box()
}

func (rc RadarChart) getBackgroundStyle() Style {
	return rc.Background.InheritFrom(rc.styleDefaultsBackground())
}

func (rc RadarChart) getCanvasStyle() Style {
	return rc.Canvas.InheritFrom(rc.styleDefaultsCanvas())
}

func (rc RadarChart) getGridStyle() Style {
	return rc.GridStyle.InheritFrom(rc.styleDefaultsGrid())
}

func (rc RadarChart) getAxisStyle() Style {
	return rc.AxisStyle.InheritFrom(rc.styleDefaultsAxis())
}

func (rc RadarChart) getLabelStyle() Style {
	return rc.LabelStyle.InheritFrom(rc.styleDefaultsLabel())
}

func (rc RadarChart) getTickLabelStyle() Style {
	return rc.TickLabelStyle.InheritFrom(rc.styleDefaultsTickLabel())
}

func (rc RadarChart) getLegendStyle() Style {
	return rc.LegendStyle.InheritFrom(rc.styleDefaultsLegend())
}

func (rc RadarChart) styleDefaultsCanvas() Style {
	return Style{
		FillColor:   rc.GetColorPalette().CanvasColor(),
		StrokeColor: rc.GetColorPalette().CanvasStrokeColor(),
		StrokeWidth: DefaultCanvasStrokeWidth,
	}
}

func (rc RadarChart) styleDefaultsGrid() Style {
	return Style{
		StrokeColor: ColorAlternateLightGray,
		StrokeWidth: DefaultAxisLineWidth,
	}
}

func (rc RadarChart) styleDefaultsAxis() Style {
	return Style{
		StrokeColor: rc.GetColorPalette().AxisStrokeColor(),
		StrokeWidth: DefaultAxisLineWidth,
	}
}

func (rc RadarChart) styleDefaultsLabel() Style {
	return Style{
		Font:      rc.GetFont(),
		FontColor: rc.GetColorPalette().TextColor(),
		FontSize:  DefaultAxisFontSize,
	}
}

func (rc RadarChart) styleDefaultsTickLabel() Style {
	return Style{
		Font:      rc.GetFont(),
		FontColor: rc.GetColorPalette().TextColor(),
		FontSize:  DefaultMinimumFontSize,
	}
}

func (rc RadarChart) styleDefaultsSeries(index int) Style {
	seriesColor := rc.GetColorPalette().GetSeriesColor(index)
	return Style{
		StrokeColor: seriesColor,
		StrokeWidth: DefaultRadarLineWidth,
		FillColor:   seriesColor.WithAlpha(DefaultRadarFillAlpha),
		DotColor:    seriesColor,
		DotWidth:    DefaultRadarDotWidth,
	}
}

func (rc RadarChart) styleDefaultsLegend() Style {
	return Style{
		FillColor:   drawing.ColorWhite,
		StrokeColor: rc.GetColorPalette().AxisStrokeColor(),
		StrokeWidth: DefaultAxisLineWidth,
		Font:        rc.GetFont(),
		FontColor:   rc.GetColorPalette().TextColor(),
		FontSize:    DefaultFontSize,
	}
}

func (rc RadarChart) styleDefaultsTitle() Style {
	return rc.TitleStyle.InheritFrom(Style{
		FontColor:           rc.GetColorPalette().TextColor(),
		Font:                rc.GetFont(),
		FontSize:            rc.getTitleFontSize(),
		TextHorizontalAlign: TextHorizontalAlignCenter,
		TextVerticalAlign:   TextVerticalAlignTop,
		TextWrap:            TextWrapWord,
	})
}

func (rc RadarChart) styleDefaultsBackground() Style {
	return Style{
		FillColor:   rc.GetColorPalette().BackgroundColor(),
		StrokeColor: rc.GetColorPalette().BackgroundStrokeColor(),
		StrokeWidth: DefaultBackgroundStrokeWidth,
	}
}

func (rc RadarChart) styleDefaultsElements() Style {
	return Style{
		Font: rc.GetFont(),
	}
}

func (rc RadarChart) getTitleFontSize() float64 {
	effectiveDimension := MinInt(rc.GetWidth(), rc.GetHeight())
	if effectiveDimension >= 2048 {
		return 48
	} else if effectiveDimension >= 1024 {
		return 24
	} else if effectiveDimension >= 512 {
		return 18
	} else if effectiveDimension >= 256 {
		return 12
	}
	return 10
}

// Box returns the chart bounds as a box.
func (rc RadarChart) Box() Box {
	dpr := rc.Background.Padding.GetRight(DefaultBackgroundPadding.Right)
	dpb := rc.Background.Padding.GetBottom(DefaultBackgroundPadding.Bottom)

	return Box{
		Top:    rc.Background.Padding.GetTop(DefaultBackgroundPadding.Top),
		Left:   rc.Background.Padding.GetLeft(DefaultBackgroundPadding.Left),
		Right:  rc.GetWidth() - dpr,
		Bottom: rc.GetHeight() - dpb,
	}
}
