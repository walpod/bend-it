package cubic

// find tangents based on given vertices and knots
type TanFinder2d interface {
	Find(vertsx, vertsy []float64, knots []float64) (
		entryTansx, entryTansy []float64, exitTansx, exitTansy []float64)
}

type CardinalTanf2d struct {
	tension float64
}

func NewCardinalTanf2d(tension float64) CardinalTanf2d {
	return CardinalTanf2d{tension: tension}
}

func NewCatmullRomTanf2d() CardinalTanf2d {
	return NewCardinalTanf2d(0)
}

func (ct CardinalTanf2d) Find(vertsx, vertsy []float64, knots []float64) (
	entryTansx, entryTansy []float64, exitTansx, exitTansy []float64) {

	n := len(vertsx)
	exitTansx = make([]float64, n)
	exitTansy = make([]float64, n)

	if len(knots) == 0 {
		// uniform -> single tangent
		entryTansx = exitTansx
		entryTansy = exitTansy
	} else {
		// non-uniform -> double tangent
		entryTansx = make([]float64, n)
		entryTansy = make([]float64, n)
	}

	if n < 2 {
		return
	}

	// first and last tangents use adjacent vertices, all others use vertices i+1 and i-1
	b := (1 - ct.tension) / 2
	exitTansx[0] = b * (vertsx[1] - vertsx[0])
	exitTansy[0] = b * (vertsy[1] - vertsy[0])
	for i := 1; i < n-1; i++ {
		exitTansx[i] = b * (vertsx[i+1] - vertsx[i-1])
		exitTansy[i] = b * (vertsy[i+1] - vertsy[i-1])
	}
	exitTansx[n-1] = b * (vertsx[n-1] - vertsx[n-2])
	exitTansy[n-1] = b * (vertsy[n-1] - vertsy[n-2])

	// non-uniform: copy exit-tangents to entry, then modify tangent lengths to reciprocal of segment length
	copy(entryTansx, exitTansx)
	copy(entryTansy, exitTansy)
	if len(knots) > 0 {
		for i := 0; i < n-1; i++ {
			segmLen := knots[i+1] - knots[i]
			exitTansx[i] /= segmLen
			exitTansy[i] /= segmLen
			entryTansx[i+1] /= segmLen
			entryTansy[i+1] /= segmLen
		}
	}

	return
}
