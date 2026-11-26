package chart

import (
	"fmt"
	"math"
	"time"

	"github.com/wcharczuk/go-chart/v2/drawing"
)

// Interface Assertions.
var (
	_ Series                  = (*TimeOHLCSeries)(nil)
	_ OHLCValuesProvider      = (*TimeOHLCSeries)(nil)
	_ BoundedValuesProvider   = (*TimeOHLCSeries)(nil)
	_ FirstValuesProvider     = (*TimeOHLCSeries)(nil)
	_ LastValuesProvider      = (*TimeOHLCSeries)(nil)
	_ ValueFormatterProvider  = (*TimeOHLCSeries)(nil)
	_ FirstOHLCValuesProvider = (*TimeOHLCSeries)(nil)
	_ OHLCLastValuesProvider  = (*TimeOHLCSeries)(nil)
)

// Default candle colors; an up candle (close > open) is green, a down
// candle (close < open) is red and a doji (close == open) is black.
var (
	// DefaultCandleUpColor is the fill and stroke color for up candles.
	DefaultCandleUpColor = ColorGreen
	// DefaultCandleDownColor is the fill and stroke color for down candles.
	DefaultCandleDownColor = ColorRed
	// DefaultCandleDojiColor is the fill and stroke color for doji candles.
	DefaultCandleDojiColor = ColorBlack
)

const (
	// DefaultCandleWidthRatio is the fraction of the distance to the nearest
	// neighboring point that a candle body occupies when the width is not
	// set explicitly.
	DefaultCandleWidthRatio = 0.65
	// DefaultSingleCandleWidthRatio is the fraction of the canvas width used
	// for a candle when it has no neighboring points.
	DefaultSingleCandleWidthRatio = 0.1
	// MinimumCandleWidth is the smallest renderable candle body width.
	MinimumCandleWidth = 1
)

// TimeOHLCSeries is a series of open, high, low and close values positioned
// on a time axis. It renders as candlesticks: a vertical wick spans the
// high-low range and a body spans the open-close range, colored according
// to whether the period closed up or down.
//
// The x values must be ordered (ascending or descending) and unique;
// Validate returns an error otherwise.
type TimeOHLCSeries struct {
	Name  string
	Style Style

	YAxis YAxisType

	// XValues are the ordered, unique timestamps for each period.
	XValues []time.Time
	// OpenValues, HighValues, LowValues and CloseValues hold the price
	// components for each period. They must all have the same length as
	// XValues.
	OpenValues  []float64
	HighValues  []float64
	LowValues   []float64
	CloseValues []float64

	// CandleWidth is the body width in pixels. When zero (the default) the
	// width is derived from the distance to neighboring points and the
	// canvas width, so non-uniform time spacing produces proportionally
	// sized candles.
	CandleWidth int

	// UpColor, DownColor and DojiColor override the default candle colors.
	// The zero value means the corresponding default is used.
	UpColor   drawing.Color
	DownColor drawing.Color
	DojiColor drawing.Color
}

// GetName returns the name of the series.
func (ohlc TimeOHLCSeries) GetName() string {
	return ohlc.Name
}

// GetStyle returns the series style.
func (ohlc TimeOHLCSeries) GetStyle() Style {
	return ohlc.Style
}

// GetYAxis returns which YAxis the series draws on.
func (ohlc TimeOHLCSeries) GetYAxis() YAxisType {
	return ohlc.YAxis
}

// Len returns the number of candles in the series.
func (ohlc TimeOHLCSeries) Len() int {
	return len(ohlc.XValues)
}

// GetOHLCValues returns the x, open, high, low and close values at index.
func (ohlc TimeOHLCSeries) GetOHLCValues(index int) (x, open, high, low, close float64) {
	x = TimeToFloat64(ohlc.XValues[index])
	open = ohlc.OpenValues[index]
	high = ohlc.HighValues[index]
	low = ohlc.LowValues[index]
	close = ohlc.CloseValues[index]
	return
}

// GetBoundedValues implements BoundedValuesProvider so the chart range
// includes the full high-low extent of every candle. y1 is the high and
// y2 is the low.
func (ohlc TimeOHLCSeries) GetBoundedValues(index int) (x, y1, y2 float64) {
	x, _, high, low, _ := ohlc.GetOHLCValues(index)
	y1 = high
	y2 = low
	return
}

// GetFirstValues returns the first x value and its close price.
func (ohlc TimeOHLCSeries) GetFirstValues() (x, y float64) {
	x, _, _, _, close := ohlc.GetOHLCValues(0)
	y = close
	return
}

// GetLastValues returns the last x value and its close price.
func (ohlc TimeOHLCSeries) GetLastValues() (x, y float64) {
	x, _, _, _, close := ohlc.GetOHLCValues(ohlc.Len() - 1)
	y = close
	return
}

// GetFirstOHLCValues returns the first set of OHLC values.
func (ohlc TimeOHLCSeries) GetFirstOHLCValues() (x, open, high, low, close float64) {
	return ohlc.GetOHLCValues(0)
}

// GetLastOHLCValues returns the last set of OHLC values.
func (ohlc TimeOHLCSeries) GetLastOHLCValues() (x, open, high, low, close float64) {
	return ohlc.GetOHLCValues(ohlc.Len() - 1)
}

// GetValueFormatters returns value formatter defaults for the series.
func (ohlc TimeOHLCSeries) GetValueFormatters() (x, y ValueFormatter) {
	x = TimeValueFormatter
	y = FloatValueFormatter
	return
}

// GetUpColor returns the color used for up candles.
func (ohlc TimeOHLCSeries) GetUpColor() drawing.Color {
	if ohlc.UpColor.IsZero() {
		return DefaultCandleUpColor
	}
	return ohlc.UpColor
}

// GetDownColor returns the color used for down candles.
func (ohlc TimeOHLCSeries) GetDownColor() drawing.Color {
	if ohlc.DownColor.IsZero() {
		return DefaultCandleDownColor
	}
	return ohlc.DownColor
}

// GetDojiColor returns the color used for doji candles.
func (ohlc TimeOHLCSeries) GetDojiColor() drawing.Color {
	if ohlc.DojiColor.IsZero() {
		return DefaultCandleDojiColor
	}
	return ohlc.DojiColor
}

// CandleColor returns the color for the candle at index based on its
// open and close values.
func (ohlc TimeOHLCSeries) CandleColor(index int) drawing.Color {
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
func (ohlc TimeOHLCSeries) Render(r Renderer, canvasBox Box, xrange, yrange Range, defaults Style) {
	style := ohlc.Style.InheritFrom(defaults)
	Draw.OHLCSeries(r, canvasBox, xrange, yrange, style, ohlc, ohlc.CandleWidth)
}

// Validate validates the series.
func (ohlc TimeOHLCSeries) Validate() error {
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

// ValidateOHLCSeries validates an OHLCValuesProvider: all slices must be
// present and equally long, x values must be finite, strictly ordered and
// unique, every o/h/l/c value must be finite, the high must be at or above
// the low, and both open and close must fall within the high-low range.
func ValidateOHLCSeries(vs OHLCValuesProvider) error {
	length := vs.Len()
	if length == 0 {
		return fmt.Errorf("ohlc series; must have at least one value")
	}

	prevX, prevOpen, prevHigh, prevLow, prevClose := vs.GetOHLCValues(0)
	if err := validateOHLCCandle(0, prevX, prevOpen, prevHigh, prevLow, prevClose); err != nil {
		return err
	}

	ascending := true
	descending := true
	for index := 1; index < length; index++ {
		x, open, high, low, close := vs.GetOHLCValues(index)
		if err := validateOHLCCandle(index, x, open, high, low, close); err != nil {
			return err
		}
		if x == prevX {
			return fmt.Errorf("ohlc series; x values must be unique, duplicate value at index %d", index)
		}
		if x < prevX {
			ascending = false
		}
		if x > prevX {
			descending = false
		}
		if !ascending && !descending {
			return fmt.Errorf("ohlc series; x values must be strictly ordered (ascending or descending), value at index %d is out of order", index)
		}
		prevX = x
	}
	return nil
}

// validateOHLCCandle validates a single candle's values.
func validateOHLCCandle(index int, x, open, high, low, close float64) error {
	if !finite(x) {
		return fmt.Errorf("ohlc series; x value at index %d must be finite", index)
	}
	for name, v := range map[string]float64{
		"open":  open,
		"high":  high,
		"low":   low,
		"close": close,
	} {
		if !finite(v) {
			return fmt.Errorf("ohlc series; %s value at index %d must be finite", name, index)
		}
	}
	if high < low {
		return fmt.Errorf("ohlc series; high (%.2f) must be greater than or equal to low (%.2f) at index %d", high, low, index)
	}
	if open < low || open > high {
		return fmt.Errorf("ohlc series; open (%.2f) must be within the high-low range [%.2f, %.2f] at index %d", open, low, high, index)
	}
	if close < low || close > high {
		return fmt.Errorf("ohlc series; close (%.2f) must be within the high-low range [%.2f, %.2f] at index %d", close, low, high, index)
	}
	return nil
}

// finite returns if a value is a usable, finite number.
func finite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}
