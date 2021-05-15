package cubic

import (
	"github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

/*
// TODO currently deactivated
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
*/

// HermiteTanFinder2d finds tangents based on given vertices and knots
type HermiteTanFinder2d interface {
	Find(vertsx, vertsy []float64, knots *bendit.Knots) (
		entryTansx, entryTansy []float64, exitTansx, exitTansy []float64)
}

type HermiteSpline2d struct {
	vertsx, vertsy         []float64
	tanFinder              HermiteTanFinder2d
	entryTansx, entryTansy []float64
	exitTansx, exitTansy   []float64
	// TODO tangents       []VertexTan2d
	knots *bendit.Knots
	canon *CanonicalSpline2d
}

func NewHermiteSpline2d(vertsx []float64, vertsy []float64,
	entryTansx []float64, entryTansy []float64, exitTansx []float64, exitTansy []float64,
	knots *bendit.Knots) *HermiteSpline2d {

	herm := &HermiteSpline2d{vertsx: vertsx, vertsy: vertsy,
		entryTansx: entryTansx, entryTansy: entryTansy, exitTansx: exitTansx, exitTansy: exitTansy, knots: knots}
	herm.Build()
	return herm
}

func NewHermiteSplineTanFinder2d(vertsx []float64, vertsy []float64, tanFinder HermiteTanFinder2d, knots *bendit.Knots) *HermiteSpline2d {
	herm := &HermiteSpline2d{vertsx: vertsx, vertsy: vertsy, tanFinder: tanFinder, knots: knots}
	herm.Build()
	return herm
}

/* TODO variant with  tangents []VertexTan2d
func (hs *HermiteSpline2d) Add(vertx, verty float64, tangent VertexTan2d) {
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

func (hs *HermiteSpline2d) SegmentCnt() int {
	return len(hs.vertsx) - 1
}

func (hs *HermiteSpline2d) Knots() *bendit.Knots {
	return hs.knots
}

// Build hermite spline by mapping to canonical representation
func (hs *HermiteSpline2d) Build() {
	hs.canon = hs.Canonical()
}

func (hs *HermiteSpline2d) Canonical() *CanonicalSpline2d {
	n := len(hs.vertsx)
	/*if len(vertsy) != n || len(entryTansx) != n || len(entryTansy) != n || len(exitTansx) != n || len(exitTansy) != n ||
		(len(knots) > 0 && len(knots) != n) {
		panic("versv, vertsy, all tangents and (optional) knots must have the same length")
	}*/
	if n >= 2 {
		if hs.tanFinder != nil {
			hs.entryTansx, hs.entryTansy, hs.exitTansx, hs.exitTansy = hs.tanFinder.Find(hs.vertsx, hs.vertsy, hs.knots)
		}

		if hs.knots.IsUniform() {
			return hs.uniCanonical()
		} else {
			return hs.nonUniCanonical()
		}
	} else if n == 1 {
		// domain with value 0 only, knots '0,0'
		cubx := NewCubicPoly(hs.vertsx[0], 0, 0, 0)
		cuby := NewCubicPoly(hs.vertsy[0], 0, 0, 0)
		return NewCanonicalSpline2d([]Cubic2d{{cubx, cuby}}, bendit.NewKnots([]float64{0, 0}))
	} else {
		return NewCanonicalSpline2d([]Cubic2d{}, hs.knots)
	}
}

func (hs *HermiteSpline2d) uniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(tangents)
	segmCnt := len(hs.vertsx) - 1

	a := mat.NewDense(4, 4, []float64{
		1, 0, 0, 0,
		0, 0, 1, 0,
		-3, 3, -2, -1,
		2, -2, 1, 1,
	})

	bvs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		bvs = append(bvs, hs.vertsx[i], hs.vertsx[i+1], hs.exitTansx[i], hs.entryTansx[i+1])
		bvs = append(bvs, hs.vertsy[i], hs.vertsy[i+1], hs.exitTansy[i], hs.entryTansy[i+1])
	}
	b := mat.NewDense(dim*segmCnt, 4, bvs).T()

	var coefs mat.Dense
	coefs.Mul(a, b)

	cubics := make([]Cubic2d, segmCnt)
	colno := 0
	for i := 0; i < segmCnt; i++ {
		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, colno), coefs.At(1, colno), coefs.At(2, colno), coefs.At(3, colno)),
			NewCubicPoly(coefs.At(0, colno+1), coefs.At(1, colno+1), coefs.At(2, colno+1), coefs.At(3, colno+1)))
		colno += 2
	}
	return NewCanonicalSpline2d(cubics, hs.knots)
}

func (hs *HermiteSpline2d) nonUniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(entryTansx) == len(entryTansy) == len(exitTansx) == len(exitTansy) == len(knots)
	segmCnt := len(hs.vertsx) - 1
	cubics := make([]Cubic2d, segmCnt)

	for i := 0; i < segmCnt; i++ {
		//tlen := hs.knots[i+1] - hs.knots[i]
		tlen := hs.knots.SegmentLength(i)
		a := mat.NewDense(4, 4, []float64{
			1, 0, 0, 0,
			0, 0, tlen, 0,
			-3, 3, -2 * tlen, -tlen,
			2, -2, tlen, tlen,
		})

		b := mat.NewDense(4, dim, []float64{
			hs.vertsx[i], hs.vertsy[i],
			hs.vertsx[i+1], hs.vertsy[i+1],
			hs.exitTansx[i], hs.exitTansy[i],
			hs.entryTansx[i+1], hs.entryTansy[i+1],
		})

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, 0), coefs.At(1, 0), coefs.At(2, 0), coefs.At(3, 0)),
			NewCubicPoly(coefs.At(0, 1), coefs.At(1, 1), coefs.At(2, 1), coefs.At(3, 1)))
	}

	return NewCanonicalSpline2d(cubics, hs.knots)
}

func (hs *HermiteSpline2d) At(t float64) (x, y float64) {
	if hs.canon != nil {
		return hs.canon.At(t)
	} else {
		return 0, 0
	}
}

func (hs *HermiteSpline2d) Fn() bendit.Fn2d {
	if hs.canon != nil {
		return hs.canon.Fn()
	} else {
		return NewCanonicalSpline2d(nil, bendit.NewUniformKnots()).Fn()
	}
}

func (hs *HermiteSpline2d) Approx(maxDist float64, collector bendit.LineCollector2d) {
	panic("implement me")
}
