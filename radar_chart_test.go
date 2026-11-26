package chart

import (
	"bytes"
	"image/png"
	"math"
	"testing"

	"github.com/wcharczuk/go-chart/v2/drawing"
	"github.com/wcharczuk/go-chart/v2/testutil"
)

func TestNormalizeRadarValue(t *testing.T) {
	testutil.AssertEqual(t, 0.0, NormalizeRadarValue(0, 100))
	testutil.AssertEqual(t, 0.0, NormalizeRadarValue(-5, 100))
	testutil.AssertEqual(t, 0.25, NormalizeRadarValue(25, 100))
	testutil.AssertEqual(t, 1.0, NormalizeRadarValue(100, 100))
	testutil.AssertEqual(t, 1.0, NormalizeRadarValue(150, 100))
	testutil.AssertEqual(t, 0.0, NormalizeRadarValue(50, 0))
	testutil.AssertEqual(t, 0.0, NormalizeRadarValue(50, -10))
	testutil.AssertEqual(t, 0.0, NormalizeRadarValue(math.NaN(), 100))
	testutil.AssertEqual(t, 0.0, NormalizeRadarValue(math.Inf(1), 100))
	testutil.AssertEqual(t, 0.0, NormalizeRadarValue(50, math.NaN()))
	testutil.AssertEqual(t, 0.0, NormalizeRadarValue(50, math.Inf(1)))
}

func TestRadarPointFourDimensions(t *testing.T) {
	// Four dimensions, radius 100, starting at the top and moving clockwise.
	// Up should be (100,0) relative to the center, right (100,100), etc.
	cx, cy := 100, 100
	radius := 100.0

	upX, upY := RadarPoint(cx, cy, radius, 0, 4, 1)
	rightX, rightY := RadarPoint(cx, cy, radius, 1, 4, 1)
	downX, downY := RadarPoint(cx, cy, radius, 2, 4, 1)
	leftX, leftY := RadarPoint(cx, cy, radius, 3, 4, 1)

	testutil.AssertEqual(t, 100, upX)
	testutil.AssertEqual(t, 0, upY)

	testutil.AssertEqual(t, 200, rightX)
	testutil.AssertEqual(t, 100, rightY)

	testutil.AssertEqual(t, 100, downX)
	testutil.AssertEqual(t, 200, downY)

	testutil.AssertEqual(t, 0, leftX)
	testutil.AssertEqual(t, 100, leftY)
}

func TestRadarPointHalfValue(t *testing.T) {
	cx, cy := 100, 100
	radius := 100.0

	// A value of 0.5 on the right axis should land 50 pixels to the right.
	x, y := RadarPoint(cx, cy, radius, 1, 4, 0.5)
	testutil.AssertEqual(t, 150, x)
	testutil.AssertEqual(t, 100, y)

	// A zero value should land on the center.
	zx, zy := RadarPoint(cx, cy, radius, 2, 4, 0)
	testutil.AssertEqual(t, cx, zx)
	testutil.AssertEqual(t, cy, zy)
}

func TestRadarChartSeriesPoints(t *testing.T) {
	// The exact scenario from the task: four dimensions, radius 100, values
	// {1, 0.5, 1, 0.5} with a per-dimension max of 1. The four vertices should
	// be up 100, right 50, down 100, left 50 pixels from the center.
	rc := RadarChart{
		Dimensions: []RadarDimension{
			{Name: "Attack", Max: 1},
			{Name: "Defense", Max: 1},
			{Name: "Speed", Max: 1},
			{Name: "Magic", Max: 1},
		},
		Series: []RadarSeries{
			{Name: "Player", Values: []float64{1, 0.5, 1, 0.5}},
		},
	}

	points := rc.SeriesPoints(0, 100, 100, 100)
	testutil.AssertLen(t, points, 4)

	testutil.AssertEqual(t, Point{100, 0}, points[0])
	testutil.AssertEqual(t, Point{150, 100}, points[1])
	testutil.AssertEqual(t, Point{100, 200}, points[2])
	testutil.AssertEqual(t, Point{50, 100}, points[3])
}

func TestRadarChartTwoOutlines(t *testing.T) {
	// Two series share the dimensions and the radial scale but produce their
	// own independent closed outlines.
	rc := RadarChart{
		Dimensions: []RadarDimension{
			{Name: "Attack", Max: 2},
			{Name: "Defense", Max: 2},
			{Name: "Speed", Max: 2},
			{Name: "Magic", Max: 2},
		},
		Series: []RadarSeries{
			{Name: "Full", Values: []float64{2, 2, 2, 2}},
			{Name: "Half", Values: []float64{1, 1, 1, 1}},
		},
	}

	cx, cy := 100, 100
	radius := 100.0

	full := rc.SeriesPoints(0, cx, cy, radius)
	testutil.AssertLen(t, full, 4)
	testutil.AssertEqual(t, Point{100, 0}, full[0])
	testutil.AssertEqual(t, Point{200, 100}, full[1])
	testutil.AssertEqual(t, Point{100, 200}, full[2])
	testutil.AssertEqual(t, Point{0, 100}, full[3])

	half := rc.SeriesPoints(1, cx, cy, radius)
	testutil.AssertLen(t, half, 4)
	testutil.AssertEqual(t, Point{100, 50}, half[0])
	testutil.AssertEqual(t, Point{150, 100}, half[1])
	testutil.AssertEqual(t, Point{100, 150}, half[2])
	testutil.AssertEqual(t, Point{50, 100}, half[3])
}

func TestRadarChartDifferingValueCounts(t *testing.T) {
	// A series with fewer values than dimensions is drawn at the center for
	// the missing dimensions; a series with extra values drops the extras.
	rc := RadarChart{
		Dimensions: []RadarDimension{
			{Name: "A", Max: 10},
			{Name: "B", Max: 10},
			{Name: "C", Max: 10},
		},
		Series: []RadarSeries{
			{Name: "Short", Values: []float64{5}},
			{Name: "Long", Values: []float64{10, 10, 10, 10, 10}},
		},
	}

	shortValues := rc.NormalizedValues(0)
	testutil.AssertLen(t, shortValues, 3)
	testutil.AssertEqual(t, 0.5, shortValues[0])
	testutil.AssertEqual(t, 0.0, shortValues[1])
	testutil.AssertEqual(t, 0.0, shortValues[2])

	longValues := rc.NormalizedValues(1)
	testutil.AssertLen(t, longValues, 3)
	testutil.AssertEqual(t, 1.0, longValues[0])
	testutil.AssertEqual(t, 1.0, longValues[1])
	testutil.AssertEqual(t, 1.0, longValues[2])
}

func TestRadarChartZeroAndNegativeAxisMax(t *testing.T) {
	rc := RadarChart{
		Dimensions: []RadarDimension{
			{Name: "Zero", Max: 0},
			{Name: "Negative", Max: -10},
			{Name: "Valid", Max: 10},
		},
		Series: []RadarSeries{
			{Values: []float64{50, 50, -5}},
		},
	}

	values := rc.NormalizedValues(0)
	testutil.AssertLen(t, values, 3)
	testutil.AssertEqual(t, 0.0, values[0])
	testutil.AssertEqual(t, 0.0, values[1])
	testutil.AssertEqual(t, 0.0, values[2])
}

func TestRadarChartNormalizedValues(t *testing.T) {
	rc := RadarChart{
		Dimensions: []RadarDimension{
			{Name: "Attack", Max: 200},
			{Name: "Defense", Max: 50},
			{Name: "Speed", Max: 100},
		},
		Series: []RadarSeries{
			{Values: []float64{100, 25}},
		},
	}

	values := rc.NormalizedValues(0)
	testutil.AssertLen(t, values, 3)
	testutil.AssertEqual(t, 0.5, values[0])
	testutil.AssertEqual(t, 0.5, values[1])
	testutil.AssertEqual(t, 0.0, values[2]) // missing value

	testutil.AssertNil(t, rc.NormalizedValues(-1))
	testutil.AssertNil(t, rc.NormalizedValues(5))
}

func TestRadarChartSeriesPointsDegenerate(t *testing.T) {
	rc := RadarChart{Dimensions: []RadarDimension{{Max: 1}}}
	testutil.AssertEmpty(t, rc.SeriesPoints(0, 0, 0, 100))
}

func TestRadarChartSingleDimension(t *testing.T) {
	// A single dimension collapses onto the vertical spoke; a zero value
	// draws a degenerate polygon at the center without panicking.
	rc := RadarChart{
		Dimensions: []RadarDimension{{Name: "Attack", Max: 100}},
		Series:     []RadarSeries{{Name: "Player", Values: []float64{50}}},
	}

	points := rc.SeriesPoints(0, 100, 100, 100)
	testutil.AssertLen(t, points, 1)
	testutil.AssertEqual(t, Point{100, 50}, points[0])

	zero := RadarChart{
		Dimensions: []RadarDimension{{Name: "Attack", Max: 100}},
		Series:     []RadarSeries{{Name: "Player", Values: []float64{0}}},
	}
	testutil.AssertEqual(t, Point{100, 100}, zero.SeriesPoints(0, 100, 100, 100)[0])

	buffer := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, rc.Render(PNG, buffer))
}

func TestRadarChartGetDPI(t *testing.T) {
	testutil.AssertEqual(t, DefaultDPI, RadarChart{}.GetDPI())
	testutil.AssertEqual(t, 192.0, RadarChart{}.GetDPI(192))
	testutil.AssertEqual(t, 128.0, RadarChart{DPI: 128}.GetDPI())
}

func TestRadarChartGetWidthAndHeight(t *testing.T) {
	unset := RadarChart{}
	testutil.AssertEqual(t, DefaultChartWidth, unset.GetWidth())
	testutil.AssertEqual(t, DefaultChartHeight, unset.GetHeight())

	set := RadarChart{Width: 640, Height: 480}
	testutil.AssertEqual(t, 640, set.GetWidth())
	testutil.AssertEqual(t, 480, set.GetHeight())
}

func TestRadarChartGetTickCount(t *testing.T) {
	testutil.AssertEqual(t, DefaultRadarTickCount, RadarChart{}.GetTickCount())
	testutil.AssertEqual(t, 3, RadarChart{TickCount: 3}.GetTickCount())
}

func TestRadarChartRenderMissingData(t *testing.T) {
	// Missing values are drawn at the center, not treated as an error.
	rc := RadarChart{
		Dimensions: []RadarDimension{
			{Name: "Attack", Max: 10},
			{Name: "Defense", Max: 10},
		},
		Series: []RadarSeries{{Name: "Player", Values: []float64{5}}},
	}

	b := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, rc.Render(PNG, b))

	// No dimensions and no series are both explicit errors.
	testutil.AssertNotNil(t, RadarChart{
		Series: []RadarSeries{{Name: "Player"}},
	}.Render(PNG, b))

	testutil.AssertNotNil(t, RadarChart{
		Dimensions: []RadarDimension{{Name: "Attack", Max: 10}},
	}.Render(PNG, b))
}

func TestRadarChartRenderPNG(t *testing.T) {
	rc := makeTestRadarChart()

	b := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, rc.Render(PNG, b))
	testutil.AssertNotZero(t, b.Len())

	// Verify the chart actually produced a decodable image.
	img, err := png.Decode(b)
	testutil.AssertNil(t, err)
	testutil.AssertNotNil(t, img)
}

func TestRadarChartRenderSVG(t *testing.T) {
	rc := makeTestRadarChart()

	b := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, rc.Render(SVG, b))
	testutil.AssertNotEmpty(t, b.String())
	testutil.AssertContains(t, b.String(), "<svg")
}

func TestRadarChartGeometryIsRendererIndependent(t *testing.T) {
	// The chart geometry must not depend on the renderer. Build the chart
	// twice, then compare the outlines computed for both renderers.
	rc := makeTestRadarChart()

	pngBuffer := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, rc.Render(PNG, pngBuffer))

	svgBuffer := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, rc.Render(SVG, svgBuffer))

	canvasBox := rc.getDefaultCanvasBox()
	cx, cy := canvasBox.Center()

	rPng, err := PNG(rc.GetWidth(), rc.GetHeight())
	testutil.AssertNil(t, err)

	rSvg, err := SVG(rc.GetWidth(), rc.GetHeight())
	testutil.AssertNil(t, err)

	testutil.AssertEqual(t, rc.getRadius(rPng, canvasBox), rc.getRadius(rSvg, canvasBox))

	radius := rc.getRadius(rPng, canvasBox)
	for seriesIndex := range rc.Series {
		pngPoints := rc.SeriesPoints(seriesIndex, cx, cy, radius)
		svgPoints := rc.SeriesPoints(seriesIndex, cx, cy, radius)
		testutil.AssertLen(t, pngPoints, len(svgPoints))
		for i := range pngPoints {
			testutil.AssertEqual(t, pngPoints[i], svgPoints[i])
		}
	}
}

func TestRadarChartSVGVertices(t *testing.T) {
	// A fully-filled four-dimension series with a fixed radius should be
	// emitted as an SVG polygon whose vertices are exactly the cardinal
	// points of the radar.
	rc := RadarChart{
		Width:          200,
		Height:         200,
		Radius:         80,
		GridStyle:      Hidden(),
		AxisStyle:      Hidden(),
		LabelStyle:     Hidden(),
		TickLabelStyle: Hidden(),
		LegendStyle:    Hidden(),
		TitleStyle:     Hidden(),
		Dimensions: []RadarDimension{
			{Name: "Attack", Max: 1},
			{Name: "Defense", Max: 1},
			{Name: "Speed", Max: 1},
			{Name: "Magic", Max: 1},
		},
		Series: []RadarSeries{
			{Name: "Player", Values: []float64{1, 1, 1, 1}},
		},
	}

	buffer := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, rc.Render(SVG, buffer))
	svg := buffer.String()

	// canvasBox center of a 200x200 chart with default background padding.
	testutil.AssertContains(t, svg, "M 100 20\nL 180 100\nL 100 180\nL 20 100\nZ")

	// And the same coordinates come out of the geometry helper.
	points := rc.SeriesPoints(0, 100, 100, 80)
	testutil.AssertEqual(t, Point{100, 20}, points[0])
	testutil.AssertEqual(t, Point{180, 100}, points[1])
	testutil.AssertEqual(t, Point{100, 180}, points[2])
	testutil.AssertEqual(t, Point{20, 100}, points[3])
}

func TestRadarChartFillColor(t *testing.T) {
	rc := makeTestRadarChart()
	style := rc.styleDefaultsSeries(0)
	testutil.AssertEqual(t, DefaultRadarFillAlpha, style.FillColor.A)
	testutil.AssertTrue(t, style.FillColor.R == rc.GetColorPalette().GetSeriesColor(0).R)
}

func TestRadarChartFillColorInherited(t *testing.T) {
	// A caller can override the whole fill, including opacity.
	rc := makeTestRadarChart()
	rc.Series[0].Style = Style{
		StrokeColor: drawing.ColorRed,
		FillColor:   drawing.ColorFromHex("ff0000").WithAlpha(128),
	}

	style := rc.Series[0].Style.InheritFrom(rc.styleDefaultsSeries(0))
	testutil.AssertEqual(t, uint8(128), style.FillColor.A)
	testutil.AssertEqual(t, drawing.ColorRed, style.StrokeColor)
}

func makeTestRadarChart() RadarChart {
	return RadarChart{
		Title:  "Character Stats",
		Width:  600,
		Height: 600,
		Radius: 200,
		Dimensions: []RadarDimension{
			{Name: "Attack", Max: 100},
			{Name: "Defense", Max: 100},
			{Name: "Speed", Max: 100},
			{Name: "Magic", Max: 100},
			{Name: "Luck", Max: 100},
		},
		Series: []RadarSeries{
			{Name: "Warrior", Values: []float64{80, 90, 40, 20, 50}},
			{Name: "Mage", Values: []float64{30, 40, 60, 95, 70}},
		},
	}
}
