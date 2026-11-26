package chart_test

import (
	"bytes"
	"image/png"
	"math"
	"strings"
	"testing"

	chart "github.com/wcharczuk/go-chart/v2"
)

// uid110OracleSeries returns the fixed two-timepoint fixture:
//
//	t0: render +2, script -1   -> upper 2, lower -1
//	t1: render +1, script +3   -> upper 4, lower 0
func uid110OracleRows() (xs []float64, rows [][]float64) {
	return []float64{0, 1}, [][]float64{{2, 1}, {-1, 3}}
}

func uid110CheckOraclePoints(t *testing.T, pts []chart.StackedAreaPoint) {
	t.Helper()
	if len(pts) != 2 {
		t.Fatalf("expected 2 points, got %d", len(pts))
	}
	if pts[0].Upper != 2 || pts[0].Top != 2 || pts[0].Max != 2 || pts[0].Positive != 2 {
		t.Fatalf("t0 upper boundary want 2, got %+v", pts[0])
	}
	if pts[0].Lower != -1 || pts[0].Bottom != -1 || pts[0].Min != -1 || pts[0].Negative != -1 {
		t.Fatalf("t0 lower boundary want -1, got %+v", pts[0])
	}
	if pts[1].Upper != 4 {
		t.Fatalf("t1 positive upper boundary want 4, got %v", pts[1].Upper)
	}
	if pts[1].Lower != 0 {
		t.Fatalf("t1 lower boundary want 0, got %v", pts[1].Lower)
	}
}

func TestUID110CoreAccumulationOracle(t *testing.T) {
	xs, rows := uid110OracleRows()

	res, err := chart.ComputeStackedArea(xs, rows)
	if err != nil {
		t.Fatalf("ComputeStackedArea: %v", err)
	}
	uid110CheckOraclePoints(t, res.Points())

	if res.Upper[0] != 2 || res.Lower[0] != -1 || res.Upper[1] != 4 {
		t.Fatalf("unexpected envelope upper=%v lower=%v", res.Upper, res.Lower)
	}
	if res.Max != 4 || res.Min != -1 {
		t.Fatalf("cumulative extrema want min -1 max 4, got min %v max %v", res.Min, res.Max)
	}
}

func TestUID110PackageHelpersOracle(t *testing.T) {
	xs, rows := uid110OracleRows()

	u, l := chart.StackPositiveNegative(rows)
	if u[0] != 2 || l[0] != -1 || u[1] != 4 {
		t.Fatalf("StackPositiveNegative got u=%v l=%v", u, l)
	}

	u2, l2 := chart.StackValues(rows)
	if u2[1] != 4 || l2[0] != -1 {
		t.Fatalf("StackValues got u=%v l=%v", u2, l2)
	}

	res := chart.AccumulateStackedArea(xs, rows[0], rows[1])
	if res == nil || res.Upper[1] != 4 || res.Lower[0] != -1 {
		t.Fatalf("variadic AccumulateStackedArea failed: %+v", res)
	}

	res2 := chart.AccumulateStackedArea(xs, rows)
	if res2 == nil || res2.Upper[0] != 2 {
		t.Fatalf("matrix AccumulateStackedArea failed: %+v", res2)
	}

	u3, l3 := chart.StackedAreaValues(xs, rows)
	if u3[1] != 4 || l3[0] != -1 {
		t.Fatalf("StackedAreaValues got u=%v l=%v", u3, l3)
	}
}

func TestUID110SeriesNestedOracle(t *testing.T) {
	sas := chart.StackedAreaSeries{
		XValues: []float64{0, 1},
		Series: []chart.StackedAreaSeries{
			{Name: "render", YValues: []float64{2, 1}},
			{Name: "script", YValues: []float64{-1, 3}},
		},
	}

	if err := sas.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	uid110CheckOraclePoints(t, sas.Stacks())
	uid110CheckOraclePoints(t, sas.Bounds())
	uid110CheckOraclePoints(t, sas.StackedBounds())

	x, lo, hi := sas.GetBoundedValues(0)
	if x != 0 || lo != -1 || hi != 2 {
		t.Fatalf("bounded[0] want (0,-1,2), got (%v,%v,%v)", x, lo, hi)
	}
	if sas.Len() != 2 {
		t.Fatalf("Len want 2 got %d", sas.Len())
	}
	_, lo1, hi1 := sas.GetBoundedValues(1)
	if lo1 != 0 || hi1 != 4 {
		t.Fatalf("bounded[1] want (_,0,4), got (%v,%v)", lo1, hi1)
	}

	entries := sas.LegendEntries(chart.Style{})
	if len(entries) != 2 || entries[0].Name != "render" || entries[1].Name != "script" {
		t.Fatalf("legend entries mismatch: %+v", entries)
	}
}

func TestUID110SeriesMatrixOracle(t *testing.T) {
	sas := chart.StackedAreaSeries{
		XValues: []float64{0, 1},
		YValues: [][]float64{{2, 1}, {-1, 3}},
		Names:   []string{"render", "script"},
	}
	if err := sas.Validate(); err != nil {
		t.Fatalf("Validate matrix: %v", err)
	}
	uid110CheckOraclePoints(t, sas.Stacks())
	if sas.LayerBounds()[1].Upper[1] != 4 {
		t.Fatalf("second layer upper at t1 want 4 (1+3), got %v", sas.LayerBounds()[1].Upper[1])
	}
}

func TestUID110NegativeDoesNotCancelPositive(t *testing.T) {
	// Same instant +2 and -1 must yield top 2 and bottom -1, not net 1.
	sas := chart.StackedAreaSeries{
		XValues: []float64{0},
		Series: []chart.StackedAreaSeries{
			{YValues: []float64{2}},
			{YValues: []float64{-1}},
		},
	}
	pts := sas.Stacks()
	if len(pts) != 1 || pts[0].Upper != 2 || pts[0].Lower != -1 {
		t.Fatalf("separation violated: %+v", pts)
	}
}

func TestUID110SinglePointCore(t *testing.T) {
	res, err := chart.ComputeStackedArea([]float64{7}, [][]float64{{2}, {-1}})
	if err != nil {
		t.Fatalf("single point compute: %v", err)
	}
	if len(res.Points()) != 1 || res.Upper[0] != 2 || res.Lower[0] != -1 {
		t.Fatalf("single point bad result: %+v", res.Points())
	}
}

func TestUID110ValidateEdgeCases(t *testing.T) {
	bad := []chart.StackedAreaSeries{
		{}, // no layers
		{XValues: []float64{0, 1}, YValues: []float64{1}},                  // length mismatch
		{XValues: []float64{0, 0}, YValues: [][]float64{{1, 2}}},           // not increasing
		{XValues: []float64{0, math.NaN()}, YValues: [][]float64{{1, 2}}},  // non-finite x
		{XValues: []float64{0, 1}, YValues: [][]float64{{1, math.Inf(1)}}}, // non-finite y
	}
	for i, s := range bad {
		if err := s.Validate(); err == nil {
			t.Fatalf("case %d expected validation error", i)
		}
	}

	// Mismatched per-layer X grids are a clear error.
	mismatch := chart.StackedAreaSeries{
		Series: []chart.StackedAreaSeries{
			{XValues: []float64{0, 1}, YValues: []float64{1, 2}},
			{XValues: []float64{0, 2}, YValues: []float64{1, 2}},
		},
	}
	if err := mismatch.Validate(); err == nil {
		t.Fatal("expected x grid mismatch error")
	}
}

func TestUID110MissingItemsAreDefined(t *testing.T) {
	// Nil/short rows are zero padded by the permissive path. Rows align from
	// index zero, so the short {1} contributes at index 0 and its tail is 0.
	sas := chart.StackedAreaSeries{
		XValues: []float64{0, 1, 2},
		YValues: [][]float64{{2, 1, 0}, nil, {1}},
	}
	pts := sas.Stacks()
	if len(pts) != 3 {
		t.Fatalf("want 3 points got %d", len(pts))
	}
	if pts[0].Upper != 3 {
		t.Fatalf("index0 should include the short row value, upper[0]=%v", pts[0].Upper)
	}
	if pts[2].Upper != 0 {
		t.Fatalf("missing tail should be zero contribution, upper[2]=%v", pts[2].Upper)
	}
}

func uid110RenderChart(t *testing.T, rp chart.RendererProvider) (*bytes.Buffer, error) {
	t.Helper()
	graph := chart.Chart{
		Title: "frame time by subsystem",
		Series: []chart.Series{
			chart.StackedAreaSeries{
				Name:    "frame",
				XValues: []float64{0, 1, 2, 3},
				Series: []chart.StackedAreaSeries{
					{Name: "render", YValues: []float64{2, 1, 3, 2}},
					{Name: "script", YValues: []float64{-1, 3, -2, 1}},
					{Name: "audio", YValues: []float64{0.5, 1, 0.2, 0.8}},
				},
			},
		},
	}
	graph.Elements = []chart.Renderable{chart.Legend(&graph)}
	buf := &bytes.Buffer{}
	err := graph.Render(rp, buf)
	return buf, err
}

func TestUID110RenderPNGAndSVG(t *testing.T) {
	pngBuf, err := uid110RenderChart(t, chart.PNG)
	if err != nil {
		t.Fatalf("PNG render: %v", err)
	}
	if pngBuf.Len() == 0 {
		t.Fatal("empty PNG output")
	}
	if img, derr := png.Decode(bytes.NewReader(pngBuf.Bytes())); derr != nil {
		t.Fatalf("PNG decode: %v", derr)
	} else if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("PNG has empty bounds: %v", img.Bounds())
	}

	svgBuf, err := uid110RenderChart(t, chart.SVG)
	if err != nil {
		t.Fatalf("SVG render: %v", err)
	}
	svg := svgBuf.String()
	if !strings.Contains(svg, "render") || !strings.Contains(svg, "script") || !strings.Contains(svg, "audio") {
		t.Fatalf("SVG missing legend layer names")
	}
	if !strings.Contains(svg, "<path") {
		t.Fatalf("SVG expected filled paths")
	}
}

func TestUID110LayerFieldSpellings(t *testing.T) {
	// Layers that name their vector Values/Data/Y instead of YValues.
	sas := chart.StackedAreaSeries{
		XValues: []float64{0, 1},
		Layers: []chart.StackedAreaSeries{
			{Name: "render", Values: []float64{2, 1}},
			{Name: "script", Y: []float64{-1, 3}},
		},
	}
	uid110CheckOraclePoints(t, sas.Stacks())

	// A []chart.Series carrying ContinuousSeries values.
	sas2 := chart.StackedAreaSeries{
		XValues: []float64{0, 1},
		Series: []chart.Series{
			chart.ContinuousSeries{Name: "render", YValues: []float64{2, 1}},
			chart.ContinuousSeries{Name: "script", YValues: []float64{-1, 3}},
		},
	}
	uid110CheckOraclePoints(t, sas2.Stacks())
}

func TestUID110ChartAcceptsContinuousSeriesLiteral(t *testing.T) {
	// Review gap #1: standalone chart Series/Layers accept []ContinuousSeries.
	c := chart.StackedAreaChart{
		Width:   320,
		Height:  200,
		XValues: []float64{0, 1},
		Series: []chart.ContinuousSeries{
			{Name: "render", YValues: []float64{2, 1}},
			{Name: "script", YValues: []float64{-1, 3}},
		},
	}
	buf := &bytes.Buffer{}
	if err := c.Render(chart.SVG, buf); err != nil {
		t.Fatalf("continuous-series chart render: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty svg")
	}

	// Variadic math helper call shapes.
	u, l := chart.StackPositiveNegative([]float64{2, 1}, []float64{-1, 3})
	if u[0] != 2 || l[0] != -1 || u[1] != 4 {
		t.Fatalf("variadic StackPositiveNegative bad: u=%v l=%v", u, l)
	}
	res, err := chart.ComputeStackedArea([]float64{0, 1}, []float64{2, 1}, []float64{-1, 3})
	if err != nil || res.Upper[1] != 4 {
		t.Fatalf("variadic ComputeStackedArea bad: %+v err=%v", res, err)
	}
}

func TestUID110StockChartSinglePointRenders(t *testing.T) {
	graph := chart.Chart{
		Width:  200,
		Height: 120,
		Series: []chart.Series{
			chart.StackedAreaSeries{
				XValues: []float64{9},
				YValues: [][]float64{{2}, {-1}},
			},
		},
	}
	for name, rp := range map[string]chart.RendererProvider{"svg": chart.SVG, "png": chart.PNG} {
		buf := &bytes.Buffer{}
		if err := graph.Render(rp, buf); err != nil {
			t.Fatalf("%s stock-chart single point should render, got %v", name, err)
		}
	}

	// Existing line behavior for a single point is intentionally unchanged:
	// no stacked area => no auto X padding => the standard zero-delta error.
	line := chart.Chart{Series: []chart.Series{chart.ContinuousSeries{
		XValues: []float64{1}, YValues: []float64{1},
	}}}
	if err := line.Render(chart.SVG, &bytes.Buffer{}); err == nil {
		t.Fatal("single-point line chart should retain its zero x-range delta error")
	}
}

func TestUID110TwoTimepointStockChart(t *testing.T) {
	graph := chart.Chart{
		Width:  300,
		Height: 200,
		Series: []chart.Series{
			chart.StackedAreaSeries{
				XValues: []float64{0, 1},
				Series: []chart.StackedAreaSeries{
					{Name: "render", YValues: []float64{2, 1}},
					{Name: "script", YValues: []float64{-1, 3}},
				},
			},
		},
	}
	for name, rp := range map[string]chart.RendererProvider{"svg": chart.SVG, "png": chart.PNG} {
		buf := &bytes.Buffer{}
		if err := graph.Render(rp, buf); err != nil {
			t.Fatalf("%s two-timepoint render: %v", name, err)
		}
		if buf.Len() == 0 {
			t.Fatalf("%s empty output", name)
		}
	}
}

func TestUID110StandaloneChartSVGPNGAndSinglePoint(t *testing.T) {
	c := chart.StackedAreaChart{
		Width:      640,
		Height:     320,
		XValues:    []float64{0, 1},
		YValues:    [][]float64{{2, 1}, {-1, 3}},
		Names:      []string{"render", "script"},
		ShowLegend: true,
	}
	for name, rp := range map[string]chart.RendererProvider{"svg": chart.SVG, "png": chart.PNG} {
		buf := &bytes.Buffer{}
		if err := c.Render(rp, buf); err != nil {
			t.Fatalf("%s render: %v", name, err)
		}
		if buf.Len() == 0 {
			t.Fatalf("%s empty output", name)
		}
	}

	one := chart.StackedAreaChart{
		Width: 200, Height: 120,
		XValues: []float64{42},
		Series: []chart.StackedAreaSeries{
			{Name: "render", YValues: []float64{2}},
			{Name: "script", YValues: []float64{-1}},
		},
	}
	buf := &bytes.Buffer{}
	if err := one.Render(chart.SVG, buf); err != nil {
		t.Fatalf("single point chart render: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("single point empty svg")
	}
}
