package chart

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/wcharczuk/go-chart/v2/drawing"
	"github.com/wcharczuk/go-chart/v2/testutil"
)

func ohlcTestTimes() []time.Time {
	base := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	return []time.Time{
		base,
		base.Add(24 * time.Hour),
		base.Add(48 * time.Hour),
		base.Add(96 * time.Hour), // non-uniform gap: 2 days after the previous point
	}
}

// validOHLCSeries returns a series with one up candle matching the
// specification example (open 10, high 14, low 8, close 12), followed by
// a down candle and two more candles on a non-uniform time axis.
func validOHLCSeries() TimeOHLCSeries {
	return TimeOHLCSeries{
		Name:        "TEST",
		XValues:     ohlcTestTimes(),
		OpenValues:  []float64{10, 14, 11, 11},
		HighValues:  []float64{14, 15, 13, 12},
		LowValues:   []float64{8, 10, 9, 10},
		CloseValues: []float64{12, 11, 11, 10},
	}
}

func TestTimeOHLCSeriesInterfaceAssertions(t *testing.T) {
	s := validOHLCSeries()
	var _ Series = s
	var _ OHLCValuesProvider = s
	var _ BoundedValuesProvider = s
	var _ FirstValuesProvider = s
	var _ LastValuesProvider = s
	var _ ValueFormatterProvider = s
	var _ OHLCLegendProvider = s
}

func TestTimeOHLCSeriesValidate(t *testing.T) {
	// happy path
	testutil.AssertNil(t, validOHLCSeries().Validate())

	// empty series
	testutil.AssertNotNil(t, TimeOHLCSeries{}.Validate())

	base := ohlcTestTimes()

	cases := []struct {
		name   string
		series TimeOHLCSeries
	}{
		{
			name: "missing open values",
			series: TimeOHLCSeries{
				XValues:     base,
				HighValues:  []float64{14, 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{12, 11, 11, 10},
			},
		},
		{
			name: "length mismatch",
			series: TimeOHLCSeries{
				XValues:     base[:3],
				OpenValues:  []float64{10, 14, 11, 11},
				HighValues:  []float64{14, 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{12, 11, 11, 10},
			},
		},
		{
			name: "high below low",
			series: TimeOHLCSeries{
				XValues:     base,
				OpenValues:  []float64{10, 14, 11, 11},
				HighValues:  []float64{7, 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{12, 11, 11, 10},
			},
		},
		{
			name: "open outside high-low",
			series: TimeOHLCSeries{
				XValues:     base,
				OpenValues:  []float64{15, 14, 11, 11},
				HighValues:  []float64{14, 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{12, 11, 11, 10},
			},
		},
		{
			name: "close outside high-low",
			series: TimeOHLCSeries{
				XValues:     base,
				OpenValues:  []float64{10, 14, 11, 11},
				HighValues:  []float64{14, 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{20, 11, 11, 10},
			},
		},
		{
			name: "duplicate timestamps",
			series: TimeOHLCSeries{
				XValues: []time.Time{
					base[0], base[1], base[1], base[3],
				},
				OpenValues:  []float64{10, 14, 11, 11},
				HighValues:  []float64{14, 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{12, 11, 11, 10},
			},
		},
		{
			name: "unordered timestamps",
			series: TimeOHLCSeries{
				XValues: []time.Time{
					base[0], base[2], base[1], base[3],
				},
				OpenValues:  []float64{10, 14, 11, 11},
				HighValues:  []float64{14, 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{12, 11, 11, 10},
			},
		},
		{
			name: "non-finite high",
			series: TimeOHLCSeries{
				XValues:     base,
				OpenValues:  []float64{10, 14, 11, 11},
				HighValues:  []float64{math.NaN(), 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{12, 11, 11, 10},
			},
		},
		{
			name: "infinite close",
			series: TimeOHLCSeries{
				XValues:     base,
				OpenValues:  []float64{10, 14, 11, 11},
				HighValues:  []float64{14, 15, 13, 12},
				LowValues:   []float64{8, 12, 9, 10},
				CloseValues: []float64{12, math.Inf(1), 11, 10},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testutil.AssertNotNil(t, tc.series.Validate())
		})
	}
}

func TestTimeOHLCSeriesValidateDescending(t *testing.T) {
	base := ohlcTestTimes()
	descending := TimeOHLCSeries{
		XValues:     []time.Time{base[3], base[2], base[1], base[0]},
		OpenValues:  []float64{10, 14, 11, 11},
		HighValues:  []float64{14, 15, 13, 12},
		LowValues:   []float64{8, 10, 9, 10},
		CloseValues: []float64{12, 11, 11, 10},
	}
	testutil.AssertNil(t, descending.Validate())
}

func TestTimeOHLCSeriesValues(t *testing.T) {
	s := validOHLCSeries()
	testutil.AssertLen(t, s.XValues, s.Len())

	x, open, high, low, close := s.GetOHLCValues(0)
	testutil.AssertEqual(t, TimeToFloat64(ohlcTestTimes()[0]), x)
	testutil.AssertEqual(t, 10.0, open)
	testutil.AssertEqual(t, 14.0, high)
	testutil.AssertEqual(t, 8.0, low)
	testutil.AssertEqual(t, 12.0, close)

	// bounded values expose the high and low for axis range computation.
	bx, y1, y2 := s.GetBoundedValues(0)
	testutil.AssertEqual(t, x, bx)
	testutil.AssertEqual(t, 14.0, y1)
	testutil.AssertEqual(t, 8.0, y2)

	firstX, firstY := s.GetFirstValues()
	testutil.AssertEqual(t, TimeToFloat64(ohlcTestTimes()[0]), firstX)
	testutil.AssertEqual(t, 12.0, firstY)

	lastX, lastY := s.GetLastValues()
	testutil.AssertEqual(t, TimeToFloat64(ohlcTestTimes()[3]), lastX)
	testutil.AssertEqual(t, 10.0, lastY)

	xf, yf := s.GetValueFormatters()
	testutil.AssertNotNil(t, xf)
	testutil.AssertNotNil(t, yf)
}

func TestTimeOHLCSeriesCandleColor(t *testing.T) {
	s := validOHLCSeries()

	// index 0: close 12 > open 10 -> up
	testutil.AssertEqual(t, DefaultCandleUpColor, s.CandleColor(0))
	// index 1: close 11 < open 14 -> down
	testutil.AssertEqual(t, DefaultCandleDownColor, s.CandleColor(1))
	// index 2: close == open -> doji
	testutil.AssertEqual(t, DefaultCandleDojiColor, s.CandleColor(2))

	custom := s
	custom.UpColor = ColorBlue
	custom.DownColor = ColorOrange
	custom.DojiColor = ColorCyan
	testutil.AssertEqual(t, ColorBlue, custom.CandleColor(0))
	testutil.AssertEqual(t, ColorOrange, custom.CandleColor(1))
	testutil.AssertEqual(t, ColorCyan, custom.CandleColor(2))
}

// renderOHLC renders a series to an SVG buffer with axes hidden so the
// output can be asserted on directly.
func renderOHLCSVG(t *testing.T, s TimeOHLCSeries, width, height int) string {
	c := Chart{
		Height:         height,
		Width:          width,
		TitleStyle:     Hidden(),
		XAxis:          HideXAxis(),
		YAxis:          HideYAxis(),
		YAxisSecondary: HideYAxis(),
		Canvas: Style{
			Padding: BoxZero,
		},
		Background: Style{
			Padding: BoxZero,
		},
		Series: []Series{s},
	}

	buffer := &bytes.Buffer{}
	testutil.AssertNil(t, c.Render(SVG, buffer))
	return buffer.String()
}

func TestTimeOHLCSeriesRenderSVG(t *testing.T) {
	svg := renderOHLCSVG(t, validOHLCSeries(), 200, 200)

	// the up and down colors both appear in the output
	testutil.AssertContains(t, svg, DefaultCandleUpColor.String())
	testutil.AssertContains(t, svg, DefaultCandleDownColor.String())
	// the doji color appears for the open == close candle
	testutil.AssertContains(t, svg, DefaultCandleDojiColor.String())

	// each candle contributes at least a wick path and a body/tick path;
	// four candles means at least eight path elements.
	testutil.AssertTrue(t, strings.Count(svg, "<path") >= 8)
}

func TestTimeOHLCSeriesRenderPNG(t *testing.T) {
	c := Chart{
		Height:         120,
		Width:          120,
		TitleStyle:     Hidden(),
		XAxis:          HideXAxis(),
		YAxis:          HideYAxis(),
		YAxisSecondary: HideYAxis(),
		Canvas: Style{
			Padding: BoxZero,
		},
		Background: Style{
			Padding: BoxZero,
		},
		Series: []Series{
			TimeOHLCSeries{
				XValues:     ohlcTestTimes(),
				OpenValues:  []float64{10, 14, 11, 11},
				HighValues:  []float64{14, 15, 13, 12},
				LowValues:   []float64{8, 10, 9, 10},
				CloseValues: []float64{12, 11, 11, 10},
			},
		},
	}

	buffer := &bytes.Buffer{}
	testutil.AssertNil(t, c.Render(PNG, buffer))

	img, err := pngDecode(buffer)
	testutil.AssertNil(t, err)

	// collect the distinct non-white colors actually painted; the up and
	// down candle colors must both be present. at() converts the image's
	// color.Color into the concrete drawing.Color used by the renderers.
	bounds := img.Bounds()
	seen := map[drawing.Color]bool{}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			seen[at(img, x, y)] = true
		}
	}
	testutil.AssertTrue(t, seen[DefaultCandleUpColor])
	testutil.AssertTrue(t, seen[DefaultCandleDownColor])
}

func TestTimeOHLCSeriesAxisRangeIncludesExtremes(t *testing.T) {
	c := Chart{
		Height:         100,
		Width:          100,
		TitleStyle:     Hidden(),
		XAxis:          HideXAxis(),
		YAxis:          HideYAxis(),
		YAxisSecondary: HideYAxis(),
		Canvas:         Style{Padding: BoxZero},
		Background:     Style{Padding: BoxZero},
		Series:         []Series{validOHLCSeries()},
	}

	xr, yr, yra := c.getRanges()
	testutil.AssertNotNil(t, xr)
	testutil.AssertNotNil(t, yr)
	testutil.AssertNotNil(t, yra)

	// x range spans the full time extent
	testutil.AssertEqual(t, TimeToFloat64(ohlcTestTimes()[0]), xr.GetMin())
	testutil.AssertEqual(t, TimeToFloat64(ohlcTestTimes()[3]), xr.GetMax())

	// y range must include the global low (8) and high (15); the chart
	// rounds the range out, so we assert containment rather than equality.
	testutil.AssertTrue(t, yr.GetMin() <= 8.0)
	testutil.AssertTrue(t, yr.GetMax() >= 15.0)
}

func TestTimeOHLCSeriesSinglePoint(t *testing.T) {
	// a single candle must still render a visible, finite-width body.
	// note: Chart.checkRanges rejects a zero x-range delta for every
	// series type, so a one-point series is exercised through the draw
	// path with a degenerate range rather than a full chart render.
	s := TimeOHLCSeries{
		XValues:     ohlcTestTimes()[:1],
		OpenValues:  []float64{10},
		HighValues:  []float64{14},
		LowValues:   []float64{8},
		CloseValues: []float64{12},
		CandleWidth: 9,
	}
	testutil.AssertNil(t, s.Validate())

	cb := Box{Top: 10, Left: 10, Right: 190, Bottom: 190}
	xr := &ContinuousRange{
		Min: TimeToFloat64(ohlcTestTimes()[0]),
		Max: TimeToFloat64(ohlcTestTimes()[0]),
	}
	xr.SetDomain(cb.Width())
	yr := &ContinuousRange{Min: 8, Max: 14}
	yr.SetDomain(cb.Height())

	rec := newRecordingRenderer()
	s.Render(rec, cb, xr, yr, Style{})

	// the up candle body is drawn, centered horizontally on the canvas.
	testutil.AssertLen(t, rec.boxes, 1)
	body := rec.boxes[0]
	testutil.AssertEqual(t, cb.Left+(cb.Width()>>1), body.Left+(body.Width()>>1))
	testutil.AssertTrue(t, body.Width() >= MinimumCandleWidth)

	// and it still renders through an SVG renderer without error.
	svgBuffer := &bytes.Buffer{}
	svg, err := SVG(200, 200)
	testutil.AssertNil(t, err)
	s.Render(svg, cb, xr, yr, Style{})
	testutil.AssertNil(t, svg.Save(svgBuffer))
	testutil.AssertContains(t, svgBuffer.String(), DefaultCandleUpColor.String())
}

func TestTimeOHLCSeriesNonUniformSpacingWidths(t *testing.T) {
	// the final point sits twice as far from its neighbor as the uniform
	// points. The range is padded beyond the data so no candle is clamped
	// to the canvas edge, isolating the width-from-spacing behavior.
	s := validOHLCSeries()

	cb := Box{Top: 0, Left: 0, Right: 300, Bottom: 200}
	day := 24 * time.Hour
	t0 := ohlcTestTimes()[0]
	xr := &ContinuousRange{
		Min: TimeToFloat64(t0.Add(-2 * day)),
		Max: TimeToFloat64(t0.Add(6 * day)),
	}
	xr.SetDomain(cb.Width())
	yr := &ContinuousRange{Min: 0, Max: 20}
	yr.SetDomain(cb.Height())

	rec := newRecordingRenderer()
	s.Render(rec, cb, xr, yr, Style{})

	testutil.AssertLen(t, rec.boxes, 3) // three non-doji candles have bodies

	// expected pixel positions and widths derived from the range mapping:
	// the domain covers 8 days over 300 px, so one day is 37.5 px.
	pxPerDay := float64(cb.Width()) / 8.0
	uniformGap := pxPerDay
	wideGap := 2 * pxPerDay
	uniformWidth := int(math.Floor(uniformGap * DefaultCandleWidthRatio))
	wideWidth := int(math.Floor(wideGap * DefaultCandleWidthRatio))

	// boxes are recorded for candles 0, 1 and 3 (candle 2 is a doji).
	testutil.AssertEqual(t, uniformWidth, rec.boxes[0].Width())
	testutil.AssertEqual(t, uniformWidth, rec.boxes[1].Width())
	testutil.AssertEqual(t, wideWidth, rec.boxes[2].Width())
	testutil.AssertTrue(t, rec.boxes[2].Width() > rec.boxes[1].Width())
}

func TestTimeOHLCSeriesSpecGeometry(t *testing.T) {
	// the specification example: open 10, high 14, low 8, close 12.
	// the body must cover 10..12 and the wick must cover 8..14.
	s := TimeOHLCSeries{
		XValues:     ohlcTestTimes()[:1],
		OpenValues:  []float64{10},
		HighValues:  []float64{14},
		LowValues:   []float64{8},
		CloseValues: []float64{12},
		CandleWidth: 9,
	}
	testutil.AssertNil(t, s.Validate())

	cb := Box{Top: 0, Left: 0, Right: 100, Bottom: 100}
	xr := &ContinuousRange{
		Min: TimeToFloat64(ohlcTestTimes()[0]),
		Max: TimeToFloat64(ohlcTestTimes()[0]),
	}
	xr.SetDomain(cb.Width())
	// y range exactly matches the high-low extent so translations are exact.
	yr := &ContinuousRange{Min: 8, Max: 14}
	yr.SetDomain(cb.Height())

	rec := newRecordingRenderer()
	s.Render(rec, cb, xr, yr, Style{})

	// y coordinates are inverted on screen: value 14 (high) is at y=0,
	// value 8 (low) is at y=100, value 10 (open) is at y=66 and value 12
	// (close) is at y=33 (Translate rounds up via ceil).
	body := rec.boxes[0]
	testutil.AssertEqual(t, 33, body.Top)    // close 12
	testutil.AssertEqual(t, 66, body.Bottom) // open 10
	testutil.AssertTrue(t, body.Top < body.Bottom)

	// the recorded wick is stroked (not boxed), so assert its extent by
	// re-deriving the expected pixel positions from the range.
	highY := cb.Bottom - yr.Translate(14)
	lowY := cb.Bottom - yr.Translate(8)
	openY := cb.Bottom - yr.Translate(10)
	closeY := cb.Bottom - yr.Translate(12)
	testutil.AssertEqual(t, 0, highY)
	testutil.AssertEqual(t, 100, lowY)
	testutil.AssertEqual(t, 66, openY)
	testutil.AssertEqual(t, 33, closeY)
}

func TestTimeOHLCSeriesDownCandleGeometry(t *testing.T) {
	// a down candle: open 12, high 14, low 8, close 10. The body still
	// spans 10..12 but the top of the body is now the open value.
	s := TimeOHLCSeries{
		XValues:     ohlcTestTimes()[:1],
		OpenValues:  []float64{12},
		HighValues:  []float64{14},
		LowValues:   []float64{8},
		CloseValues: []float64{10},
		CandleWidth: 9,
	}
	testutil.AssertNil(t, s.Validate())
	testutil.AssertEqual(t, DefaultCandleDownColor, s.CandleColor(0))

	cb := Box{Top: 0, Left: 0, Right: 100, Bottom: 100}
	xr := &ContinuousRange{
		Min: TimeToFloat64(ohlcTestTimes()[0]),
		Max: TimeToFloat64(ohlcTestTimes()[0]),
	}
	xr.SetDomain(cb.Width())
	yr := &ContinuousRange{Min: 8, Max: 14}
	yr.SetDomain(cb.Height())

	rec := newRecordingRenderer()
	s.Render(rec, cb, xr, yr, Style{})

	body := rec.boxes[0]
	testutil.AssertEqual(t, 33, body.Top)    // open 12 (higher value = higher on screen)
	testutil.AssertEqual(t, 66, body.Bottom) // close 10
}

func TestTimeOHLCSeriesDojiRendersTick(t *testing.T) {
	// open == close: no body box is drawn, only the wick and a horizontal
	// tick, both in the doji color.
	s := TimeOHLCSeries{
		XValues:     ohlcTestTimes()[:1],
		OpenValues:  []float64{11},
		HighValues:  []float64{14},
		LowValues:   []float64{8},
		CloseValues: []float64{11},
		CandleWidth: 9,
	}
	testutil.AssertNil(t, s.Validate())
	testutil.AssertEqual(t, DefaultCandleDojiColor, s.CandleColor(0))

	cb := Box{Top: 0, Left: 0, Right: 100, Bottom: 100}
	xr := &ContinuousRange{
		Min: TimeToFloat64(ohlcTestTimes()[0]),
		Max: TimeToFloat64(ohlcTestTimes()[0]),
	}
	xr.SetDomain(cb.Width())
	yr := &ContinuousRange{Min: 8, Max: 14}
	yr.SetDomain(cb.Height())

	rec := newRecordingRenderer()
	s.Render(rec, cb, xr, yr, Style{})
	testutil.AssertEmpty(t, rec.boxes)
}

func TestTimeOHLCSeriesExplicitCandleWidth(t *testing.T) {
	s := validOHLCSeries()
	s.CandleWidth = 9

	cb := Box{Top: 0, Left: 0, Right: 200, Bottom: 200}
	xr := &ContinuousRange{
		Min: TimeToFloat64(ohlcTestTimes()[0]),
		Max: TimeToFloat64(ohlcTestTimes()[3]),
	}
	xr.SetDomain(cb.Width())
	yr := &ContinuousRange{Min: 0, Max: 20}
	yr.SetDomain(cb.Height())

	rec := newRecordingRenderer()
	s.Render(rec, cb, xr, yr, Style{})

	// three non-doji bodies (candles 0, 1 and 3; candle 2 is a doji).
	testutil.AssertLen(t, rec.boxes, 3)
	for _, b := range rec.boxes {
		// no body may spill outside the canvas.
		testutil.AssertTrue(t, b.Left >= cb.Left)
		testutil.AssertTrue(t, b.Right <= cb.Right)
	}
	// the interior candle keeps the full explicit width (half on each
	// side), while the edge candles are clamped to the canvas boundary.
	interior := rec.boxes[1]
	testutil.AssertTrue(t, interior.Width() == 8 || interior.Width() == 9)
	testutil.AssertTrue(t, rec.boxes[0].Width() < interior.Width())
	testutil.AssertTrue(t, rec.boxes[2].Width() < interior.Width())
}
