package chart

import (
	"io"

	"github.com/golang/freetype/truetype"
)

// StackedAreaChart is a convenience chart that renders a positive/negative
// separated stacked area. It builds a standard Chart internally, so SVG and
// PNG output share axes, ranges and legend behavior exactly.
type StackedAreaChart struct {
	Title      string
	TitleStyle Style

	ColorPalette ColorPalette

	Width  int
	Height int
	DPI    float64

	Background Style
	Canvas     Style

	XAxis XAxis
	YAxis YAxis

	Font        *truetype.Font
	defaultFont *truetype.Font

	// XValues is the shared increasing grid for matrix-style input.
	XValues []float64
	// X is a compatibility alias for XValues.
	X []float64
	// YValues and Values are the component matrix; one row per component.
	YValues [][]float64
	Values  [][]float64
	// Data and Rows are additional matrix spellings.
	Data [][]float64
	Rows [][]float64

	// Series/Layers are the component layer collections. They accept
	// []StackedAreaSeries, []ContinuousSeries, []chart.Series, [][]float64 or
	// a single layer.
	Series interface{}
	Layers interface{}

	// Components optionally holds ContinuousSeries layers.
	Components []ContinuousSeries

	// Names labels matrix rows for the legend.
	Names []string

	// ShowLegend appends a layer legend using the standard Legend renderable.
	ShowLegend bool

	Elements []Renderable
}

// GetDPI returns the dpi for the chart.
func (c StackedAreaChart) GetDPI(defaults ...float64) float64 {
	if c.DPI == 0 {
		if len(defaults) > 0 {
			return defaults[0]
		}
		return DefaultDPI
	}
	return c.DPI
}

// GetFont returns the text font.
func (c StackedAreaChart) GetFont() *truetype.Font {
	if c.Font == nil {
		return c.defaultFont
	}
	return c.Font
}

// GetWidth returns the chart width or the default value.
func (c StackedAreaChart) GetWidth() int {
	if c.Width == 0 {
		return DefaultChartWidth
	}
	return c.Width
}

// GetHeight returns the chart height or the default value.
func (c StackedAreaChart) GetHeight() int {
	if c.Height == 0 {
		return DefaultChartHeight
	}
	return c.Height
}

// AsChart converts the configuration into a standard Chart.
func (c StackedAreaChart) AsChart() *Chart {
	series := c.areaSeries()

	xAxis := c.XAxis
	yAxis := c.YAxis

	ch := &Chart{
		Title:        c.Title,
		TitleStyle:   c.TitleStyle,
		ColorPalette: c.ColorPalette,
		Width:        c.GetWidth(),
		Height:       c.GetHeight(),
		DPI:          c.DPI,
		Background:   c.Background,
		Canvas:       c.Canvas,
		XAxis:        xAxis,
		YAxis:        yAxis,
		Font:         c.Font,
		Series:       []Series{series},
	}
	if c.ShowLegend {
		ch.Elements = append(ch.Elements, Legend(ch))
	}
	ch.Elements = append(ch.Elements, c.Elements...)

	// A single grid point gives a zero X delta which the stock range checks
	// reject. Give the X axis a half-unit padding so the point still renders.
	if series.Len() == 1 && (ch.XAxis.Range == nil || ch.XAxis.Range.IsZero()) {
		x, _, _ := series.GetBoundedValues(0)
		ch.XAxis.Range = &ContinuousRange{Min: x - 0.5, Max: x + 0.5}
	}
	return ch
}

// Render renders the chart with the given renderer provider to the writer.
func (c StackedAreaChart) Render(rp RendererProvider, w io.Writer) error {
	ch := c.AsChart()
	if c.Font == nil {
		if f, err := GetDefaultFont(); err == nil {
			ch.Font = f
		}
	}
	return ch.Render(rp, w)
}

func (c StackedAreaChart) areaSeries() StackedAreaSeries {
	xs := c.XValues
	if len(xs) == 0 {
		xs = c.X
	}
	layers := c.chartLayers(c.Series)
	if len(layers) == 0 {
		layers = c.chartLayers(c.Layers)
	}
	if len(layers) == 0 && len(c.Components) > 0 {
		layers = c.chartLayers(c.Components)
	}
	if len(layers) > 0 {
		// Propagate the shared grid to layers that do not declare their own.
		if len(xs) > 0 {
			for i := range layers {
				if len(layers[i].XValues) == 0 && len(layers[i].X) == 0 {
					layers[i].XValues = xs
				}
			}
		}
		return StackedAreaSeries{
			Name:    c.Title,
			XValues: xs,
			Series:  layers,
		}
	}
	matrix := c.YValues
	if len(matrix) == 0 {
		matrix = c.Values
	}
	if len(matrix) == 0 {
		matrix = c.Data
	}
	if len(matrix) == 0 {
		matrix = c.Rows
	}
	return StackedAreaSeries{
		Name:    c.Title,
		XValues: xs,
		YValues: matrix,
		Names:   c.Names,
	}
}

// chartLayers normalizes any supported layer collection into concrete
// StackedAreaSeries layers.
func (c StackedAreaChart) chartLayers(v interface{}) []StackedAreaSeries {
	if v == nil {
		return nil
	}
	raw := (StackedAreaSeries{}).saCollect(v)
	out := make([]StackedAreaSeries, 0, len(raw))
	for _, l := range raw {
		out = append(out, StackedAreaSeries{
			Name:    l.name,
			Style:   l.style,
			XValues: l.x,
			YValues: l.y,
		})
	}
	return out
}

// StackedAreaLegend returns the standard legend, which expands the layers of
// any LegendEntriesProvider series (including StackedAreaSeries).
func StackedAreaLegend(c *Chart, userDefaults ...Style) Renderable {
	return Legend(c, userDefaults...)
}

// NewStackedAreaSeries builds a shared-grid StackedAreaSeries from variadic
// component rows. Each row must align to xs.
func NewStackedAreaSeries(name string, xs []float64, rows ...[]float64) *StackedAreaSeries {
	return &StackedAreaSeries{
		Name:    name,
		XValues: xs,
		YValues: rows,
	}
}

// NewStackedAreaChart builds a StackedAreaChart from a shared grid and a
// component matrix.
func NewStackedAreaChart(width, height int, xs []float64, rows [][]float64) *StackedAreaChart {
	return &StackedAreaChart{
		Width:   width,
		Height:  height,
		XValues: xs,
		YValues: rows,
	}
}

// RenderStackedArea is a one-call convenience that renders a positive/
// negative separated stacked area to w using rp.
func RenderStackedArea(rp RendererProvider, w io.Writer, width, height int, xs []float64, rows [][]float64) error {
	c := NewStackedAreaChart(width, height, xs, rows)
	return c.Render(rp, w)
}

// DivergingStackedAreaSeries is a naming alias for StackedAreaSeries.
type DivergingStackedAreaSeries = StackedAreaSeries

// PositiveNegativeStackedAreaSeries is a naming alias for StackedAreaSeries.
type PositiveNegativeStackedAreaSeries = StackedAreaSeries

// StackAreaSeries is a naming alias for StackedAreaSeries.
type StackAreaSeries = StackedAreaSeries

// AreaSeries is a single/stacked area series alias for compatibility.
type AreaSeries = StackedAreaSeries
