package chart

import (
	"fmt"

	"github.com/wcharczuk/go-chart/v2/drawing"
)

// Interface assertions.
var (
	_ Series                = (*StackedAreaSeries)(nil)
	_ BoundedValuesProvider = (*StackedAreaSeries)(nil)
	_ LegendEntriesProvider = (*StackedAreaSeries)(nil)
)

// LegendEntry is a single row in a chart legend for a multi-layer series.
type LegendEntry struct {
	Name  string
	Label string
	Text  string
	Style Style
}

// LegendEntriesProvider is implemented by series that expand into more than
// one legend entry (for example the layers of a stacked area chart).
type LegendEntriesProvider interface {
	LegendEntries(defaults ...Style) []LegendEntry
}

// StackedAreaSeries draws a positive/negative separated stacked area.
//
// Components are accumulated independently above and below the zero line at
// every shared X grid point: positives stack upward from zero, negatives
// stack downward from zero, and the two totals never cancel.
//
// The type accepts several equivalent input spellings to fit how callers
// assemble data:
//
//   - Set XValues and YValues to describe this series as a single layer.
//   - Set Series or Layers to a slice of StackedAreaSeries (or
//     ContinuousSeries) describing the component layers.
//   - Set YValues / Values / Data / Rows to a [][]float64 component matrix
//     aligned to XValues.
//
// YValues and the collection fields are intentionally permissive (interface{})
// so both []float64 layer rows and [][]float64 matrices are accepted.
type StackedAreaSeries struct {
	Name  string
	Style Style
	YAxis YAxisType

	// XValues is the shared, strictly increasing X grid.
	XValues []float64
	// X is a compatibility alias for XValues.
	X []float64

	// YValues is either this layer's []float64 values or a [][]float64 matrix
	// of component rows.
	YValues interface{}
	// Y is a compatibility vector field for a single layer.
	Y []float64

	// Series and Layers are the component layer collections. They accept
	// []StackedAreaSeries, []ContinuousSeries, [][]float64 or a single value.
	Series interface{}
	Layers interface{}

	// Components and Continuous hold ContinuousSeries layer collections.
	Components       []ContinuousSeries
	ContinuousSeries []ContinuousSeries

	// Matrix input spellings; each outer entry is one component row aligned
	// to XValues.
	Values  interface{}
	Data    interface{}
	Rows    interface{}
	Matrix  [][]float64
	YMatrix [][]float64
	Ys      [][]float64

	// YValuesByTime is a TIME-MAJOR matrix: YValuesByTime[t] holds the
	// component values at time/grid index t. It is transposed before stacking.
	YValuesByTime [][]float64

	// Names/Labels optionally name matrix layers; Styles optionally styles
	// them in order.
	Names  []string
	Labels []string
	Styles []Style

	// FillOpacity multiplies the layer fill alpha (0..255); zero uses a
	// library default.
	FillOpacity uint8

	// HideBoundaryLines suppresses the upper/lower boundary strokes. By
	// default both boundaries are stroked.
	HideBoundaryLines bool
	// ShowBoundaryLines is accepted for source compatibility; boundaries are
	// drawn by default and this only forces them on. Use HideBoundaryLines to
	// suppress.
	ShowBoundaryLines bool
	// ShowDots draws dots at every boundary point (mainly useful for sparse
	// or single point data).
	ShowDots bool
}

// GetName returns the series (container) name.
func (sas StackedAreaSeries) GetName() string { return sas.Name }

// GetStyle returns the container style.
func (sas StackedAreaSeries) GetStyle() Style { return sas.Style }

// GetYAxis returns which Y axis the series draws on.
func (sas StackedAreaSeries) GetYAxis() YAxisType { return sas.YAxis }

// Len implements BoundedValuesProvider and returns the grid length.
func (sas StackedAreaSeries) Len() int {
	xs, _, _ := sas.saPrepared()
	return len(xs)
}

// GetBoundedValues returns the global (x, lower, upper) cumulative envelope
// at index so the chart Y range includes the cumulative extrema.
func (sas StackedAreaSeries) GetBoundedValues(index int) (x, y1, y2 float64) {
	res := sas.strictOrPadded()
	if res == nil || index < 0 || index >= len(res.X) {
		return 0, 0, 0
	}
	return res.X[index], res.Lower[index], res.Upper[index]
}

// Validate validates the shared grid and every layer.
func (sas StackedAreaSeries) Validate() error {
	layers := sas.saLayers()
	if len(layers) == 0 {
		return errStackedAreaNoLayers
	}
	xs, err := sas.saGrid(layers)
	if err != nil {
		return err
	}
	if len(xs) == 0 {
		return errStackedAreaEmptyGrid
	}
	rows := make([][]float64, len(layers))
	for i, l := range layers {
		if len(l.y) != len(xs) {
			return saErrLengthMismatch(i, len(l.y), len(xs))
		}
		rows[i] = l.y
	}
	_, _, _, _, err = saAccumulate(xs, rows)
	return err
}

// Stacks returns the cumulative envelope points (one per grid index). It is
// the primary computed accessor and never returns an error for malformed
// input; use Validate or Compute for strict error reporting.
func (sas StackedAreaSeries) Stacks() []StackedAreaPoint {
	res := sas.strictOrPadded()
	if res == nil {
		return nil
	}
	return res.Points()
}

// Bounds is an alias for Stacks.
func (sas StackedAreaSeries) Bounds() []StackedAreaPoint { return sas.Stacks() }

// StackedBounds is an alias for Stacks.
func (sas StackedAreaSeries) StackedBounds() []StackedAreaPoint { return sas.Stacks() }

// Points is an alias for Stacks.
func (sas StackedAreaSeries) Points() []StackedAreaPoint { return sas.Stacks() }

// GetBounds is an alias for Stacks.
func (sas StackedAreaSeries) GetBounds() []StackedAreaPoint { return sas.Stacks() }

// Boundaries is an alias for Stacks.
func (sas StackedAreaSeries) Boundaries() []StackedAreaPoint { return sas.Stacks() }

// CumulativeBounds is an alias for Stacks.
func (sas StackedAreaSeries) CumulativeBounds() []StackedAreaPoint { return sas.Stacks() }

// GetStacks is an alias for Stacks.
func (sas StackedAreaSeries) GetStacks() []StackedAreaPoint { return sas.Stacks() }

// StackPoints is an alias for Stacks.
func (sas StackedAreaSeries) StackPoints() []StackedAreaPoint { return sas.Stacks() }

// Envelope returns the global upper and lower cumulative slices.
func (sas StackedAreaSeries) Envelope() (upper, lower []float64) {
	res := sas.strictOrPadded()
	if res == nil {
		return nil, nil
	}
	return res.Upper, res.Lower
}

// UpperBounds returns the positive cumulative upper line.
func (sas StackedAreaSeries) UpperBounds() []float64 {
	u, _ := sas.Envelope()
	return u
}

// LowerBounds returns the negative cumulative lower line.
func (sas StackedAreaSeries) LowerBounds() []float64 {
	_, l := sas.Envelope()
	return l
}

// Totals returns the global (upper, lower) cumulative slices.
func (sas StackedAreaSeries) Totals() (upper, lower []float64) { return sas.Envelope() }

// UpperAt returns the positive cumulative upper boundary at grid index.
func (sas StackedAreaSeries) UpperAt(index int) float64 {
	res := sas.strictOrPadded()
	if res == nil {
		return 0
	}
	return res.UpperAt(index)
}

// LowerAt returns the negative cumulative lower boundary at grid index.
func (sas StackedAreaSeries) LowerAt(index int) float64 {
	res := sas.strictOrPadded()
	if res == nil {
		return 0
	}
	return res.LowerAt(index)
}

// BoundAt returns the cumulative (lower, upper) boundary pair at index.
func (sas StackedAreaSeries) BoundAt(index int) (lower, upper float64) {
	res := sas.strictOrPadded()
	if res == nil {
		return 0, 0
	}
	return res.BoundAt(index)
}

// Extrema returns the cumulative global (min, max) extrema.
func (sas StackedAreaSeries) Extrema() (min, max float64) {
	res := sas.strictOrPadded()
	if res == nil {
		return 0, 0
	}
	return res.Min, res.Max
}

// BoundsMatrix returns the global [lower, upper] pair per grid index.
func (sas StackedAreaSeries) BoundsMatrix() [][2]float64 {
	res := sas.strictOrPadded()
	if res == nil {
		return nil
	}
	return res.BoundsMatrix()
}

// StackedValues returns every layer's [lower, upper] pair per grid index.
func (sas StackedAreaSeries) StackedValues() [][][2]float64 {
	res := sas.strictOrPadded()
	if res == nil {
		return nil
	}
	return res.StackedValues()
}

// GetStackedValues is an alias for StackedValues.
func (sas StackedAreaSeries) GetStackedValues() [][][2]float64 { return sas.StackedValues() }

// StackMatrix is an alias for StackedValues.
func (sas StackedAreaSeries) StackMatrix() [][][2]float64 { return sas.StackedValues() }

// LayerMatrix is an alias for StackedValues.
func (sas StackedAreaSeries) LayerMatrix() [][][2]float64 { return sas.StackedValues() }

// LayerBounds returns every layer's boundary lines.
func (sas StackedAreaSeries) LayerBounds() []StackedAreaLayerBounds {
	res := sas.strictOrPadded()
	if res == nil {
		return nil
	}
	return res.Layers
}

// StackedLayers is an alias for LayerBounds.
func (sas StackedAreaSeries) StackedLayers() []StackedAreaLayerBounds { return sas.LayerBounds() }

// Compute strictly accumulates the series and returns the full result or an
// error for malformed input.
func (sas StackedAreaSeries) Compute() (*StackedAreaResult, error) {
	layers := sas.saLayers()
	xs, err := sas.saGrid(layers)
	if err != nil {
		return nil, err
	}
	rows := make([][]float64, len(layers))
	names := make([]string, len(layers))
	for i, l := range layers {
		if len(l.y) != len(xs) {
			return nil, saErrLengthMismatch(i, len(l.y), len(xs))
		}
		rows[i] = l.y
		names[i] = l.name
	}
	u, lo, ll, lu, err := saAccumulate(xs, rows)
	if err != nil {
		return nil, err
	}
	return saBuildResult(xs, rows, u, lo, ll, lu, names), nil
}

// CheckedStacks is the error-returning variant of Stacks.
func (sas StackedAreaSeries) CheckedStacks() ([]StackedAreaPoint, error) {
	res, err := sas.Compute()
	if err != nil {
		return nil, err
	}
	return res.Points(), nil
}

// TryStacks is an alias for CheckedStacks.
func (sas StackedAreaSeries) TryStacks() ([]StackedAreaPoint, error) { return sas.CheckedStacks() }

// LegendEntries implements LegendEntriesProvider. When no default Style is
// supplied the zero style is used (palette colors still apply).
func (sas StackedAreaSeries) LegendEntries(defaults ...Style) []LegendEntry {
	var def Style
	if len(defaults) > 0 {
		def = defaults[0]
	}
	layers := sas.saLayers()
	entries := make([]LegendEntry, 0, len(layers))
	for i, l := range layers {
		if l.style.Hidden {
			continue
		}
		name := l.name
		entries = append(entries, LegendEntry{
			Name:  name,
			Label: name,
			Text:  name,
			Style: sas.saLayerStyle(i, l.style, def),
		})
	}
	return entries
}

// GetLegendEntries is a no-argument alias for LegendEntries().
func (sas StackedAreaSeries) GetLegendEntries() []LegendEntry {
	return sas.LegendEntries()
}

// Render implements Series and draws every layer band and its boundaries.
func (sas StackedAreaSeries) Render(r Renderer, canvasBox Box, xrange, yrange Range, defaults Style) {
	res := sas.strictOrPadded()
	if res == nil || len(res.X) == 0 {
		return
	}
	layers := sas.saLayers()

	cb, cl := canvasBox.Bottom, canvasBox.Left

	n := len(res.X)
	px := make([]int, n)
	pyUpper := make([][]int, len(res.Layers))
	pyLower := make([][]int, len(res.Layers))
	for i := 0; i < n; i++ {
		px[i] = cl + xrange.Translate(res.X[i])
	}
	for li := range res.Layers {
		pyUpper[li] = make([]int, n)
		pyLower[li] = make([]int, n)
		for i := 0; i < n; i++ {
			pyUpper[li][i] = cb - yrange.Translate(res.Layers[li].Upper[i])
			pyLower[li][i] = cb - yrange.Translate(res.Layers[li].Lower[i])
		}
	}

	drawLines := !sas.HideBoundaryLines || sas.ShowBoundaryLines

	for li := range res.Layers {
		var layerStyle Style
		if li < len(layers) {
			layerStyle = sas.saLayerStyle(li, layers[li].style, defaults)
		} else {
			layerStyle = sas.saLayerStyle(li, Style{}, defaults)
		}

		// Fill the band: upper edge forward, lower edge backward.
		fill := layerStyle
		if fill.FillColor.IsZero() {
			fill.FillColor = layerStyle.GetStrokeColor(DefaultColorPalette.GetSeriesColor(li))
		}
		alpha := sas.FillOpacity
		if alpha == 0 {
			alpha = 90
		}
		fill.FillColor = fill.FillColor.WithAlpha(alpha)
		fill.GetFillOptions().WriteDrawingOptionsToRenderer(r)

		if n == 1 {
			// Degenerate band: paint a visible vertical segment between the
			// two boundaries so single-point data has a defined render result.
			x := px[0]
			stroke := layerStyle.GetStrokeOptions()
			if stroke.StrokeWidth == 0 {
				stroke.StrokeWidth = DefaultSeriesLineWidth
			}
			stroke.WriteDrawingOptionsToRenderer(r)
			r.MoveTo(x, pyUpper[li][0])
			r.LineTo(x, pyLower[li][0])
			r.Stroke()
			r.Circle(saSinglePointRadius, x, pyUpper[li][0])
			r.FillStroke()
			r.Circle(saSinglePointRadius, x, pyLower[li][0])
			r.FillStroke()
		} else {
			r.MoveTo(px[0], pyUpper[li][0])
			for i := 1; i < n; i++ {
				r.LineTo(px[i], pyUpper[li][i])
			}
			for i := n - 1; i >= 0; i-- {
				r.LineTo(px[i], pyLower[li][i])
			}
			r.Close()
			r.Fill()
		}

		if drawLines && n > 1 {
			stroke := layerStyle.GetStrokeOptions()
			if stroke.StrokeWidth == 0 {
				stroke.StrokeWidth = DefaultSeriesLineWidth
			}
			stroke.WriteDrawingOptionsToRenderer(r)

			r.MoveTo(px[0], pyUpper[li][0])
			for i := 1; i < n; i++ {
				r.LineTo(px[i], pyUpper[li][i])
			}
			r.Stroke()

			r.MoveTo(px[0], pyLower[li][0])
			for i := 1; i < n; i++ {
				r.LineTo(px[i], pyLower[li][i])
			}
			r.Stroke()
		}

		if sas.ShowDots || n == 1 {
			dot := layerStyle.GetDotOptions()
			dot.WriteDrawingOptionsToRenderer(r)
			dotWidth := layerStyle.GetDotWidth()
			for i := 0; i < n; i++ {
				r.Circle(dotWidth, px[i], pyUpper[li][i])
				r.FillStroke()
				r.Circle(dotWidth, px[i], pyLower[li][i])
				r.FillStroke()
			}
		}
	}
}

// saLayerStyle resolves the style for layer index. Explicit layer styles win
// over the container style, which wins over the per-index palette color; the
// chart defaults mainly contribute font metrics.
func (sas StackedAreaSeries) saLayerStyle(index int, layer Style, defaults Style) Style {
	paletteColor := DefaultColorPalette.GetSeriesColor(index)

	stroke := firstNonZeroColor(layer.StrokeColor, sas.Style.StrokeColor, paletteColor)
	fill := firstNonZeroColor(layer.FillColor, sas.Style.FillColor, paletteColor)
	dot := firstNonZeroColor(layer.DotColor, sas.Style.DotColor, stroke)

	merged := layer.InheritFrom(sas.Style).InheritFrom(defaults)
	merged.StrokeColor = stroke
	merged.FillColor = fill
	merged.DotColor = dot
	if merged.StrokeWidth == 0 {
		merged.StrokeWidth = DefaultSeriesLineWidth
	}
	return merged
}

func firstNonZeroColor(colors ...drawing.Color) drawing.Color {
	for _, c := range colors {
		if !c.IsZero() {
			return c
		}
	}
	return drawing.Color{}
}

// saPrepared resolves grid + padded rows for provider use.
func (sas StackedAreaSeries) saPrepared() ([]float64, [][]float64, []saLayer) {
	layers := sas.saLayers()
	if len(layers) == 0 {
		return nil, nil, nil
	}
	xs, err := sas.saGrid(layers)
	if err != nil || len(xs) == 0 {
		return nil, nil, layers
	}
	rows := make([][]float64, len(layers))
	for i, l := range layers {
		y := make([]float64, len(xs))
		for j := 0; j < len(xs) && j < len(l.y); j++ {
			if isFiniteFloat(l.y[j]) {
				y[j] = l.y[j]
			}
		}
		rows[i] = y
	}
	return xs, rows, layers
}

func (sas StackedAreaSeries) strictOrPadded() *StackedAreaResult {
	xs, rows, layers := sas.saPrepared()
	if len(xs) == 0 || len(rows) == 0 {
		return nil
	}
	u, lo, ll, lu, err := saAccumulate(xs, rows)
	if err != nil {
		return nil
	}
	names := make([]string, len(layers))
	for i, l := range layers {
		names[i] = l.name
	}
	return saBuildResult(xs, rows, u, lo, ll, lu, names)
}

type saLayer struct {
	name  string
	style Style
	x     []float64
	y     []float64
}

func (sas StackedAreaSeries) saLayers() []saLayer {
	var layers []saLayer

	// 1. Permissive collection fields.
	for _, col := range []interface{}{sas.Series, sas.Layers} {
		layers = append(layers, sas.saCollect(col)...)
	}
	if len(layers) > 0 {
		return sas.saWithMeta(layers)
	}

	// 2. Strong ContinuousSeries collections.
	for _, col := range [][]ContinuousSeries{sas.ContinuousSeries, sas.Components} {
		for _, cs := range col {
			layers = append(layers, saLayer{name: cs.Name, style: cs.Style, x: cs.XValues, y: cs.YValues})
		}
	}
	if len(layers) > 0 {
		return sas.saWithMeta(layers)
	}

	// 3. Explicit time-major matrix, transposed to component rows.
	if len(sas.YValuesByTime) > 0 {
		for _, row := range saTransposeTimeRows(sas.YValuesByTime) {
			layers = append(layers, saLayer{x: sas.XValues, y: row})
		}
		return sas.saWithMeta(layers)
	}

	// 4. Matrix spellings (component-major: one row per component).
	for _, m := range []interface{}{sas.YValues, sas.Values, sas.Data, sas.Rows, sas.Matrix, sas.YMatrix, sas.Ys} {
		if rows := saCoerceMatrix(m); len(rows) > 0 {
			for i, row := range rows {
				name := sas.saMatrixName(i)
				st := Style{}
				if i < len(sas.Styles) {
					st = sas.Styles[i]
				}
				layers = append(layers, saLayer{name: name, style: st, x: sas.XValues, y: row})
			}
			return layers
		}
	}

	// 4. Own vector values as a single layer.
	if y := saOwnFloatSlice(sas); len(y) > 0 {
		x := sas.XValues
		if len(x) == 0 {
			x = sas.X
		}
		return []saLayer{{name: sas.Name, style: sas.Style, x: x, y: y}}
	}

	return sas.saWithMeta(layers)
}

func (sas StackedAreaSeries) saMatrixName(i int) string {
	if i < len(sas.Names) && sas.Names[i] != "" {
		return sas.Names[i]
	}
	if i < len(sas.Labels) && sas.Labels[i] != "" {
		return sas.Labels[i]
	}
	return ""
}

func (sas StackedAreaSeries) saWithMeta(layers []saLayer) []saLayer {
	for i := range layers {
		if layers[i].name == "" {
			layers[i].name = sas.saMatrixName(i)
		}
		if layers[i].style.IsZero() && i < len(sas.Styles) {
			layers[i].style = sas.Styles[i]
		}
		if len(layers[i].x) == 0 {
			if len(sas.XValues) > 0 {
				layers[i].x = sas.XValues
			} else {
				layers[i].x = sas.X
			}
		}
	}
	return layers
}

func (sas StackedAreaSeries) saCollect(v interface{}) []saLayer {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case StackedAreaSeries:
		x := t.XValues
		if len(x) == 0 {
			x = t.X
		}
		return []saLayer{{name: t.Name, style: t.Style, x: x, y: saOwnFloatSlice(t)}}
	case []StackedAreaSeries:
		out := make([]saLayer, 0, len(t))
		for _, s := range t {
			out = append(out, sas.saCollect(s)...)
		}
		return out
	case ContinuousSeries:
		return []saLayer{{name: t.Name, style: t.Style, x: t.XValues, y: t.YValues}}
	case []ContinuousSeries:
		out := make([]saLayer, 0, len(t))
		for _, c := range t {
			out = append(out, sas.saCollect(c)...)
		}
		return out
	case []Series:
		out := make([]saLayer, 0, len(t))
		for _, s := range t {
			out = append(out, sas.saCollect(s)...)
		}
		return out
	case []float64:
		return []saLayer{{y: t}}
	case [][]float64:
		out := make([]saLayer, 0, len(t))
		for _, row := range t {
			out = append(out, saLayer{y: row})
		}
		return out
	}
	if vp, ok := v.(ValuesProvider); ok {
		l := vp.Len()
		xs := make([]float64, l)
		ys := make([]float64, l)
		for i := 0; i < l; i++ {
			xs[i], ys[i] = vp.GetValues(i)
		}
		name := ""
		if np, ok := v.(NameProvider); ok {
			name = np.GetName()
		}
		return []saLayer{{name: name, x: xs, y: ys}}
	}
	if rows := saCoerceMatrix(v); rows != nil {
		out := make([]saLayer, 0, len(rows))
		for _, row := range rows {
			out = append(out, saLayer{y: row})
		}
		return out
	}
	return nil
}

// saOwnFloatSlice resolves a single layer's vector from any of the supported
// value field spellings.
func saOwnFloatSlice(t StackedAreaSeries) []float64 {
	if len(t.Y) > 0 {
		return t.Y
	}
	for _, v := range []interface{}{t.YValues, t.Values, t.Data, t.Rows} {
		if y := saAnyToFloatSlice(v); len(y) > 0 {
			return y
		}
	}
	return saAnyToFloatSlice(t.YValues)
}

func (sas StackedAreaSeries) saGrid(layers []saLayer) ([]float64, error) {
	if len(sas.XValues) > 0 {
		return sas.XValues, saCheckGrid(sas.XValues)
	}
	if len(sas.X) > 0 {
		return sas.X, saCheckGrid(sas.X)
	}
	for _, l := range layers {
		if len(l.x) > 0 {
			if err := saCheckGrid(l.x); err != nil {
				return nil, err
			}
			// X grid consistency: every layer that declares X must match.
			for _, other := range layers {
				if len(other.x) > 0 && !saSameFloats(l.x, other.x) {
					return nil, errStackedAreaGridMismatch
				}
			}
			return l.x, nil
		}
	}
	n := 0
	for _, l := range layers {
		if len(l.y) > n {
			n = len(l.y)
		}
	}
	xs := make([]float64, n)
	for i := range xs {
		xs[i] = float64(i)
	}
	return xs, nil
}

func saCheckGrid(xs []float64) error {
	if len(xs) == 0 {
		return errStackedAreaEmptyGrid
	}
	for i, x := range xs {
		if !isFiniteFloat(x) {
			return saErrGridNonFinite(i)
		}
		if i > 0 && xs[i-1] >= x {
			return saErrGridNotIncreasing(i)
		}
	}
	return nil
}

func saSameFloats(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func saCoerceMatrix(v interface{}) [][]float64 {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case [][]float64:
		return t
	case []float64:
		// A bare vector is not a matrix spelling here; callers use YValues.
		return nil
	}
	rows := saCoerceValueToRows(v)
	// saCoerceValueToRows wraps a bare []float64 as one row; for matrix
	// spellings that is acceptable too.
	return rows
}

// saSinglePointRadius is the marker radius used for single point stacks.
const saSinglePointRadius = 2.5

var (
	errStackedAreaNoLayers     = saError("stacked area; please provide at least one series/layer")
	errStackedAreaEmptyGrid    = saError("stacked area; x grid must not be empty")
	errStackedAreaGridMismatch = saError("stacked area; layer x grids do not match")
)

type saError string

func (e saError) Error() string { return string(e) }

func saErrLengthMismatch(layer, got, want int) error {
	return fmt.Errorf("stacked area; layer %d has %d values for %d x grid points", layer, got, want)
}
func saErrGridNonFinite(i int) error {
	return fmt.Errorf("stacked area; x grid value at index %d is not finite", i)
}
func saErrGridNotIncreasing(i int) error {
	return fmt.Errorf("stacked area; x grid must be strictly increasing at index %d", i)
}
