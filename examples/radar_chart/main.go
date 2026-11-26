package main

//go:generate go run main.go

import (
	"os"

	"github.com/wcharczuk/go-chart/v2"
)

func main() {
	rc := chart.RadarChart{
		Title:  "Character Stats",
		Width:  600,
		Height: 600,
		Dimensions: []chart.RadarDimension{
			{Name: "Attack", Max: 100},
			{Name: "Defense", Max: 100},
			{Name: "Speed", Max: 100},
			{Name: "Magic", Max: 100},
			{Name: "Luck", Max: 100},
		},
		Series: []chart.RadarSeries{
			{Name: "Warrior", Values: []float64{80, 90, 40, 20, 50}},
			{Name: "Mage", Values: []float64{30, 40, 60, 95, 70}},
			{Name: "Rogue", Values: []float64{50, 30, 95, 45, 80}},
		},
	}

	f, _ := os.Create("output.png")
	defer f.Close()
	rc.Render(chart.PNG, f)
}
