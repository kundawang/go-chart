package chart

import (
	"bytes"
	"testing"

	"github.com/wcharczuk/go-chart/v2/testutil"
)

func TestLegend(t *testing.T) {
	// replaced new assertions helper

	graph := Chart{
		Series: []Series{
			ContinuousSeries{
				Name:    "A test series",
				XValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
				YValues: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			},
		},
	}

	//note we have to do this as a separate step because we need a reference to graph
	graph.Elements = []Renderable{
		Legend(&graph),
	}
	buf := bytes.NewBuffer([]byte{})
	err := graph.Render(PNG, buf)
	testutil.AssertNil(t, err)
	testutil.AssertNotZero(t, buf.Len())
}

func TestLegendOHLCSwatch(t *testing.T) {
	series := TimeOHLCSeries{
		Name:        "Candles",
		XValues:     ohlcTestTimes(),
		OpenValues:  []float64{10, 14, 11, 11},
		HighValues:  []float64{14, 15, 13, 12},
		LowValues:   []float64{8, 10, 9, 10},
		CloseValues: []float64{12, 11, 11, 10},
	}

	graph := Chart{
		Series: []Series{series},
	}
	graph.Elements = []Renderable{Legend(&graph)}

	// SVG so the up/down swatch colors are directly inspectable.
	buf := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, graph.Render(SVG, buf))
	svg := buf.String()
	testutil.AssertContains(t, svg, series.GetUpColor().String())
	testutil.AssertContains(t, svg, series.GetDownColor().String())
	testutil.AssertContains(t, svg, "Candles")
}

func TestLegendThinOHLCSwatch(t *testing.T) {
	series := validOHLCSeries()
	graph := Chart{Series: []Series{series}}
	graph.Elements = []Renderable{LegendThin(&graph)}

	buf := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, graph.Render(SVG, buf))
	svg := buf.String()
	testutil.AssertContains(t, svg, series.GetUpColor().String())
	testutil.AssertContains(t, svg, series.GetDownColor().String())
}

func TestLegendLeftOHLCSwatch(t *testing.T) {
	series := validOHLCSeries()
	graph := Chart{Series: []Series{series}}
	graph.Elements = []Renderable{LegendLeft(&graph)}

	buf := bytes.NewBuffer([]byte{})
	testutil.AssertNil(t, graph.Render(SVG, buf))
	svg := buf.String()
	testutil.AssertContains(t, svg, series.GetUpColor().String())
	testutil.AssertContains(t, svg, series.GetDownColor().String())
}
