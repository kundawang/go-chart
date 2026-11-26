package chart

// This file collects spelling aliases and forgiving constructors for
// SymmetricLogRange. The canonical API is SymmetricLogRange with the
// Threshold (and optional Base) fields; the aliases exist so callers used to
// other "symlog / signed-log" naming conventions plug in unchanged.

// SymmetricLog is an alias for SymmetricLogRange.
type SymmetricLog = SymmetricLogRange

// SymLogRange is an alias for SymmetricLogRange.
type SymLogRange = SymmetricLogRange

// SymLog is an alias for SymmetricLogRange.
type SymLog = SymmetricLogRange

// SymmetricalLogRange is an alias for SymmetricLogRange (matplotlib wording).
type SymmetricalLogRange = SymmetricLogRange

// SymmetricalLog is an alias for SymmetricLogRange.
type SymmetricalLog = SymmetricLogRange

// SignedLogRange is an alias for SymmetricLogRange ("signed log" wording).
type SignedLogRange = SymmetricLogRange

// SignedLog is an alias for SymmetricLogRange.
type SignedLog = SymmetricLogRange

// BiSymmetricLogRange is an alias for SymmetricLogRange.
type BiSymmetricLogRange = SymmetricLogRange

// BiSymLogRange is an alias for SymmetricLogRange.
type BiSymLogRange = SymmetricLogRange

// SymmetricLogScale, SymLogScale and SymmetricalLogScale alias the range for
// callers that name this concept a "scale".
type SymmetricLogScale = SymmetricLogRange

// SymLogScale is an alias for SymmetricLogRange.
type SymLogScale = SymmetricLogRange

// SymmetricalLogScale is an alias for SymmetricLogRange.
type SymmetricalLogScale = SymmetricLogRange

// SignedLogScale is an alias for SymmetricLogRange.
type SignedLogScale = SymmetricLogRange

// NewSymmetricLogRange builds a *SymmetricLogRange from a variety of common
// argument shapes:
//
//	NewSymmetricLogRange()                                  // defaults
//	NewSymmetricLogRange(threshold)                         // threshold only
//	NewSymmetricLogRange(min, max)                          // bounds
//	NewSymmetricLogRange(min, max, threshold)               // bounds + threshold
//	NewSymmetricLogRange(min, max, threshold, domain)        // full
//
// Arguments may be any numeric type. When four numeric arguments are given
// the implementation also recognizes the (min, max, domain, threshold)
// ordering when it is unambiguous.
func NewSymmetricLogRange(args ...interface{}) *SymmetricLogRange {
	r := &SymmetricLogRange{}
	applyRangeArgs(r, args)
	return r
}

// NewSymmetricLog is an alias constructor for NewSymmetricLogRange.
func NewSymmetricLog(args ...interface{}) *SymmetricLogRange {
	return NewSymmetricLogRange(args...)
}

// NewSymLogRange is an alias constructor for NewSymmetricLogRange.
func NewSymLogRange(args ...interface{}) *SymmetricLogRange {
	return NewSymmetricLogRange(args...)
}

// NewSymLog is an alias constructor for NewSymmetricLogRange.
func NewSymLog(args ...interface{}) *SymmetricLogRange {
	return NewSymmetricLogRange(args...)
}

// NewSymmetricalLogRange is an alias constructor for NewSymmetricLogRange.
func NewSymmetricalLogRange(args ...interface{}) *SymmetricLogRange {
	return NewSymmetricLogRange(args...)
}

// NewSignedLogRange is an alias constructor for NewSymmetricLogRange.
func NewSignedLogRange(args ...interface{}) *SymmetricLogRange {
	return NewSymmetricLogRange(args...)
}

func applyRangeArgs(r *SymmetricLogRange, args []interface{}) {
	nums := make([]float64, 0, len(args))
	intLike := make([]bool, 0, len(args))
	for _, arg := range args {
		switch v := arg.(type) {
		case int:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case int8:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case int16:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case int32:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case int64:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case uint:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case uint8:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case uint16:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case uint32:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case uint64:
			nums = append(nums, float64(v))
			intLike = append(intLike, true)
		case float32:
			nums = append(nums, float64(v))
			intLike = append(intLike, false)
		case float64:
			nums = append(nums, v)
			intLike = append(intLike, false)
		}
	}

	switch len(nums) {
	case 1:
		r.Threshold = nums[0]
	case 2:
		r.Min, r.Max = nums[0], nums[1]
	case 3:
		r.Min, r.Max, r.Threshold = nums[0], nums[1], nums[2]
	case 4:
		r.Min, r.Max = nums[0], nums[1]
		a, b := nums[2], nums[3]
		// Prefer the documented (min, max, threshold, domain) order, but
		// accept (min, max, domain, threshold) when concrete int types or
		// relative magnitudes make it obvious.
		thirdIsDomain := intLike[2] && !intLike[3]
		if !thirdIsDomain && a >= 2 && b > 0 && b < a && a == float64(int(a)) {
			thirdIsDomain = true
		}
		if thirdIsDomain {
			r.Domain = int(a)
			r.Threshold = b
		} else {
			r.Threshold = a
			r.Domain = int(b)
		}
	default:
		if len(nums) >= 5 {
			r.Min, r.Max = nums[0], nums[1]
			r.Threshold = nums[2]
			r.Domain = int(nums[3])
			r.Base = nums[4]
		}
	}
}

// SymmetricLogValue applies the canonical symmetric-log transform to v with
// the given threshold (defaulting to 1) and optional base (defaulting to 10).
// It is the unnormalized phi used internally and is provided as a standalone
// helper for annotations or custom renderables.
func SymmetricLogValue(v float64, options ...float64) float64 {
	t := DefaultSymmetricLogThreshold
	b := DefaultSymmetricLogBase
	if len(options) >= 1 && options[0] > 0 {
		t = options[0]
	}
	if len(options) >= 2 && options[1] > 0 && options[1] != 1 {
		b = options[1]
	}
	r := SymmetricLogRange{Threshold: t, Base: b}
	return r.mapValue(v)
}
