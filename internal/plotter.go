package internal

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func Histogram(results []*SpeedResult, timeUnit string) {
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
