package chart_test

import (
	"bytes"
	"encoding/xml"
	"math"
	"regexp"
	"strconv"
	"testing"
	"time"

	chart "github.com/wcharczuk/go-chart/v2"
)

// This file verifies the optional tick-label collision avoidance layout
// end-to-end through the exported API. All helpers carry the avdVerify prefix
// to avoid identifier collisions with any other test files.

func avdVerifyFont(t *testing.T) *chart.Style {
	t.Helper()
	f, err := chart.GetDefaultFont()
	if err != nil {
		t.Fatalf("default font: %v", err)
	}
	return &chart.Style{Font: f, FontSize: chart.DefaultAxisFontSize}
}

func avdVerifyDateTicks() []chart.Tick {
	days := []string{
		"2024-01-01", "2024-01-02", "2024-01-03",
		"2024-01-04", "2024-01-05", "2024-01-06",
	}
	ticks := make([]chart.Tick, 0, len(days))
	for _, day := range days {
		tt, _ := time.Parse("2006-01-02", day)
		ticks = append(ticks, chart.Tick{Value: chart.TimeToFloat64(tt), Label: day})
	}
	return ticks
}

func avdVerifyDisjoint(boxes []chart.Box) bool {
	for i := 0; i < len(boxes); i++ {
		for j := i + 1; j < len(boxes); j++ {
			a, b := boxes[i], boxes[j]
			if a.Left < b.Right && a.Right > b.Left &&
				a.Top < b.Bottom && a.Bottom > b.Top {
				return false
			}
		}
	}
	return true
}

func avdVerifyRenderer(t *testing.T, kind string, w, h int) chart.Renderer {
	t.Helper()
	var (
		r   chart.Renderer
		err error
	)
	if kind == "png" {
		r, err = chart.PNG(w, h)
	} else {
		r, err = chart.SVG(w, h)
	}
	if err != nil {
		t.Fatalf("%s renderer: %v", kind, err)
	}
	r.SetDPI(chart.DefaultDPI)
	return r
}

// Headline scenario: a 120px projection with six adjacent long date labels
// keeps the first and last labels with pairwise non-overlapping boxes.
func TestAvdVerifyNarrow120EndsAndBoxes(t *testing.T) {
	ticks := avdVerifyDateTicks()
	style := *avdVerifyFont(t)
	r := avdVerifyRenderer(t, "svg", 160, 100)

	ra := &chart.ContinuousRange{Min: ticks[0].Value, Max: ticks[5].Value, Domain: 120}
	canvas := chart.NewBox(0, 0, 120, 60)

	layout := chart.ComputeTickLabelLayout(r, chart.AxisOrientationHorizontal, canvas, style, ticks,
		chart.TickLabelLayoutOptions{Project: func(t chart.Tick) int { return ra.Translate(t.Value) }})

	if got := layout.VisibleIndices(); len(got) < 2 || got[0] != 0 || got[len(got)-1] != 5 {
		t.Fatalf("expected first/last retained, got %v", got)
	}
	if !layout.Visible[0] || !layout.Visible[5] {
		t.Fatalf("end anchors must be visible: %+v", layout.Visible)
	}
	if !avdVerifyDisjoint(layout.Boxes) {
		t.Fatalf("visible boxes overlap: %+v", layout.Boxes)
	}

	xa := chart.XAxis{AvoidOverlappingLabels: true}
	got := xa.GetVisibleTicks(r, canvas, ra, style, ticks)
	if len(got) != len(layout.VisibleTicks) {
		t.Fatalf("axis method returned %d ticks, layout %d", len(got), len(layout.VisibleTicks))
	}
}

// SVG and PNG renderers must make the exact same selection because layout
// measurement is canonical.
func TestAvdVerifySVGPNGSelectionMatches(t *testing.T) {
	ticks := avdVerifyDateTicks()
	style := *avdVerifyFont(t)
	ra := &chart.ContinuousRange{Min: ticks[0].Value, Max: ticks[5].Value, Domain: 120}
	canvas := chart.NewBox(0, 0, 120, 60)

	xa := chart.XAxis{AvoidOverlappingLabels: true}
	svgTicks := xa.GetVisibleTicks(avdVerifyRenderer(t, "svg", 160, 100), canvas, ra, style, ticks)
	pngTicks := xa.GetVisibleTicks(avdVerifyRenderer(t, "png", 160, 100), canvas, ra, style, ticks)

	if len(svgTicks) != len(pngTicks) {
		t.Fatalf("svg kept %d labels, png %d", len(svgTicks), len(pngTicks))
	}
	for i := range svgTicks {
		if svgTicks[i].Label != pngTicks[i].Label || svgTicks[i].Value != pngTicks[i].Value {
			t.Fatalf("selection mismatch at %d: svg=%q png=%q", i, svgTicks[i].Label, pngTicks[i].Label)
		}
	}
}

// A wide axis must keep every label.
func TestAvdVerifyWideKeepsAll(t *testing.T) {
	ticks := avdVerifyDateTicks()
	style := *avdVerifyFont(t)
	r := avdVerifyRenderer(t, "svg", chart.DefaultChartWidth+40, 120)

	ra := &chart.ContinuousRange{Min: ticks[0].Value, Max: ticks[5].Value, Domain: chart.DefaultChartWidth}
	canvas := chart.NewBox(0, 0, chart.DefaultChartWidth, 80)

	layout := chart.ComputeTickLabelLayout(r, chart.AxisOrientationHorizontal, canvas, style, ticks,
		chart.TickLabelLayoutOptions{Project: func(t chart.Tick) int { return ra.Translate(t.Value) }})
	if got := layout.VisibleIndices(); len(got) != 6 {
		t.Fatalf("wide canvas should keep all 6 labels, got %v", got)
	}
}

// The Y axis thins dense labels, keeps end/zero anchors and never overlaps.
func TestAvdVerifyYAxisZeroAnchor(t *testing.T) {
	labels := []string{"-12000.00", "-9000.00", "-6000.00", "-3000.00", "0.00", "3000.00", "6000.00", "9000.00", "12000.00"}
	ticks := make([]chart.Tick, len(labels))
	for i, label := range labels {
		ticks[i] = chart.Tick{Value: float64(i - 4), Label: label}
	}
	style := *avdVerifyFont(t)
	r := avdVerifyRenderer(t, "svg", 220, 120)

	ra := &chart.ContinuousRange{Min: -4, Max: 4, Domain: 60}
	canvas := chart.NewBox(0, 40, 180, 100)
	layout := chart.ComputeTickLabelLayout(r, chart.AxisOrientationVertical, canvas, style, ticks,
		chart.TickLabelLayoutOptions{
			YAxisType: chart.YAxisPrimary,
			Project:   func(t chart.Tick) int { return ra.Translate(t.Value) },
		})

	if !layout.Visible[0] || !layout.Visible[len(ticks)-1] {
		t.Fatalf("end anchors must be retained: %+v", layout.Visible)
	}
	if !layout.Visible[4] {
		t.Fatalf("zero tick must be retained: %+v", layout.Visible)
	}
	if !avdVerifyDisjoint(layout.Boxes) {
		t.Fatalf("y label boxes overlap: %+v", layout.Boxes)
	}
	if len(layout.VisibleTicks) >= len(ticks) {
		t.Fatalf("dense y axis should thin labels, kept %d", len(layout.VisibleTicks))
	}

	if !chart.IsZeroTick(chart.Tick{Value: 0}) || chart.IsZeroTick(chart.Tick{Value: 0.5}) {
		t.Fatal("IsZeroTick classification wrong")
	}
}

// Rotated labels are handled via their rotated bounding boxes.
func TestAvdVerifyRotated45And90(t *testing.T) {
	ticks := avdVerifyDateTicks()
	base := *avdVerifyFont(t)

	for _, rot := range []float64{45, -45, 90} {
		style := base
		style.TextRotationDegrees = rot
		r := avdVerifyRenderer(t, "svg", 320, 260)
		ra := &chart.ContinuousRange{Min: ticks[0].Value, Max: ticks[5].Value, Domain: 240}
		canvas := chart.NewBox(0, 0, 240, 120)

		layout := chart.ComputeTickLabelLayout(r, chart.AxisOrientationHorizontal, canvas, style, ticks,
			chart.TickLabelLayoutOptions{Project: func(t chart.Tick) int { return ra.Translate(t.Value) }})
		if !layout.Visible[0] || !layout.Visible[5] {
			t.Fatalf("rot=%v: end anchors missing %v", rot, layout.VisibleIndices())
		}
		if !avdVerifyDisjoint(layout.Boxes) {
			t.Fatalf("rot=%v: boxes overlap %+v", rot, layout.Boxes)
		}
	}
}

// The pure selection helper is stable across runs and prioritizes ends.
func TestAvdVerifySelectionStable(t *testing.T) {
	boxes := make([]chart.Box, 6)
	for i := range boxes {
		cx := i * 20
		boxes[i] = chart.Box{Top: 0, Left: cx - 20, Right: cx + 20, Bottom: 12}
	}
	prio := []int{
		chart.TickLabelPriorityFirst, 0, 0, 0, 0, chart.TickLabelPriorityLast,
	}
	a := chart.SelectNonOverlappingBoxes(boxes, prio, 0)
	b := chart.SelectNonOverlappingBoxes(boxes, prio, 0)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("selection not stable: %v vs %v", a, b)
		}
	}
	if !a[0] || !a[5] {
		t.Fatalf("ends must be prioritized: %v", a)
	}
	var kept []chart.Box
	for i, v := range a {
		if v {
			kept = append(kept, boxes[i])
		}
	}
	if !avdVerifyDisjoint(kept) {
		t.Fatal("selected boxes overlap")
	}
}

// Disabled by default: an axis with no switch reports disabled and renders
// every label.
func TestAvdVerifyOffByDefaultRendersAll(t *testing.T) {
	ticks := avdVerifyDateTicks()
	style := *avdVerifyFont(t)
	ra := &chart.ContinuousRange{Min: ticks[0].Value, Max: ticks[5].Value, Domain: 120}
	canvas := chart.NewBox(0, 0, 120, 80)

	xa := chart.XAxis{}
	r := avdVerifyRenderer(t, "svg", 160, 120)
	var buf bytes.Buffer
	xa.Render(r, canvas, ra, style, ticks)
	if err := r.Save(&buf); err != nil {
		t.Fatal(err)
	}
	texts := avdVerifyDateTexts(t, buf.Bytes())
	if len(texts) != 6 {
		t.Fatalf("default axis should render all 6 labels, got %d", len(texts))
	}
}

// End-to-end chart at 120px keeps the end date labels with disjoint tight
// SVG boxes, both with and without the primary Y axis.
func TestAvdVerifyChart120EndToEnd(t *testing.T) {
	ticks := avdVerifyDateTicks()
	var xv, yv []float64
	for i, tk := range ticks {
		xv = append(xv, tk.Value)
		yv = append(yv, float64(i)+1)
	}
	style := *avdVerifyFont(t)

	for _, showY := range []bool{false, true} {
		c := chart.Chart{
			Width:  120,
			Height: 200,
			XAxis:  chart.XAxis{Ticks: ticks, AvoidOverlappingLabels: true},
			YAxis:  chart.YAxis{Style: chart.Style{Hidden: !showY}},
			YAxisSecondary: chart.YAxis{
				Style: chart.Style{Hidden: true},
			},
			Series: []chart.Series{chart.ContinuousSeries{XValues: xv, YValues: yv}},
		}
		var buf bytes.Buffer
		if err := c.Render(chart.SVG, &buf); err != nil {
			t.Fatalf("showY=%v render: %v", showY, err)
		}
		texts := avdVerifyDateTexts(t, buf.Bytes())
		if len(texts) < 2 {
			t.Fatalf("showY=%v: expected end labels, got %d", showY, len(texts))
		}
		if texts[0].Body != "2024-01-01" || texts[len(texts)-1].Body != "2024-01-06" {
			t.Fatalf("showY=%v: ends wrong: %q .. %q", showY, texts[0].Body, texts[len(texts)-1].Body)
		}
		boxes := avdVerifyHorizontalBoxes(t, texts, style)
		if !avdVerifyDisjoint(boxes) {
			t.Fatalf("showY=%v: emitted boxes overlap %+v", showY, boxes)
		}
	}
}

// Top-level Chart switches and the Style-based switch must enable the same
// end-label preserving layout.
func TestAvdVerifyChartAndStyleSwitches(t *testing.T) {
	ticks := avdVerifyDateTicks()
	var xv, yv []float64
	for i, tk := range ticks {
		xv = append(xv, tk.Value)
		yv = append(yv, float64(i)+1)
	}
	series := []chart.Series{chart.ContinuousSeries{XValues: xv, YValues: yv}}
	hiddenY := chart.YAxis{Style: chart.Style{Hidden: true}}
	hiddenY2 := chart.YAxis{Style: chart.Style{Hidden: true}}

	// Chart-level alias switch.
	c1 := chart.Chart{
		Width: 120, Height: 200,
		AvoidOverlappingTickLabels: true,
		XAxis:                      chart.XAxis{Ticks: ticks},
		YAxis:                      hiddenY,
		YAxisSecondary:             hiddenY2,
		Series:                     series,
	}
	var b1 bytes.Buffer
	if err := c1.Render(chart.SVG, &b1); err != nil {
		t.Fatal(err)
	}
	if texts := avdVerifyDateTexts(t, b1.Bytes()); len(texts) < 2 {
		t.Fatalf("chart-level switch: expected end labels, got %d", len(texts))
	}

	// Style-level switch on the tick style.
	c2 := chart.Chart{
		Width: 120, Height: 200,
		XAxis: chart.XAxis{
			Ticks:     ticks,
			TickStyle: chart.Style{AvoidOverlappingLabels: true},
		},
		YAxis:          hiddenY,
		YAxisSecondary: hiddenY2,
		Series:         series,
	}
	var b2 bytes.Buffer
	if err := c2.Render(chart.SVG, &b2); err != nil {
		t.Fatal(err)
	}
	if texts := avdVerifyDateTexts(t, b2.Bytes()); len(texts) < 2 {
		t.Fatalf("style switch: expected end labels, got %d", len(texts))
	}
}

func TestAvdVerifyChartWideKeepsAll(t *testing.T) {
	ticks := avdVerifyDateTicks()
	var xv, yv []float64
	for i, tk := range ticks {
		xv = append(xv, tk.Value)
		yv = append(yv, float64(i)+1)
	}
	c := chart.Chart{
		Width:  chart.DefaultChartWidth,
		Height: 300,
		XAxis:  chart.XAxis{Ticks: ticks, AvoidOverlappingLabels: true},
		YAxis:  chart.YAxis{Style: chart.Style{Hidden: true}},
		YAxisSecondary: chart.YAxis{
			Style: chart.Style{Hidden: true},
		},
		Series: []chart.Series{chart.ContinuousSeries{XValues: xv, YValues: yv}},
	}
	var buf bytes.Buffer
	if err := c.Render(chart.SVG, &buf); err != nil {
		t.Fatal(err)
	}
	if texts := avdVerifyDateTexts(t, buf.Bytes()); len(texts) != 6 {
		t.Fatalf("wide chart should keep 6 labels, got %d", len(texts))
	}
}

// --- SVG parsing / tight box reconstruction ---

type avdVerifyText struct {
	X         float64 `xml:"x,attr"`
	Y         float64 `xml:"y,attr"`
	Transform string  `xml:"transform,attr"`
	Body      string  `xml:",chardata"`
}

func avdVerifyDateTexts(t *testing.T, raw []byte) []avdVerifyText {
	t.Helper()
	var root struct {
		Texts []avdVerifyText `xml:"text"`
	}
	if err := xml.Unmarshal(raw, &root); err != nil {
		t.Fatalf("parse svg: %v", err)
	}
	var out []avdVerifyText
	for _, tx := range root.Texts {
		if len(tx.Body) == 10 && tx.Body[4] == '-' && tx.Body[7] == '-' {
			out = append(out, tx)
		}
	}
	return out
}

func avdVerifyMeasure(t *testing.T, label string, style chart.Style) (int, int) {
	t.Helper()
	style.TextRotationDegrees = 0
	r := avdVerifyRenderer(t, "svg", 16, 16)
	b := chart.Draw.MeasureText(r, label, style)
	return b.Width(), b.Height()
}

var avdVerifyRotateRE = regexp.MustCompile(`rotate\(\s*([-\d.]+)\s*,\s*([-\d.]+)\s*,\s*([-\d.]+)\s*\)`)

func avdVerifyHorizontalBoxes(t *testing.T, texts []avdVerifyText, style chart.Style) []chart.Box {
	t.Helper()
	boxes := make([]chart.Box, 0, len(texts))
	for _, tx := range texts {
		w, h := avdVerifyMeasure(t, tx.Body, style)
		x := int(math.Round(tx.X))
		y := int(math.Round(tx.Y))
		if tx.Transform == "" {
			boxes = append(boxes, chart.Box{Top: y - h, Left: x, Right: x + w, Bottom: y})
			continue
		}
		m := avdVerifyRotateRE.FindStringSubmatch(tx.Transform)
		if m == nil {
			boxes = append(boxes, chart.Box{Top: y - h, Left: x, Right: x + w, Bottom: y})
			continue
		}
		deg, _ := strconv.ParseFloat(m[1], 64)
		px, _ := strconv.ParseFloat(m[2], 64)
		py, _ := strconv.ParseFloat(m[3], 64)
		theta := chart.DegreesToRadians(deg)
		corners := [4][2]float64{
			{px, py - float64(h)},
			{px + float64(w), py - float64(h)},
			{px + float64(w), py},
			{px, py},
		}
		var minX, minY = math.MaxFloat64, math.MaxFloat64
		var maxX, maxY = -math.MaxFloat64, -math.MaxFloat64
		for _, c := range corners {
			dx, dy := c[0]-px, c[1]-py
			rx := px + dx*math.Cos(theta) - dy*math.Sin(theta)
			ry := py + dx*math.Sin(theta) + dy*math.Cos(theta)
			minX, maxX = math.Min(minX, rx), math.Max(maxX, rx)
			minY, maxY = math.Min(minY, ry), math.Max(maxY, ry)
		}
		boxes = append(boxes, chart.Box{
			Top: int(math.Floor(minY)), Left: int(math.Floor(minX)),
			Right: int(math.Ceil(maxX)), Bottom: int(math.Ceil(maxY)),
		})
	}
	return boxes
}
