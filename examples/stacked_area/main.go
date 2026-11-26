package main

import (
	"os"

	"github.com/wcharczuk/go-chart/v2"
)

func main() {
	// Frame time per subsystem across frames. Render and script can run
	// "under budget" offsets (negative values) and the two sides must stack
	// independently rather than cancel.
	graph := chart.Chart{
		Title:  "Frame time by subsystem",
		Width:  960,
		Height: 480,
		XAxis: chart.XAxis{
			Name: "frame",
		},
		YAxis: chart.YAxis{
			Name: "ms",
		},
		Series: []chart.Series{
			chart.StackedAreaSeries{
				Name:    "frame time",
				XValues: []float64{0, 1, 2, 3, 4, 5, 6, 7},
				Series: []chart.StackedAreaSeries{
					{Name: "render", YValues: []float64{2, 3, 2, 4, 3, 5, 4, 6}},
					{Name: "script", YValues: []float64{1, -1, 2, 3, -2, 2, 3, 1}},
					{Name: "audio", YValues: []float64{0.4, 0.6, 0.3, 0.8, 0.5, 0.7, 0.4, 0.6}},
				},
			},
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	png, err := os.Create("output.png")
	if err != nil {
		panic(err)
	}
	if err := graph.Render(chart.PNG, png); err != nil {
		panic(err)
	}
	png.Close()

	svg, err := os.Create("output.svg")
	if err != nil {
		panic(err)
	}
	if err := graph.Render(chart.SVG, svg); err != nil {
		panic(err)
	}
	svg.Close()
}
