go-chart
========
[![Continuous Integration](https://github.com/wcharczuk/go-chart/actions/workflows/ci.yml/badge.svg)](https://github.com/wcharczuk/go-chart/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/wcharczuk/go-chart)](https://goreportcard.com/report/github.com/wcharczuk/go-chart)

Package `chart` is a very simple golang native charting library that supports timeseries and continuous line charts.

Master should now be on the v3.x codebase, which overhauls the api significantly. Per usual, see `examples` for more information.

# Installation

To install `chart` run the following:

```bash
> go get github.com/wcharczuk/go-chart/v2@latest
```

Most of the components are interchangeable so feel free to crib whatever you want.

# Output Examples

Spark Lines:

![](https://raw.githubusercontent.com/wcharczuk/go-chart/master/_images/tvix_ltm.png)

Single axis:

![](https://raw.githubusercontent.com/wcharczuk/go-chart/master/_images/goog_ltm.png)

Two axis:

![](https://raw.githubusercontent.com/wcharczuk/go-chart/master/_images/two_axis.png)

# Other Chart Types

Pie Chart:

![](https://raw.githubusercontent.com/wcharczuk/go-chart/master/_images/pie_chart.png)

The code for this chart can be found in `examples/pie_chart/main.go`.

Stacked Bar:

![](https://raw.githubusercontent.com/wcharczuk/go-chart/master/_images/stacked_bar.png)

The code for this chart can be found in `examples/stacked_bar/main.go`.

Positive / Negative Separated Stacked Area:

`chart.StackedAreaSeries` accumulates positive and negative components
independently on a shared, increasing X grid. Positives stack upward from the
zero line and negatives stack downward, so a `+2` and a `-1` at the same point
produce cumulative boundaries `2` and `-1` (they never cancel into `1`). Each
layer's upper and lower boundaries and the band between them are drawn, the Y
axis spans the cumulative extrema, and the standard legend expands to one
entry per layer. The same configuration renders to SVG or PNG.

```go
graph := chart.Chart{
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
```

`ComputeStackedArea` provides the same accumulation without rendering. The
code for a full frame-time example can be found in
`examples/stacked_area/main.go`.

# Code Examples

Actual chart configurations and examples can be found in the `./examples/` directory. They are simple CLI programs that write to `output.png` (they are also updated with `go generate`.

# Usage

Everything starts with the `chart.Chart` object. The bare minimum to draw a chart would be the following:

```golang

import (
    ...
    "bytes"
    ...
    "github.com/wcharczuk/go-chart/v2" //exposes "chart"
)

graph := chart.Chart{
    Series: []chart.Series{
        chart.ContinuousSeries{
            XValues: []float64{1.0, 2.0, 3.0, 4.0},
            YValues: []float64{1.0, 2.0, 3.0, 4.0},
        },
    },
}

buffer := bytes.NewBuffer([]byte{})
err := graph.Render(chart.PNG, buffer)
```

Explanation of the above: A `chart` can have many `Series`, a `Series` is a collection of things that need to be drawn according to the X range and the Y range(s).

Here, we have a single series with x range values as float64s, rendered to a PNG. Note; we can pass any type of `io.Writer` into `Render(...)`, meaning that we can render the chart to a file or a resonse or anything else that implements `io.Writer`.

# API Overview

Everything on the `chart.Chart` object has defaults that can be overriden. Whenever a developer sets a property on the chart object, it is to be assumed that value will be used instead of the default.

The best way to see the api in action is to look at the examples in the `./_examples/` directory.

# Design Philosophy

I wanted to make a charting library that used only native golang, that could be stood up on a server (i.e. it had built in fonts).

The goal with the API itself is to have the "zero value be useful", and to require the user to not code more than they absolutely needed.

# Contributions

Contributions are welcome though this library is in a holding pattern for the forseable future.
