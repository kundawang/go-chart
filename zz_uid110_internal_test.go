package chart

import (
	"testing"
)

func TestUID110InternalRangeIncludesCumulativeExtrema(t *testing.T) {
	c := Chart{
		Series: []Series{
			StackedAreaSeries{
				XValues: []float64{0, 1},
				Series: []StackedAreaSeries{
					{Name: "render", YValues: []float64{2, 1}},
					{Name: "script", YValues: []float64{-1, 3}},
				},
			},
		},
	}
	xr, yr, yra := c.getRanges()
	_ = yra

	if xr.GetMin() > 0 || xr.GetMax() < 1 {
		t.Fatalf("x range should span grid, got [%v,%v]", xr.GetMin(), xr.GetMax())
	}
	// Raw auto range must cover cumulative lower -1 and upper 4. Rounding may
	// widen it but must never shrink inside the extrema.
	if yr.GetMin() > -1 {
		t.Fatalf("y min %v must be <= cumulative lower -1", yr.GetMin())
	}
	if yr.GetMax() < 4 {
		t.Fatalf("y max %v must be >= cumulative upper 4", yr.GetMax())
	}
	if err := c.checkRanges(xr, yr, yra); err != nil {
		t.Fatalf("checkRanges: %v", err)
	}
}

func TestUID110InternalMatrixOrientation(t *testing.T) {
	s := StackedAreaSeries{
		XValues: []float64{0, 1},
		YValues: [][]float64{{2, 1}, {-1, 3}},
	}
	pts := s.Stacks()
	if pts[0].Upper != 2 || pts[0].Lower != -1 || pts[1].Upper != 4 {
		t.Fatalf("bad oracle stacks: %+v %+v", pts[0], pts[1])
	}
}
