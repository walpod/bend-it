package cubspl

type SplineFn2d func(t float64) (x, y float64)

type HermiteSpline2d struct {
	vertices []HermiteVertex2d
	knots    []float64
}

func NewHermiteSpline2d(vertices []HermiteVertex2d) *HermiteSpline2d {
	return &HermiteSpline2d{vertices: vertices}
}

func (hs HermiteSpline2d) VertexCnt() int {
	return len(hs.vertices)
}

func (hs HermiteSpline2d) SegmentCnt() int {
	sc := len(hs.vertices) - 1
	if sc < 0 {
		return 0
	} else {
		return sc
	}
}
