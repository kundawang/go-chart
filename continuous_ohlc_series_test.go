package chart

import (
	"bytes"
	"math"
	"testing"

	"github.com/wcharczuk/go-chart/v2/testutil"
)

func validContinuousOHLCSeries() ContinuousOHLCSeries {
	return ContinuousOHLCSeries{
		Name:        "NUM",
		XValues:     []float64{1, 2, 3, 5}, // non-uniform final gap
		OpenValues:  []float64{10, 14, 11, 11},
		HighValues:  []float64{14, 15, 13, 12},
		LowValues:   []float64{8, 10, 9, 10},
		CloseValues: []float64{12, 11, 11, 10},
	}
}

func TestContinuousOHLCSeriesInterfaces(t *testing.T) {
	s := validContinuousOHLCSeries()
	var _ Series = s
	var _ OHLCValuesProvider = s
	var _ BoundedValuesProvider = s
	var _ OHLCLegendProvider = s

	x, open, high, low, close := s.GetOHLCValues(0)
	testutil.AssertEqual(t, 1.0, x)
	testutil.AssertEqual(t, 10.0, open)
	testutil.AssertEqual(t, 14.0, high)
	testutil.AssertEqual(t, 8.0, low)
	testutil.AssertEqual(t, 12.0, close)

	xf, yf := s.GetValueFormatters()
	testutil.AssertNotNil(t, xf)
	testutil.AssertNotNil(t, yf)
}

func TestContinuousOHLCSeriesValidate(t *testing.T) {
	testutil.AssertNil(t, validContinuousOHLCSeries().Validate())
	testutil.AssertNotNil(t, ContinuousOHLCSeries{}.Validate())

	bad := validContinuousOHLCSeries()
	bad.HighValues[0] = 5 // below the low of 8
	testutil.AssertNotNil(t, bad.Validate())

	bad = validContinuousOHLCSeries()
	bad.XValues[2] = 2 // duplicate x
	testutil.AssertNotNil(t, bad.Validate())

	bad = validContinuousOHLCSeries()
	bad.CloseValues[1] = math.Inf(-1)
	testutil.AssertNotNil(t, bad.Validate())
}

func TestContinuousOHLCSeriesRenderThroughChart(t *testing.T) {
	// unlike a single time point, a multi-point numeric series renders
	// end to end through Chart.Render in both renderers.
	series := validContinuousOHLCSeries()

	for _, provider := range []RendererProvider{PNG, SVG} {
		c := Chart{
			Height:         120,
			Width:          240,
			TitleStyle:     Hidden(),
			XAxis:          HideXAxis(),
			YAxis:          HideYAxis(),
			YAxisSecondary: HideYAxis(),
			Canvas:         Style{Padding: BoxZero},
			Background:     Style{Padding: BoxZero},
			Series:         []Series{series},
		}

		buffer := &bytes.Buffer{}
		testutil.AssertNil(t, c.Render(provider, buffer))
		testutil.AssertNotZero(t, buffer.Len())
	}
}

func TestContinuousOHLCSeriesAxisRange(t *testing.T) {
	c := Chart{
		Height:         100,
		Width:          100,
		TitleStyle:     Hidden(),
		XAxis:          HideXAxis(),
		YAxis:          HideYAxis(),
		YAxisSecondary: HideYAxis(),
		Canvas:         Style{Padding: BoxZero},
		Background:     Style{Padding: BoxZero},
		Series:         []Series{validContinuousOHLCSeries()},
	}

	xr, yr, _ := c.getRanges()
	testutil.AssertEqual(t, 1.0, xr.GetMin())
	testutil.AssertEqual(t, 5.0, xr.GetMax())
	testutil.AssertTrue(t, yr.GetMin() <= 8.0)
	testutil.AssertTrue(t, yr.GetMax() >= 15.0)
}
