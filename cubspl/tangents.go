package cubspl

// estimate tangents based on given vertices
type TanEstimator interface {
	Estimate(vertsx, vertsy []float64) (entryTansx, entryTansy []float64, exitTansx, exitTansy []float64)
}

type CardinalTan struct {
	tension float64
}

func NewCardinalTan(tension float64) CardinalTan {
	return CardinalTan{tension: tension}
}

func NewCatmullRomTan() CardinalTan {
	return NewCardinalTan(0)
}

func (ct CardinalTan) Estimate(vertsx, vertsy []float64) (entryTansx, entryTansy []float64, exitTansx, exitTansy []float64) {
	n := len(vertsx)
	exitTansx = make([]float64, n)
	exitTansy = make([]float64, n)
	// single tangent
	entryTansx = exitTansx
	entryTansy = exitTansx

	if n < 2 {
		return
	}

	b := (1 - ct.tension) / 2
	exitTansx[0] = b * (vertsx[1] - vertsx[0])
	exitTansy[0] = b * (vertsy[1] - vertsy[0])
	for i := 1; i < n-1; i++ {
		exitTansx[i] = b * (vertsx[i+1] - vertsx[i-1])
		exitTansy[i] = b * (vertsy[i+1] - vertsy[i-1])
	}
	exitTansx[n-1] = b * (vertsx[n-1] - vertsx[n-2])
	exitTansy[n-1] = b * (vertsy[n-1] - vertsy[n-2])

	return
}
