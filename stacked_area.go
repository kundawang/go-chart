package chart

import (
	"errors"
	"fmt"
	"math"
	"reflect"
)

// StackedAreaPoint is the cumulative positive/negative envelope of a whole
// stack at a single shared X grid position.
//
// Upper is the top of the independently accumulated positive stack (the sum
// of all positive components at the point) and Lower is the bottom of the
// independently accumulated negative stack (the sum of all negative
// components). Positive and negative values never cancel each other.
type StackedAreaPoint struct {
	// Index is the position of the point in the shared X grid.
	Index int
	// X is the shared grid value.
	X float64

	// Lower is the lower cumulative boundary (negative stack bottom).
	Lower float64
	// Upper is the upper cumulative boundary (positive stack top).
	Upper float64

	// Bottom and Top are aliases for Lower and Upper respectively.
	Bottom float64
	Top    float64

	// Min and Max are aliases for Lower and Upper respectively.
	Min float64
	Max float64

	// Negative is the total negative accumulation and Positive is the total
	// positive accumulation at the point.
	Negative float64
	Positive float64

	// D3-style / baseline aliases: Y0/Base equal Lower and Y1 equals Upper.
	Y0   float64
	Y1   float64
	Base float64
}

// BoundAt returns the (Lower, Upper) cumulative boundaries of the point.
func (p StackedAreaPoint) BoundAt() (lower, upper float64) {
	return p.Lower, p.Upper
}

// StackedAreaBand is the lower/upper boundary of a single layer at one grid
// position. The band always satisfies Lower <= Upper.
type StackedAreaBand struct {
	Index int
	X     float64
	Lower float64
	Upper float64

	Bottom float64
	Top    float64
	Min    float64
	Max    float64
}

// StackedAreaLayerBounds holds one layer's lower and upper boundary lines
// across the shared X grid.
type StackedAreaLayerBounds struct {
	Name  string
	Lower []float64
	Upper []float64
}

// StackedAreaResult is the fully accumulated output of a positive/negative
// separated stack over a shared increasing X grid.
type StackedAreaResult struct {
	// X is the shared increasing grid.
	X []float64
	// Upper is the global positive cumulative top per grid index.
	Upper []float64
	// Lower is the global negative cumulative bottom per grid index.
	Lower []float64
	// Positive and Negative alias Upper and Lower.
	Positive []float64
	Negative []float64
	// Layers holds every layer's own lower/upper boundary lines.
	Layers []StackedAreaLayerBounds
	// Bands holds every layer's per-point bands, layer first then grid index.
	Bands [][]StackedAreaBand
	// Min and Max are the cumulative extrema across the whole grid.
	Min float64
	Max float64

	// Additional slice-name aliases for the cumulative lines. All alias
	// fields share the same underlying slices as Upper/Lower/Positive/Negative.
	Pos         []float64
	Neg         []float64
	Tops        []float64
	Bottoms     []float64
	High        []float64
	Low         []float64
	UpperBounds []float64
	LowerBounds []float64
}

// Len returns the number of grid points.
func (r *StackedAreaResult) Len() int {
	if r == nil {
		return 0
	}
	return len(r.X)
}

// Point returns the envelope point at the given grid index.
func (r *StackedAreaResult) Point(index int) StackedAreaPoint {
	lo, hi := r.Lower[index], r.Upper[index]
	return StackedAreaPoint{
		Index:    index,
		X:        r.X[index],
		Lower:    lo,
		Upper:    hi,
		Bottom:   lo,
		Top:      hi,
		Min:      lo,
		Max:      hi,
		Negative: lo,
		Positive: hi,
		Y0:       lo,
		Y1:       hi,
		Base:     lo,
	}
}

// Points returns the envelope as a slice of StackedAreaPoint.
func (r *StackedAreaResult) Points() []StackedAreaPoint {
	if r == nil {
		return nil
	}
	out := make([]StackedAreaPoint, len(r.X))
	for i := range r.X {
		out[i] = r.Point(i)
	}
	return out
}

// BoundsMatrix returns the global [lower, upper] boundary pair per grid index.
func (r *StackedAreaResult) BoundsMatrix() [][2]float64 {
	if r == nil {
		return nil
	}
	out := make([][2]float64, len(r.X))
	for i := range r.X {
		out[i] = [2]float64{r.Lower[i], r.Upper[i]}
	}
	return out
}

// StackedValues returns every layer's [lower, upper] pair per grid index:
// stackedValues[layer][index] = [lower, upper].
func (r *StackedAreaResult) StackedValues() [][][2]float64 {
	if r == nil {
		return nil
	}
	out := make([][][2]float64, len(r.Layers))
	for li := range r.Layers {
		out[li] = make([][2]float64, len(r.X))
		for i := range r.X {
			out[li][i] = [2]float64{r.Layers[li].Lower[i], r.Layers[li].Upper[i]}
		}
	}
	return out
}

// LayerMatrix is an alias for StackedValues.
func (r *StackedAreaResult) LayerMatrix() [][][2]float64 { return r.StackedValues() }

// Envelope returns the global upper and lower cumulative slices.
func (r *StackedAreaResult) Envelope() (upper, lower []float64) {
	if r == nil {
		return nil, nil
	}
	return r.Upper, r.Lower
}

// UpperAt returns the positive cumulative upper boundary at index.
func (r *StackedAreaResult) UpperAt(index int) float64 { return r.Upper[index] }

// LowerAt returns the negative cumulative lower boundary at index.
func (r *StackedAreaResult) LowerAt(index int) float64 { return r.Lower[index] }

// BoundAt returns the (lower, upper) cumulative boundaries at index.
func (r *StackedAreaResult) BoundAt(index int) (lower, upper float64) {
	return r.Lower[index], r.Upper[index]
}

// StackedArea is a compatibility alias for StackedAreaResult.
type StackedArea = StackedAreaResult

// StackedAreaBounds is a compatibility alias for StackedAreaPoint.
type StackedAreaBounds = StackedAreaPoint

// StackPoint is a compatibility alias for StackedAreaPoint.
type StackPoint = StackedAreaPoint

// StackBounds is a compatibility alias for StackedAreaPoint.
type StackBounds = StackedAreaPoint

// saAccumulate performs the canonical positive/negative separated
// accumulation. xs must be a finite, strictly increasing grid of length n and
// every row must have length n. It returns the global upper/lower lines and
// each layer's lower/upper boundary lines.
func saAccumulate(xs []float64, rows [][]float64) (upper, lower []float64, layerLower, layerUpper [][]float64, err error) {
	n := len(xs)
	if n == 0 {
		return nil, nil, nil, nil, errors.New("stacked area; x grid must not be empty")
	}
	for i, x := range xs {
		if !isFiniteFloat(x) {
			return nil, nil, nil, nil, fmt.Errorf("stacked area; x grid value at index %d is not finite", i)
		}
		if i > 0 && xs[i-1] >= x {
			return nil, nil, nil, nil, fmt.Errorf("stacked area; x grid must be strictly increasing at index %d", i)
		}
	}
	for li, row := range rows {
		if row == nil {
			continue
		}
		if len(row) != n {
			return nil, nil, nil, nil, fmt.Errorf("stacked area; layer %d has %d values for %d x grid points", li, len(row), n)
		}
		for i, v := range row {
			if !isFiniteFloat(v) {
				return nil, nil, nil, nil, fmt.Errorf("stacked area; layer %d value at index %d is not finite", li, i)
			}
		}
	}

	upper = make([]float64, n)
	lower = make([]float64, n)
	layerLower = make([][]float64, len(rows))
	layerUpper = make([][]float64, len(rows))
	for li := range rows {
		layerLower[li] = make([]float64, n)
		layerUpper[li] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		var posAcc, negAcc float64
		for li, row := range rows {
			var v float64
			if row != nil {
				v = row[i]
			}
			if v >= 0 {
				layerLower[li][i] = posAcc
				layerUpper[li][i] = posAcc + v
				posAcc += v
			} else {
				layerUpper[li][i] = negAcc
				layerLower[li][i] = negAcc + v
				negAcc += v
			}
		}
		upper[i] = posAcc
		lower[i] = negAcc
	}
	return upper, lower, layerLower, layerUpper, nil
}

// saPadRows normalizes rows onto an n-point grid. Nil rows become zero rows
// and short rows are zero padded, giving "missing item" a defined result.
// Non-finite entries are replaced with zero so rendering stays total; strict
// validation is available through Validate / ComputeStackedArea.
func saPadRows(n int, rows [][]float64) [][]float64 {
	out := make([][]float64, len(rows))
	for li, row := range rows {
		cp := make([]float64, n)
		for i := 0; i < n && i < len(row); i++ {
			if isFiniteFloat(row[i]) {
				cp[i] = row[i]
			}
		}
		out[li] = cp
	}
	return out
}

func isFiniteFloat(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// ComputeStackedArea accumulates a positive/negative separated stack for the
// shared increasing grid xs and one value row per component.
//
// Rows may be supplied either as a single [][]float64 matrix or as individual
// []float64 row arguments. It returns a descriptive error for an empty/
// non-increasing/non-finite grid, non-finite values or misaligned rows.
func ComputeStackedArea(xs []float64, rowsArg ...interface{}) (*StackedAreaResult, error) {
	rows := saCoerceRows(rowsArg)
	upper, lower, ll, lu, err := saAccumulate(xs, rows)
	if err != nil {
		return nil, err
	}
	return saBuildResult(xs, rows, upper, lower, ll, lu, nil), nil
}

// AccumulateStackedArea is a permissive entry point. Values may be supplied as
// a single [][]float64 matrix or as individual []float64 rows (variadic). It
// never returns an error for missing/short/non-finite entries: those are
// treated as zero contributions. A grid problem yields a nil result.
func AccumulateStackedArea(xs []float64, values ...interface{}) *StackedAreaResult {
	var rows [][]float64
	if len(values) == 1 {
		// Allow AccumulateStackedArea(xs, matrix) where matrix is [][]float64.
		if m, ok := values[0].([][]float64); ok {
			rows = m
		}
	}
	if rows == nil {
		rows = saCoerceRows(values)
	}
	if len(xs) == 0 {
		return nil
	}
	padded := saPadRows(len(xs), rows)
	upper, lower, ll, lu, err := saAccumulate(xs, padded)
	if err != nil {
		// Grid itself is unusable; fall back to a zero envelope so callers
		// never have to handle a panic.
		return nil
	}
	return saBuildResult(xs, padded, upper, lower, ll, lu, nil)
}

// StackedAreaValues is a signature-flexible compatibility helper. The shared
// X grid is the first argument; component rows may follow either as a single
// [][]float64 or as individual []float64 arguments.
func StackedAreaValues(xs []float64, values ...interface{}) (upper, lower []float64) {
	res := AccumulateStackedArea(xs, values...)
	if res == nil {
		return nil, nil
	}
	return res.Upper, res.Lower
}

// StackPositiveNegative accumulates positive and negative totals separately
// for an aligned component matrix and returns the global upper and lower
// cumulative lines.
func StackPositiveNegative(rowsArg ...interface{}) (upper, lower []float64) {
	rows := saCoerceRows(rowsArg)
	n := 0
	for _, r := range rows {
		if len(r) > n {
			n = len(r)
		}
	}
	xs := make([]float64, n)
	for i := range xs {
		xs[i] = float64(i)
	}
	padded := saPadRows(n, rows)
	u, l, _, _, err := saAccumulate(xs, padded)
	if err != nil {
		return nil, nil
	}
	return u, l
}

// StackValues is a compatibility alias for StackPositiveNegative.
func StackValues(rowsArg ...interface{}) (upper, lower []float64) {
	return StackPositiveNegative(rowsArg...)
}

// AccumulateStackedAreaByTime accepts time-major input: timeRows[t] is the
// set of component values at time/grid index t. It transposes to component
// rows before accumulating.
func AccumulateStackedAreaByTime(xs []float64, timeRows [][]float64) *StackedAreaResult {
	if len(timeRows) == 0 {
		return nil
	}
	components := 0
	for _, tr := range timeRows {
		if len(tr) > components {
			components = len(tr)
		}
	}
	rows := make([][]float64, components)
	for c := range rows {
		rows[c] = make([]float64, len(timeRows))
		for t := range timeRows {
			if c < len(timeRows[t]) {
				rows[c][t] = timeRows[t][c]
			}
		}
	}
	res, err := ComputeStackedArea(xs, rows)
	if err != nil {
		return AccumulateStackedArea(xs, rows)
	}
	return res
}

// StackByTime is the time-major variant of StackPositiveNegative.
func StackByTime(timeRows [][]float64) (upper, lower []float64) {
	n := len(timeRows)
	xs := make([]float64, n)
	for i := range xs {
		xs[i] = float64(i)
	}
	res := AccumulateStackedAreaByTime(xs, timeRows)
	if res == nil {
		return nil, nil
	}
	return res.Upper, res.Lower
}

// saTransposeTimeRows converts time-major rows (timeRows[t] = component values
// at time t) into component-major rows (rows[c][t]).
func saTransposeTimeRows(timeRows [][]float64) [][]float64 {
	if len(timeRows) == 0 {
		return nil
	}
	components := 0
	for _, tr := range timeRows {
		if len(tr) > components {
			components = len(tr)
		}
	}
	rows := make([][]float64, components)
	for c := range rows {
		rows[c] = make([]float64, len(timeRows))
		for t := range timeRows {
			if c < len(timeRows[t]) {
				rows[c][t] = timeRows[t][c]
			}
		}
	}
	return rows
}

func saBuildResult(xs []float64, rows [][]float64, upper, lower []float64, ll, lu [][]float64, names []string) *StackedAreaResult {
	res := &StackedAreaResult{
		X:           xs,
		Upper:       upper,
		Lower:       lower,
		Positive:    upper,
		Negative:    lower,
		Pos:         upper,
		Neg:         lower,
		Tops:        upper,
		Bottoms:     lower,
		High:        upper,
		Low:         lower,
		UpperBounds: upper,
		LowerBounds: lower,
		Min:         math.Inf(1),
		Max:         math.Inf(-1),
	}
	res.Layers = make([]StackedAreaLayerBounds, len(rows))
	res.Bands = make([][]StackedAreaBand, len(rows))
	for li := range rows {
		name := ""
		if li < len(names) {
			name = names[li]
		}
		res.Layers[li] = StackedAreaLayerBounds{Name: name, Lower: ll[li], Upper: lu[li]}
		res.Bands[li] = make([]StackedAreaBand, len(xs))
		for i := range xs {
			lo, hi := ll[li][i], lu[li][i]
			res.Bands[li][i] = StackedAreaBand{
				Index: i, X: xs[i], Lower: lo, Upper: hi,
				Bottom: lo, Top: hi, Min: lo, Max: hi,
			}
		}
	}
	for i := range xs {
		res.Min = math.Min(res.Min, lower[i])
		res.Max = math.Max(res.Max, upper[i])
	}
	if len(xs) == 0 {
		res.Min, res.Max = 0, 0
	}
	return res
}

// saCoerceRows turns heterogeneous variadic values into a row matrix.
func saCoerceRows(values []interface{}) [][]float64 {
	var rows [][]float64
	for _, v := range values {
		rows = append(rows, saCoerceValueToRows(v)...)
	}
	return rows
}

func saCoerceValueToRows(v interface{}) [][]float64 {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []float64:
		return [][]float64{t}
	case [][]float64:
		return t
	case []interface{}:
		rows := make([][]float64, 0, len(t))
		for _, e := range t {
			rows = append(rows, saCoerceValueToRows(e)...)
		}
		return rows
	case StackedAreaSeries:
		if r := saAnyToFloatSlice(t.YValues); r != nil {
			return [][]float64{r}
		}
		return nil
	case []StackedAreaSeries:
		rows := make([][]float64, 0, len(t))
		for _, s := range t {
			if r := saAnyToFloatSlice(s.YValues); r != nil {
				rows = append(rows, r)
			}
		}
		return rows
	case ContinuousSeries:
		return [][]float64{t.YValues}
	case []ContinuousSeries:
		rows := make([][]float64, 0, len(t))
		for _, s := range t {
			rows = append(rows, s.YValues)
		}
		return rows
	}
	// Reflection fallback for named slice types.
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if rv.Type().Elem().Kind() == reflect.Float64 {
			out := make([]float64, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				out[i] = rv.Index(i).Float()
			}
			return [][]float64{out}
		}
		if rv.Type().Elem().Kind() == reflect.Slice {
			rows := make([][]float64, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				rows = append(rows, saCoerceValueToRows(rv.Index(i).Interface())...)
			}
			return rows
		}
	}
	return nil
}

// saAnyToFloatSlice accepts []float64, named float slices and []interface{}.
func saAnyToFloatSlice(v interface{}) []float64 {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []float64:
		return t
	case []interface{}:
		out := make([]float64, 0, len(t))
		for _, e := range t {
			switch n := e.(type) {
			case float64:
				out = append(out, n)
			case int:
				out = append(out, float64(n))
			case int64:
				out = append(out, float64(n))
			}
		}
		return out
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		if rv.Type().Elem().Kind() == reflect.Float64 {
			out := make([]float64, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				out[i] = rv.Index(i).Float()
			}
			return out
		}
	}
	return nil
}
