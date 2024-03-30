package internal

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func VerifyPlotFormats(formats string) ([]string, error) {
	validFormats := []string{"hist", "histogram", "box", "boxplot", "bar", "errorbar"}
	formatList := strings.Split(strings.ToLower(formats), ",")
	for _, f := range formatList {
		if !slices.Contains(validFormats, f) {
			return nil, fmt.Errorf("invalid export format: %s", f)
		}
	}
	return formatList, nil
}

func histogram(results []*SpeedResult, timeUnit string) {
	p := plot.New()
	p.Title.Text = "Histogram"
	p.X.Label.Text = timeUnit

	for i, result := range results {
		v := make(plotter.Values, len(result.Times))
		copy(v, result.Times)

		h, err := plotter.NewHist(v, 16)
		if err != nil {
			panic(err)
		}
		// h.Normalize(1)
		h.FillColor = colors[i%len(colors)]
		p.Legend.Add(result.Command, h)
		p.Add(h)
	}
	p.Legend.Top = true

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 4*vg.Inch, "hist.png"); err != nil {
		panic(err)
	}
}

func barPlot(results []*SpeedResult, timeUnit string) {
	p := plot.New()
	p.Title.Text = "Bar Chart"
	p.Y.Label.Text = fmt.Sprintf("Mean times (in %s)", timeUnit)
	meanTimes := make(plotter.Values, len(results))
	copy(meanTimes, MapFunc[[]*SpeedResult, []float64](func(sr *SpeedResult) float64 { return sr.AverageElapsed }, results))

	w := vg.Points(20)
	bars, err := plotter.NewBarChart(meanTimes, w)
	if err != nil {
		panic(err)
	}
	bars.LineStyle.Width = vg.Length(0)
	bars.Color = plotutil.Color(0)

	p.Add(bars)

	p.NominalX(MapFunc[[]*SpeedResult, []string](func(r *SpeedResult) string { return r.Command }, results)...)

	barWidth := max(3, len(results))
	if err := p.Save(font.Length(barWidth)*vg.Inch, 3*vg.Inch, "barchart.png"); err != nil {
		panic(err)
	}
}

func Plot(plotFormats []string, results []*SpeedResult, timeUnit time.Duration) {
	if slices.Contains(plotFormats, "all") {
		plotFormats = []string{"histogram", "bar", "errorbar", "boxplot"}
	}
	for _, plotFormat := range plotFormats {
		switch plotFormat {
		case "hist", "histogram":
			histogram(results, timeUnit.String()[1:])
		case "bar":
			barPlot(results, timeUnit.String()[1:])
		}
	}
}
