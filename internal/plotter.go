package internal

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func VerifyPlotFormats(formats string) ([]string, error) {
	validFormats := []string{"hist", "histogram", "box", "boxplot", "bar", "errorbar", "bubble"}
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

func Plot(plotFormats []string, results []*SpeedResult, timeUnit time.Duration) {
	
}