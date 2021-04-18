package cubic

// entry and exit tangents for given vertex
type VertexTan2d interface {
	EntryTan() (lx, ly float64)
	ExitTan() (mx, my float64)
}

type SingleTan2d struct {
	Mx, My float64
}

func NewSingleTan2d(mx float64, my float64) *SingleTan2d {
	return &SingleTan2d{Mx: mx, My: my}
}

func (st *SingleTan2d) EntryTan() (lx, ly float64) {
	// entry = exit tangent
	return st.Mx, st.My
}

func (st *SingleTan2d) ExitTan() (mx, my float64) {
	return st.Mx, st.My
}

/*
type HermiteBuilder2d struct {
	vertsx, vertsy []float64
	tangents       []VertexTan2d
	// TODO slopeEstimator
	knots []float64 // TODO uniform - non-uniform
}
*/
/*
func NewHermiteSplineBuilder2d(vertsx, vertsy []float64, tangents []VertexTan2d, knots []float64) *HermiteBuilder2d {
	n := len(vertsx)
	if len(vertsy) != n || len(tangents) != n || len(knots) != n {
		panic("versv, vertsy, tangents and knots must all have the same length")
	}
	return &HermiteBuilder2d{vertsx: vertsx, vertsy: vertsy, tangents: tangents, knots: knots}
}
*/

/*
func (hs *HermiteBuilder2d) VertexCnt() int {
	return len(hs.vertsx)
}

func (hs *HermiteBuilder2d) SegmentCnt() int {
	if len(hs.vertsx) > 0 {
		return len(hs.vertsx) - 1
	} else {
		return 0
	}
}

func (hs *HermiteBuilder2d) Knot0() float64 {
	if len(hs.knots) == 0 {
		return 0 // TODO
	} else {
		return hs.knots[0]
	}
}

func (hs *HermiteBuilder2d) KnotN() float64 {
	lk := len(hs.knots)
	if lk == 0 {
		return -1 // TODO
	} else {
		return hs.knots[lk-1]
	}
}

func (hs *HermiteBuilder2d) Add(vertx, verty float64, tangent VertexTan2d) {
	hs.vertsx = append(hs.vertsx, vertx)
	hs.vertsy = append(hs.vertsy, verty)
	hs.tangents = append(hs.tangents, tangent)
	hs.knots = append(hs.knots, hs.KnotN()+1) // TODO currently for uniform splines
}

func (hs *HermiteBuilder2d) Build() bendit.Fn2d {
	n := hs.VertexCnt()
	entryTansx := make([]float64, n)
	entryTansy := make([]float64, n)
	exitTansx := make([]float64, n)
	exitTansy := make([]float64, n)
	for i := 0; i < len(hs.tangents); i++ {
		entryTanx, entryTany := hs.tangents[i].EntryTan()
		entryTansx[i] = entryTanx
		entryTansy[i] = entryTany
		exitTanx, exitTany := hs.tangents[i].ExitTan()
		exitTansx[i] = exitTanx
		exitTansy[i] = exitTany
	}
	return BuildHermiteSpline2d(hs.vertsx, hs.vertsy, entryTansx, entryTansy, exitTansx, exitTansy, hs.knots)
}
*/
