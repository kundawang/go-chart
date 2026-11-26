package chart

import (
	"image"
	"image/png"
	"io"

	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// pngDecode decodes a PNG image from r.
func pngDecode(r io.Reader) (image.Image, error) {
	return png.Decode(r)
}

// recordingRenderer captures the boxes and stroked line segments issued to
// it so tests can assert on candle geometry without rendering pixels.
type recordingRenderer struct {
	boxes       []Box
	strokeColor drawing.Color
	fillColor   drawing.Color
	path        []image.Point
}

func newRecordingRenderer() *recordingRenderer {
	return &recordingRenderer{}
}

func (rr *recordingRenderer) ResetStyle() {}

func (rr *recordingRenderer) GetDPI() float64 { return DefaultDPI }

func (rr *recordingRenderer) SetDPI(float64) {}

func (rr *recordingRenderer) SetClassName(string) {}

func (rr *recordingRenderer) SetStrokeColor(c drawing.Color) { rr.strokeColor = c }

func (rr *recordingRenderer) SetFillColor(c drawing.Color) { rr.fillColor = c }

func (rr *recordingRenderer) SetStrokeWidth(float64) {}

func (rr *recordingRenderer) SetStrokeDashArray([]float64) {}

func (rr *recordingRenderer) MoveTo(x, y int) {
	rr.path = append(rr.path, image.Point{X: x, Y: y})
}

func (rr *recordingRenderer) LineTo(x, y int) {
	rr.path = append(rr.path, image.Point{X: x, Y: y})
}

func (rr *recordingRenderer) QuadCurveTo(_, _, _, _ int) {}

func (rr *recordingRenderer) ArcTo(_, _ int, _, _, _, _ float64) {}

func (rr *recordingRenderer) Close() {}

func (rr *recordingRenderer) Stroke() {
	rr.path = nil
}

func (rr *recordingRenderer) Fill() {}

func (rr *recordingRenderer) FillStroke() {
	if len(rr.path) >= 4 {
		// Draw.Box emits the four corners in order.
		tl, tr, br, bl := rr.path[0], rr.path[1], rr.path[2], rr.path[3]
		rr.boxes = append(rr.boxes, Box{
			Top:    tl.Y,
			Left:   tl.X,
			Right:  tr.X,
			Bottom: bl.Y,
		})
		_ = br
	}
	rr.path = nil
}

func (rr *recordingRenderer) Circle(float64, int, int) {}

func (rr *recordingRenderer) SetFont(*truetype.Font) {}

func (rr *recordingRenderer) SetFontColor(drawing.Color) {}

func (rr *recordingRenderer) SetFontSize(float64) {}

func (rr *recordingRenderer) Text(string, int, int) {}

func (rr *recordingRenderer) MeasureText(string) Box { return Box{} }

func (rr *recordingRenderer) SetTextRotation(float64) {}

func (rr *recordingRenderer) ClearTextRotation() {}

func (rr *recordingRenderer) Save(io.Writer) error { return nil }

var _ Renderer = (*recordingRenderer)(nil)
