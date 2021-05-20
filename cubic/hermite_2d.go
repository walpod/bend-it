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

type HermiteVertex2d struct {
	x, y                 float64
	entryTanx, entryTany float64
	exitTanx, exitTany   float64
}

func NewHermiteVertex2d(x float64, y float64, entryTanx float64, entryTany float64, exitTanx float64, exitTany float64) *HermiteVertex2d {
	return &HermiteVertex2d{x: x, y: y, entryTanx: entryTanx, entryTany: entryTany, exitTanx: exitTanx, exitTany: exitTany}
}

// HermiteTanFinder2d finds tangents based on given vertices and knots
/*type HermiteTanFinder2d interface {
	Find(vertsx, vertsy []float64, knots *bendit.Knots) (
		entryTansx, entryTansy []float64, exitTansx, exitTansy []float64)
}*/

// HermiteTanFinder2d finds tangents based on given vertices and knots
type HermiteTanFinder2d interface {
	Find(knots *bendit.Knots, verts []*HermiteVertex2d)
}

type HermiteSpline2d struct {
	knots     *bendit.Knots
	verts     []*HermiteVertex2d
	tanFinder HermiteTanFinder2d
	/*vertsx, vertsy         []float64
	entryTansx, entryTansy []float64
	exitTansx, exitTansy   []float64*/
	canon *CanonicalSpline2d
}

func NewHermiteSpline2d(knots *bendit.Knots, verts ...*HermiteVertex2d) *HermiteSpline2d {
	herm := &HermiteSpline2d{knots: knots, verts: verts}
	herm.Build() // TODO don't build automatically
	return herm
}

func NewHermiteSplineTanFinder2d(knots *bendit.Knots, tanFinder HermiteTanFinder2d, verts ...*HermiteVertex2d) *HermiteSpline2d {
	herm := &HermiteSpline2d{knots: knots, verts: verts, tanFinder: tanFinder}
	herm.Build() // TODO don't build automatically
	return herm
}

/*
func NewHermiteSpline2d(knots *bendit.Knots, vertsx []float64, vertsy []float64,
	entryTansx []float64, entryTansy []float64, exitTansx []float64, exitTansy []float64) *HermiteSpline2d {

	herm := &HermiteSpline2d{vertsx: vertsx, vertsy: vertsy,
		entryTansx: entryTansx, entryTansy: entryTansy, exitTansx: exitTansx, exitTansy: exitTansy, knots: knots}
	herm.Build()
	return herm
}

func NewHermiteSplineTanFinder2d(knots *bendit.Knots, vertsx []float64, vertsy []float64, tanFinder HermiteTanFinder2d) *HermiteSpline2d {
	herm := &HermiteSpline2d{vertsx: vertsx, vertsy: vertsy, tanFinder: tanFinder, knots: knots}
	herm.Build()
	return herm
}
*/

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
	segmCnt := len(hs.verts) - 1
	if segmCnt >= 0 {
		return segmCnt
	} else {
		return 0
	}
}

func (hs *HermiteSpline2d) Knots() *bendit.Knots {
	return hs.knots
}

// Build hermite spline by mapping to canonical representation
func (hs *HermiteSpline2d) Build() {
	hs.canon = hs.Canonical()
}

func (hs *HermiteSpline2d) Canonical() *CanonicalSpline2d {
	n := len(hs.verts)
	if n >= 2 {
		if hs.tanFinder != nil {
			hs.tanFinder.Find(hs.knots, hs.verts)
		}

		if hs.knots.IsUniform() {
			return hs.uniCanonical()
		} else {
			return hs.nonUniCanonical()
		}
	} else if n == 1 {
		return NewOneVertexCanonicalSpline2d(hs.verts[0].x, hs.verts[0].y)
	} else {
		return NewCanonicalSpline2d(hs.knots)
	}
}

func (hs *HermiteSpline2d) uniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: segmCnt >= 1, bs.knots.IsUniform()
	segmCnt := hs.SegmentCnt()

	a := mat.NewDense(4, 4, []float64{
		1, 0, 0, 0,
		0, 0, 1, 0,
		-3, 3, -2, -1,
		2, -2, 1, 1,
	})

	bvs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := hs.verts[i], hs.verts[i+1]
		bvs = append(bvs, vstart.x, vend.x, vstart.exitTanx, vend.entryTanx)
		bvs = append(bvs, vstart.y, vend.y, vstart.exitTany, vend.entryTany)
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
	return NewCanonicalSpline2d(hs.knots, cubics...)
}

func (hs *HermiteSpline2d) nonUniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(entryTansx) == len(entryTansy) == len(exitTansx) == len(exitTansy) == len(knots)
	segmCnt := hs.SegmentCnt()
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

		vstart, vend := hs.verts[i], hs.verts[i+1]
		b := mat.NewDense(4, dim, []float64{
			vstart.x, vstart.y,
			vend.x, vend.y,
			vstart.exitTanx, vstart.exitTany,
			vend.exitTanx, vend.exitTany,
		})

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, 0), coefs.At(1, 0), coefs.At(2, 0), coefs.At(3, 0)),
			NewCubicPoly(coefs.At(0, 1), coefs.At(1, 1), coefs.At(2, 1), coefs.At(3, 1)))
	}

	return NewCanonicalSpline2d(hs.knots, cubics...)
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
		return NewCanonicalSpline2d(bendit.NewUniformKnots()).Fn()
	}
}

func (hs *HermiteSpline2d) Approx(maxDist float64, collector bendit.LineCollector2d) {
	panic("implement me")
}
