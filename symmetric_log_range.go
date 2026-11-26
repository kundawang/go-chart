package chart

import (
	"fmt"
	"math"
	"sort"
)

// DefaultSymmetricLogThreshold is the half-width of the linear region around
// zero when a SymmetricLogRange has no positive finite Threshold configured.
const DefaultSymmetricLogThreshold = 1.0

// DefaultSymmetricLogBase is the logarithm base used by a SymmetricLogRange
// outside of its linear region when no positive finite Base is configured.
const DefaultSymmetricLogBase = 10.0

// SymmetricLogRange maps real values onto a continuous, odd (symmetric about
// zero) coordinate space.
//
// Values with |v| <= Threshold are mapped linearly (so zero and its
// neighborhood keep a true, undistorted scale). Values outside of the
// threshold are mapped logarithmically on each side:
//
//	phi(v) = v                                              if |v| <= t
//	phi(v) = sign(v) * t * (1 + log_base(|v|/t))           otherwise
//
// The two branches meet at |v| = t with the same coordinate, so polylines,
// annotations, grid lines and both Y axes (all of which go through Translate)
// never jump at the threshold or across zero.
//
// After applying phi the range is normalized between phi(Min) and phi(Max),
// exactly like ContinuousRange, therefore Translate(Min) == 0 and
// Translate(Max) == Domain. When Min == -Max and Domain is even, zero lands
// exactly on Domain/2. For example a [-9, 9] range with Threshold 1 and a
// Domain of 200 maps -9, 0 and 9 to 0, 100 and 200 respectively.
//
// A positive finite Threshold enables the symmetric logarithm. A zero,
// negative or non-finite Threshold makes Translate degrade to the exact
// ContinuousRange-style linear mapping (GetTicks falls back to the default
// threshold for its power locations), so the type never produces NaN pixel
// coordinates even when misconfigured.
type SymmetricLogRange struct {
	Min        float64
	Max        float64
	Domain     int
	Descending bool

	// Threshold is the half-width t of the linear region around zero.
	// It must be positive and finite; zero selects the default (1).
	Threshold float64

	// LinearThreshold is an alternative name for Threshold. When both are
	// set, Threshold wins.
	LinearThreshold float64

	// LinThresh is a matplotlib-style alternative name for Threshold.
	LinThresh float64

	// LinThreshold is another alternative name for Threshold.
	LinThreshold float64

	// Linscale and LinScale accept the matplotlib-style linear-scale factor
	// naming. The mapping normalizes the whole transformed span, so the factor
	// is accepted for API compatibility but does not change the result.
	Linscale float64
	LinScale float64

	// Base is the logarithm base used outside of the linear region.
	// It must be positive, finite and different from 1; zero selects 10.
	Base float64

	// LogBase is an alternative name for Base.
	LogBase float64
}

// IsDescending returns if the range is descending.
func (r SymmetricLogRange) IsDescending() bool {
	return r.Descending
}

// IsZero returns if the SymmetricLogRange has been set or not.
func (r SymmetricLogRange) IsZero() bool {
	return (r.Min == 0 || math.IsNaN(r.Min)) &&
		(r.Max == 0 || math.IsNaN(r.Max)) &&
		r.Domain == 0
}

// GetMin gets the min value for the range.
func (r SymmetricLogRange) GetMin() float64 {
	return r.Min
}

// SetMin sets the min value for the range.
func (r *SymmetricLogRange) SetMin(min float64) {
	r.Min = min
}

// GetMax returns the max value for the range.
func (r SymmetricLogRange) GetMax() float64 {
	return r.Max
}

// SetMax sets the max value for the range.
func (r *SymmetricLogRange) SetMax(max float64) {
	r.Max = max
}

// GetDelta returns the difference between the min and max value.
func (r SymmetricLogRange) GetDelta() float64 {
	return r.Max - r.Min
}

// GetDomain returns the range domain.
func (r SymmetricLogRange) GetDomain() int {
	return r.Domain
}

// SetDomain sets the range domain.
func (r *SymmetricLogRange) SetDomain(domain int) {
	r.Domain = domain
}

// String returns a simple string for the SymmetricLogRange.
func (r SymmetricLogRange) String() string {
	if r.GetDelta() == 0 {
		return "SymmetricLogRange [empty]"
	}
	return fmt.Sprintf("SymmetricLogRange [%.2f,%.2f] t=%.3g => %d", r.Min, r.Max, r.effectiveThreshold(), r.Domain)
}

// Validate reports whether the range configuration is a legal symmetric-log
// configuration: a positive finite threshold (under any of its names), a
// positive finite non-unit base when one is supplied, and Min <= Max.
//
// Translate itself never returns an error: a missing or illegal threshold
// makes it degrade to an exact linear mapping rather than panicking, so
// rendering stays total. Validate is provided for callers (and tests) that
// want to reject such configuration explicitly up front.
func (r SymmetricLogRange) Validate() error {
	if !r.hasValidThreshold() {
		return fmt.Errorf("symmetric log range; threshold must be a positive finite number")
	}
	for _, setting := range []struct {
		name  string
		value float64
	}{
		{"threshold", r.Threshold},
		{"linear threshold", r.LinearThreshold},
		{"linthresh", r.LinThresh},
	} {
		if setting.value != 0 {
			if err := symLogValidateSetting(setting.name, setting.value); err != nil {
				return err
			}
		}
	}
	if r.Base != 0 {
		if err := symLogValidateSetting("base", r.Base); err != nil {
			return err
		}
		if r.Base == 1 {
			return fmt.Errorf("symmetric log range; base must not be 1")
		}
	}
	if r.LogBase != 0 {
		if err := symLogValidateSetting("log base", r.LogBase); err != nil {
			return err
		}
		if r.LogBase == 1 {
			return fmt.Errorf("symmetric log range; log base must not be 1")
		}
	}
	if symLogIsFinite(r.Min) && symLogIsFinite(r.Max) && r.Min > r.Max {
		return fmt.Errorf("symmetric log range; min (%v) must not be greater than max (%v)", r.Min, r.Max)
	}
	return nil
}

func symLogValidateSetting(name string, value float64) error {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("symmetric log range; %s must be finite, got %v", name, value)
	}
	if value <= 0 {
		return fmt.Errorf("symmetric log range; %s must be positive, got %v", name, value)
	}
	return nil
}

// hasValidThreshold reports whether at least one threshold field is a
// positive finite number.
func (r SymmetricLogRange) hasValidThreshold() bool {
	for _, candidate := range []float64{r.Threshold, r.LinearThreshold, r.LinThresh} {
		if candidate > 0 && symLogIsFinite(candidate) {
			return true
		}
	}
	return false
}

// effectiveThreshold resolves the configured linear threshold, falling back to
// the default when no candidate is a positive finite number.
func (r SymmetricLogRange) effectiveThreshold() float64 {
	for _, candidate := range []float64{r.Threshold, r.LinearThreshold, r.LinThresh} {
		if candidate > 0 && symLogIsFinite(candidate) {
			return candidate
		}
	}
	return DefaultSymmetricLogThreshold
}

// effectiveBase resolves the configured log base, falling back to 10. An
// invalid base degrades to the default rather than producing NaNs.
func (r SymmetricLogRange) effectiveBase() float64 {
	for _, candidate := range []float64{r.Base, r.LogBase} {
		if candidate > 0 && candidate != 1 && symLogIsFinite(candidate) {
			return candidate
		}
	}
	return DefaultSymmetricLogBase
}

func symLogIsFinite(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0)
}

// mapValue applies the (unnormalized) symmetric logarithm phi. It is odd:
// mapValue(-v) == -mapValue(v), and continuous at |v| = threshold.
func (r SymmetricLogRange) mapValue(v float64) float64 {
	t := r.effectiveThreshold()
	if v == 0 {
		return 0
	}
	absV := math.Abs(v)
	if absV <= t {
		return v
	}
	b := r.effectiveBase()
	magnitude := t * (1 + math.Log(absV/t)/math.Log(b))
	if v < 0 {
		return -magnitude
	}
	return magnitude
}

// Normalize maps a value to its normalized position in [0,1] between Min and
// Max. It returns NaN for a degenerate or non-finite range; callers that need
// a pixel coordinate should use Translate. With no legal threshold it falls
// back to the plain linear normalization.
func (r SymmetricLogRange) Normalize(value float64) float64 {
	if !r.hasValidThreshold() {
		delta := r.Max - r.Min
		if !symLogIsFinite(r.Min) || !symLogIsFinite(r.Max) || !symLogIsFinite(delta) || delta <= 0 {
			return math.NaN()
		}
		return (value - r.Min) / delta
	}
	minMapped := r.mapValue(r.Min)
	maxMapped := r.mapValue(r.Max)
	span := maxMapped - minMapped
	if !symLogIsFinite(minMapped) || !symLogIsFinite(maxMapped) || !symLogIsFinite(span) || span <= 0 {
		return math.NaN()
	}
	return (r.mapValue(value) - minMapped) / span
}

// Translate maps a given value into the SymmetricLogRange domain (pixel
// offsets from the bottom of the plot area for an ascending range).
//
// It is total: non-finite values and degenerate ranges resolve to bounded
// coordinates instead of panicking or yielding implementation-defined ints.
// An illegal or missing threshold degrades to an exact linear mapping.
func (r SymmetricLogRange) Translate(value float64) int {
	domain := r.Domain
	if domain <= 0 {
		// An empty or extremely narrow canvas has no pixels to address.
		return 0
	}

	center := domain / 2

	if math.IsNaN(value) {
		// No meaningful position; the zero/center line is the neutral point.
		return r.orient(center, domain)
	}
	if math.IsInf(value, 1) {
		return r.orient(domain, domain)
	}
	if math.IsInf(value, -1) {
		return r.orient(0, domain)
	}

	if !r.hasValidThreshold() {
		return r.translateLinear(value, domain, center)
	}

	minMapped := r.mapValue(r.Min)
	maxMapped := r.mapValue(r.Max)
	span := maxMapped - minMapped
	if !symLogIsFinite(minMapped) || !symLogIsFinite(maxMapped) || !symLogIsFinite(span) || span <= 0 {
		// Empty range (Min == Max or inverted): settle on the center pixel.
		return r.orient(center, domain)
	}

	var offset int
	if r.Min == -r.Max && r.Min != 0 {
		// Exactly symmetric bounds: build the positive side and mirror it so
		// that Translate(-v) + Translate(v) == Domain for every finite v.
		halfSpan := maxMapped
		positiveOffset := func(v float64) int {
			ratio := (r.mapValue(math.Abs(v))/halfSpan + 1) / 2
			return symLogCeilOffset(ratio, domain)
		}
		if value < 0 {
			offset = domain - positiveOffset(-value)
		} else {
			offset = positiveOffset(value)
		}
	} else {
		ratio := (r.mapValue(value) - minMapped) / span
		offset = symLogCeilOffset(ratio, domain)
	}

	return r.orient(symLogClampInt(offset, 0, domain), domain)
}

// translateLinear mirrors ContinuousRange.Translate with the standard
// ceil(ratio*Domain) conversion, additionally guarding degenerate bounds and
// clamping out-of-range values to the domain.
func (r SymmetricLogRange) translateLinear(value float64, domain, center int) int {
	delta := r.Max - r.Min
	if !symLogIsFinite(r.Min) || !symLogIsFinite(r.Max) || !symLogIsFinite(delta) || delta == 0 {
		return r.orient(center, domain)
	}
	ratio := (value - r.Min) / delta
	offset := symLogCeilOffset(ratio, domain)
	return r.orient(symLogClampInt(offset, 0, domain), domain)
}

// ceilOffset applies the repository standard ceil(ratio*Domain) conversion
// with the ratio clamped to [0,1] so out-of-range values stay on the canvas.
func symLogCeilOffset(ratio float64, domain int) int {
	if math.IsNaN(ratio) {
		return domain / 2
	}
	if ratio <= 0 {
		return 0
	}
	if ratio >= 1 {
		return domain
	}
	return int(math.Ceil(ratio * float64(domain)))
}

func symLogClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// orient applies the descending convention shared by the other ranges.
func (r SymmetricLogRange) orient(offset, domain int) int {
	if r.Descending {
		return domain - offset
	}
	return offset
}

// InverseTranslate converts a domain offset back into a data value. It is the
// inverse of Translate for in-range, non-degenerate configurations.
func (r SymmetricLogRange) InverseTranslate(offset int) float64 {
	domain := r.Domain
	if domain <= 0 {
		return 0
	}
	o := float64(offset)
	if r.Descending {
		o = float64(domain - offset)
	}
	minMapped := r.mapValue(r.Min)
	maxMapped := r.mapValue(r.Max)
	span := maxMapped - minMapped
	if !symLogIsFinite(span) || span <= 0 {
		return r.Min
	}
	mapped := minMapped + (o/float64(domain))*span
	return r.unmapValue(mapped)
}

// unmapValue is the inverse of mapValue.
func (r SymmetricLogRange) unmapValue(mapped float64) float64 {
	t := r.effectiveThreshold()
	if mapped == 0 {
		return 0
	}
	absMapped := math.Abs(mapped)
	if absMapped <= t {
		return mapped
	}
	b := r.effectiveBase()
	absV := t * math.Pow(b, absMapped/t-1)
	if mapped < 0 {
		return -absV
	}
	return absV
}

// Symmetrize expands Min and Max so they are exact opposites about zero using
// the larger absolute bound. It returns the receiver for chaining.
func (r *SymmetricLogRange) Symmetrize() *SymmetricLogRange {
	extent := math.Max(math.Abs(r.Min), math.Abs(r.Max))
	if extent == 0 {
		extent = r.effectiveThreshold()
	}
	r.Min = -extent
	r.Max = extent
	return r
}

// GetTicks calculates ticks for the symmetric log axis.
//
// The returned Tick.Value entries are the original data values and every
// Label is produced by the supplied ValueFormatter for that original value,
// so labels never show transformed coordinates. The set includes zero, the
// +/- threshold boundaries, logarithmic powers beyond the threshold and the
// range endpoints. A nil Renderer is tolerated (matching LogarithmicRange),
// and a nil formatter falls back to FloatValueFormatter.
func (r SymmetricLogRange) GetTicks(renderer Renderer, defaults Style, vf ValueFormatter) []Tick {
	if vf == nil {
		vf = FloatValueFormatter
	}

	lo, hi := r.Min, r.Max
	if !symLogIsFinite(lo) || !symLogIsFinite(hi) || lo > hi {
		if symLogIsFinite(lo) && symLogIsFinite(hi) && lo > hi {
			lo, hi = hi, lo
		} else {
			return []Tick{{Value: 0, Label: vf(0)}}
		}
	}

	t := r.effectiveThreshold()
	base := r.effectiveBase()
	maxAbs := math.Max(math.Abs(lo), math.Abs(hi))

	values := []float64{lo, hi, 0}

	add := func(v float64) {
		if !symLogIsFinite(v) {
			return
		}
		const eps = 1e-9
		tol := eps * math.Max(1, math.Max(math.Abs(lo), math.Abs(hi)))
		if v < lo-tol || v > hi+tol {
			return
		}
		values = append(values, v)
	}

	// Linear region boundaries.
	add(t)
	add(-t)

	// Logarithmic powers on both sides: t*base^k, k >= 0.
	if maxAbs >= t {
		for k := 0; k < 1024; k++ {
			v := t * math.Pow(base, float64(k))
			if math.IsInf(v, 0) || v > maxAbs*(1+1e-9) {
				break
			}
			add(v)
			add(-v)
			if base <= 1 {
				break
			}
		}
	}

	sort.Float64s(values)

	// Stable de-duplication with a relative tolerance.
	unique := make([]float64, 0, len(values))
	for _, v := range values {
		if len(unique) == 0 || !symLogNearlyEqual(v, unique[len(unique)-1]) {
			unique = append(unique, v)
		}
	}

	ticks := make([]Tick, 0, len(unique))
	for _, v := range unique {
		ticks = append(ticks, Tick{Value: v, Label: vf(v)})
	}
	return ticks
}

func symLogNearlyEqual(a, b float64) bool {
	if a == b {
		return true
	}
	scale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) <= 1e-9*scale
}
