package chart

import (
	"fmt"

	"github.com/wcharczuk/go-chart/v2/drawing"
)

// Interface Assertions.
var (
	_ Series                  = (*ContinuousOHLCSeries)(nil)
	_ OHLCValuesProvider      = (*ContinuousOHLCSeries)(nil)
	_ BoundedValuesProvider   = (*ContinuousOHLCSeries)(nil)
	_ FirstValuesProvider     = (*ContinuousOHLCSeries)(nil)
	_ LastValuesProvider      = (*ContinuousOHLCSeries)(nil)
	_ ValueFormatterProvider  = (*ContinuousOHLCSeries)(nil)
	_ FirstOHLCValuesProvider = (*ContinuousOHLCSeries)(nil)
	_ OHLCLastValuesProvider  = (*ContinuousOHLCSeries)(nil)
	_ OHLCLegendProvider      = (*ContinuousOHLCSeries)(nil)
)

// ContinuousOHLCSeries is a series of open, high, low and close values
// positioned on a numeric (continuous) x axis rather than a time axis. It
// is the numeric-axis counterpart of TimeOHLCSeries and renders the same
// way: a high-low wick and an open-close body colored by direction.
//
// The x values must be ordered (ascending or descending) and unique.
type ContinuousOHLCSeries struct {
	Name  string
	Style Style

	YAxis YAxisType

	XValueFormatter ValueFormatter
	YValueFormatter ValueFormatter

	XValues     []float64
	OpenValues  []float64
	HighValues  []float64
	LowValues   []float64
	CloseValues []float64

	// CandleWidth is the body width in pixels; zero means auto-sized from
	// the distance to neighboring points.
	CandleWidth int

	UpColor   drawing.Color
	DownColor drawing.Color
	DojiColor drawing.Color
}

// GetName returns the name of the series.
func (ohlc ContinuousOHLCSeries) GetName() string { return ohlc.Name }

// GetStyle returns the series style.
func (ohlc ContinuousOHLCSeries) GetStyle() Style { return ohlc.Style }

// GetYAxis returns which YAxis the series draws on.
func (ohlc ContinuousOHLCSeries) GetYAxis() YAxisType { return ohlc.YAxis }

// Len returns the number of candles in the series.
func (ohlc ContinuousOHLCSeries) Len() int { return len(ohlc.XValues) }

// GetOHLCValues returns the x, open, high, low and close values at index.
func (ohlc ContinuousOHLCSeries) GetOHLCValues(index int) (x, open, high, low, close float64) {
	return ohlc.XValues[index],
		ohlc.OpenValues[index],
		ohlc.HighValues[index],
		ohlc.LowValues[index],
		ohlc.CloseValues[index]
}

// GetBoundedValues implements BoundedValuesProvider so the axis range
// includes the full high-low extent of every candle.
func (ohlc ContinuousOHLCSeries) GetBoundedValues(index int) (x, y1, y2 float64) {
	x, _, high, low, _ := ohlc.GetOHLCValues(index)
	return x, high, low
}

// GetFirstValues returns the first x value and its close price.
func (ohlc ContinuousOHLCSeries) GetFirstValues() (x, y float64) {
	x, _, _, _, close := ohlc.GetOHLCValues(0)
	return x, close
}

// GetLastValues returns the last x value and its close price.
func (ohlc ContinuousOHLCSeries) GetLastValues() (x, y float64) {
	x, _, _, _, close := ohlc.GetOHLCValues(ohlc.Len() - 1)
	return x, close
}

// GetFirstOHLCValues returns the first set of OHLC values.
func (ohlc ContinuousOHLCSeries) GetFirstOHLCValues() (x, open, high, low, close float64) {
	return ohlc.GetOHLCValues(0)
}

// GetLastOHLCValues returns the last set of OHLC values.
func (ohlc ContinuousOHLCSeries) GetLastOHLCValues() (x, open, high, low, close float64) {
	return ohlc.GetOHLCValues(ohlc.Len() - 1)
}

// GetValueFormatters returns value formatter defaults for the series.
func (ohlc ContinuousOHLCSeries) GetValueFormatters() (x, y ValueFormatter) {
	if ohlc.XValueFormatter != nil {
		x = ohlc.XValueFormatter
	} else {
		x = FloatValueFormatter
	}
	if ohlc.YValueFormatter != nil {
		y = ohlc.YValueFormatter
	} else {
		y = FloatValueFormatter
	}
	return
}

// GetUpColor returns the color used for up candles.
func (ohlc ContinuousOHLCSeries) GetUpColor() drawing.Color {
	if ohlc.UpColor.IsZero() {
		return DefaultCandleUpColor
	}
	return ohlc.UpColor
}

// GetDownColor returns the color used for down candles.
func (ohlc ContinuousOHLCSeries) GetDownColor() drawing.Color {
	if ohlc.DownColor.IsZero() {
		return DefaultCandleDownColor
	}
	return ohlc.DownColor
}

// GetDojiColor returns the color used for doji candles.
func (ohlc ContinuousOHLCSeries) GetDojiColor() drawing.Color {
	if ohlc.DojiColor.IsZero() {
		return DefaultCandleDojiColor
	}
	return ohlc.DojiColor
}

// CandleColor returns the color for the candle at index.
func (ohlc ContinuousOHLCSeries) CandleColor(index int) drawing.Color {
	open := ohlc.OpenValues[index]
	close := ohlc.CloseValues[index]
	switch {
	case close > open:
		return ohlc.GetUpColor()
	case close < open:
		return ohlc.GetDownColor()
	default:
		return ohlc.GetDojiColor()
	}
}

// Render renders the series.
func (ohlc ContinuousOHLCSeries) Render(r Renderer, canvasBox Box, xrange, yrange Range, defaults Style) {
	style := ohlc.Style.InheritFrom(defaults)
	Draw.OHLCSeries(r, canvasBox, xrange, yrange, style, ohlc, ohlc.CandleWidth)
}

// Validate validates the series.
func (ohlc ContinuousOHLCSeries) Validate() error {
	if len(ohlc.XValues) == 0 {
		return fmt.Errorf("ohlc series; must have xvalues set")
	}
	if len(ohlc.OpenValues) == 0 {
		return fmt.Errorf("ohlc series; must have open values set")
	}
	if len(ohlc.HighValues) == 0 {
		return fmt.Errorf("ohlc series; must have high values set")
	}
	if len(ohlc.LowValues) == 0 {
		return fmt.Errorf("ohlc series; must have low values set")
	}
	if len(ohlc.CloseValues) == 0 {
		return fmt.Errorf("ohlc series; must have close values set")
	}
	if len(ohlc.OpenValues) != len(ohlc.XValues) ||
		len(ohlc.HighValues) != len(ohlc.XValues) ||
		len(ohlc.LowValues) != len(ohlc.XValues) ||
		len(ohlc.CloseValues) != len(ohlc.XValues) {
		return fmt.Errorf("ohlc series; open, high, low and close values must have the same length as xvalues")
	}
	return ValidateOHLCSeries(ohlc)
}
