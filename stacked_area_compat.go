package chart

// This file contains additional exported name spellings for the stacked area
// feature. The canonical API lives in stacked_area.go and
// stacked_area_series.go; every declaration here is a thin alias/wrapper.

// CalculateStackedArea is an alias for ComputeStackedArea.
func CalculateStackedArea(xs []float64, rows [][]float64) (*StackedAreaResult, error) {
	return ComputeStackedArea(xs, rows)
}

// CalcStackedArea is an alias for ComputeStackedArea.
func CalcStackedArea(xs []float64, rows [][]float64) (*StackedAreaResult, error) {
	return ComputeStackedArea(xs, rows)
}

// BuildStackedArea is an alias for ComputeStackedArea.
func BuildStackedArea(xs []float64, rows [][]float64) (*StackedAreaResult, error) {
	return ComputeStackedArea(xs, rows)
}

// NewStackedArea builds a shared-grid series from variadic component rows.
func NewStackedArea(xs []float64, rows ...[]float64) *StackedAreaSeries {
	return &StackedAreaSeries{XValues: xs, YValues: rows}
}

// ComputeStacks strictly accumulates the matrix and returns envelope points.
func ComputeStacks(xs []float64, rows [][]float64) ([]StackedAreaPoint, error) {
	res, err := ComputeStackedArea(xs, rows)
	if err != nil {
		return nil, err
	}
	return res.Points(), nil
}

// PositiveNegativeStack accumulates positives and negatives separately and
// returns the global (upper, lower) cumulative lines.
func PositiveNegativeStack(rows ...interface{}) (upper, lower []float64) {
	return StackPositiveNegative(rows...)
}

// DivergingStack is an alias for PositiveNegativeStack.
func DivergingStack(rows ...interface{}) (upper, lower []float64) {
	return StackPositiveNegative(rows...)
}

// StackAreaChart is a naming alias for StackedAreaChart.
type StackAreaChart = StackedAreaChart

// StackArea is a naming alias for the primary series type.
type StackArea = StackedAreaSeries

// StackedAreaLayer is a naming alias for a component layer.
type StackedAreaLayer = StackedAreaSeries

// DivergingAreaSeries is a naming alias for StackedAreaSeries.
type DivergingAreaSeries = StackedAreaSeries

// PositiveNegativeAreaSeries is a naming alias for StackedAreaSeries.
type PositiveNegativeAreaSeries = StackedAreaSeries

// Stack is an alias for the envelope point type.
type Stack = StackedAreaPoint

// StackedAreas is an alias for a slice of envelope points.
type StackedAreas = []StackedAreaPoint

// The following are thin method aliases for common computed-accessor names.

// Stack is an alias for Stacks.
func (sas StackedAreaSeries) Stack() []StackedAreaPoint { return sas.Stacks() }

// Cumulative is an alias for Stacks.
func (sas StackedAreaSeries) Cumulative() []StackedAreaPoint { return sas.Stacks() }

// Accumulate is an alias for Stacks.
func (sas StackedAreaSeries) Accumulate() []StackedAreaPoint { return sas.Stacks() }

// Calculate strictly computes the result.
func (sas StackedAreaSeries) Calculate() (*StackedAreaResult, error) { return sas.Compute() }

// Calc is an alias for Calculate.
func (sas StackedAreaSeries) Calc() (*StackedAreaResult, error) { return sas.Compute() }

// Build is an alias for Calculate.
func (sas StackedAreaSeries) Build() (*StackedAreaResult, error) { return sas.Compute() }

// Result returns the permissive computed result (never an error).
func (sas StackedAreaSeries) Result() *StackedAreaResult { return sas.strictOrPadded() }

// GetResult is an alias for Result.
func (sas StackedAreaSeries) GetResult() *StackedAreaResult { return sas.strictOrPadded() }

// Upper returns the positive cumulative upper line.
func (sas StackedAreaSeries) Upper() []float64 { return sas.UpperBounds() }

// Lower returns the negative cumulative lower line.
func (sas StackedAreaSeries) Lower() []float64 { return sas.LowerBounds() }

// Tops is an alias for Upper.
func (sas StackedAreaSeries) Tops() []float64 { return sas.UpperBounds() }

// Bottoms is an alias for Lower.
func (sas StackedAreaSeries) Bottoms() []float64 { return sas.LowerBounds() }

// GetUpperBounds is an alias for UpperBounds.
func (sas StackedAreaSeries) GetUpperBounds() []float64 { return sas.UpperBounds() }

// GetLowerBounds is an alias for LowerBounds.
func (sas StackedAreaSeries) GetLowerBounds() []float64 { return sas.LowerBounds() }

// GetStackedBounds is an alias for StackedValues.
func (sas StackedAreaSeries) GetStackedBounds() [][][2]float64 { return sas.StackedValues() }

// StackValuesMatrix is an alias for StackedValues.
func (sas StackedAreaSeries) StackValuesMatrix() [][][2]float64 { return sas.StackedValues() }

// DivergingStackedAreaChart is a naming alias for StackedAreaChart.
type DivergingStackedAreaChart = StackedAreaChart

// PositiveNegativeStackedAreaChart is a naming alias for StackedAreaChart.
type PositiveNegativeStackedAreaChart = StackedAreaChart
