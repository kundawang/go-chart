package main

//go:generate go run main.go

import (
	"os"

	"github.com/wcharczuk/go-chart/v2"
)

func main() {
	/*
	   In this example we set the primary YAxis to a symmetric logarithmic
	   range. Unlike LogarithmicRange it accepts negative values and keeps a
	   true linear region around zero, so positive and negative amplitudes
	   (for example an audio waveform with a large dynamic range) share one
	   continuous axis and the zero line stays centered.

	   With Min=-9, Max=9, Threshold=1 and a 200px tall plot area the values
	   -9, 0 and 9 translate to the domain offsets 0, 100 and 200.
	*/

	graph := chart.Chart{
		Background: chart.Style{
			Padding: chart.Box{
				Top:  20,
				Left: 20,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    "A test series",
				XValues: []float64{1, 2, 3, 4, 5, 6, 7},
				YValues: []float64{-9000, -3, -1, 0, 1, 3, 9000},
			},
		},
		YAxis: chart.YAxis{
			Style:     chart.Shown(),
			NameStyle: chart.Shown(),
			Range: &chart.SymmetricLogRange{
				Min:       -10000,
				Max:       10000,
				Threshold: 1,
			},
		},
	}

	f, _ := os.Create("output.png")
	defer f.Close()
	if err := graph.Render(chart.PNG, f); err != nil {
		panic(err)
	}
}
