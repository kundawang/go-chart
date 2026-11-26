package chart

import "github.com/wcharczuk/go-chart/v2/drawing"

// OHLCValuesProvider is a type that produces open, high, low and close
// values for a given index. The x value is the position on the (ordered)
// horizontal axis, expressed as a float64 so both time and numeric axes
// are supported.
type OHLCValuesProvider interface {
	Len() int
	GetOHLCValues(index int) (x, open, high, low, close float64)
}

// FirstOHLCValuesProvider is a special type of value provider that can
// return its (potentially computed) first set of OHLC values.
type FirstOHLCValuesProvider interface {
	GetFirstOHLCValues() (x, open, high, low, close float64)
}

// OHLCLastValuesProvider is a special type of value provider that can
// return its (potentially computed) last set of OHLC values.
type OHLCLastValuesProvider interface {
	GetLastOHLCValues() (x, open, high, low, close float64)
}

// OHLCLegendProvider is implemented by series that render as candles and
// want a candle swatch rather than a plain line in the legend.
type OHLCLegendProvider interface {
	GetUpColor() drawing.Color
	GetDownColor() drawing.Color
}
