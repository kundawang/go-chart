package main

//go:generate go run main.go

import (
	"os"
	"time"

	"github.com/wcharczuk/go-chart/v2"
)

func main() {
	dates := xvalues()

	candles := chart.TimeOHLCSeries{
		Name: "SPY",
		XValues: dates,
		OpenValues:  []float64{212.47, 210.80, 209.30, 207.90, 208.70, 206.50, 204.90, 205.90, 207.10, 209.00},
		HighValues:  []float64{213.20, 212.00, 210.50, 209.10, 210.00, 208.00, 206.00, 207.40, 209.40, 211.00},
		LowValues:   []float64{211.30, 210.00, 207.50, 206.20, 207.80, 204.80, 203.10, 204.50, 206.40, 208.20},
		CloseValues: []float64{210.90, 209.40, 208.00, 208.60, 206.60, 205.00, 205.70, 207.20, 208.90, 210.40},
	}

	// a closing-price line overlaid on the candles for context.
	closeLine := chart.TimeSeries{
		Name: "SPY - Close",
		Style: chart.Style{
			StrokeColor:     chart.ColorAlternateBlue,
			StrokeWidth:     1.0,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		XValues: dates,
		YValues: candles.CloseValues,
	}

	graph := chart.Chart{
		Title: "SPY OHLC",
		XAxis: chart.XAxis{
			TickPosition: chart.TickPositionBetweenTicks,
		},
		YAxis: chart.YAxis{},
		Series: []chart.Series{
			candles,
			closeLine,
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	pngFile, err := os.Create("output.png")
	if err != nil {
		panic(err)
	}
	defer pngFile.Close()
	if err := graph.Render(chart.PNG, pngFile); err != nil {
		panic(err)
	}

	svgFile, err := os.Create("output.svg")
	if err != nil {
		panic(err)
	}
	defer svgFile.Close()
	if err := graph.Render(chart.SVG, svgFile); err != nil {
		panic(err)
	}
}

func xvalues() []time.Time {
	raw := []string{
		"2016-07-01", "2016-07-05", "2016-07-06", "2016-07-07", "2016-07-08",
		"2016-07-11", "2016-07-12", "2016-07-15", "2016-07-19", "2016-07-20",
	}
	var dates []time.Time
	for _, ts := range raw {
		parsed, err := time.Parse(chart.DefaultDateFormat, ts)
		if err != nil {
			panic(err)
		}
		dates = append(dates, parsed)
	}
	return dates
}
