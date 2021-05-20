package cubic

import (
	"github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

type HermiteVertex2d struct {
	x, y                 float64
	entryTanx, entryTany float64
	exitTanx, exitTany   float64
}

func NewHermiteVertex2d(x float64, y float64, entryTanx float64, entryTany float64, exitTanx float64, exitTany float64) *HermiteVertex2d {
	return &HermiteVertex2d{x: x, y: y, entryTanx: entryTanx, entryTany: entryTany, exitTanx: exitTanx, exitTany: exitTany}
}

func NewRawHermiteVertex2d(x float64, y float64) *HermiteVertex2d {
	return NewHermiteVertex2d(x, y, 0, 0, 0, 0)
}

// HermiteTanFinder2d finds tangents based on given vertices and knots
type HermiteTanFinder2d interface {
	Find(knots *bendit.Knots, verts []*HermiteVertex2d)
}

type HermiteSpline2d struct {
	knots     *bendit.Knots
	verts     []*HermiteVertex2d
	tanFinder HermiteTanFinder2d
	canon     *CanonicalSpline2d
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

func (sp *HermiteSpline2d) SegmentCnt() int {
	segmCnt := len(sp.verts) - 1
	if segmCnt >= 0 {
		return segmCnt
	} else {
		return 0
	}
}

func (sp *HermiteSpline2d) Knots() *bendit.Knots {
	return sp.knots
}

// Build hermite spline by mapping to canonical representation
func (sp *HermiteSpline2d) Build() {
	sp.canon = sp.Canonical()
}

func (sp *HermiteSpline2d) Canonical() *CanonicalSpline2d {
	n := len(sp.verts)
	if n >= 2 {
		if sp.tanFinder != nil {
			sp.tanFinder.Find(sp.knots, sp.verts)
		}

		if sp.knots.IsUniform() {
			return sp.uniCanonical()
		} else {
			return sp.nonUniCanonical()
		}
	} else if n == 1 {
		return NewOneVertexCanonicalSpline2d(sp.verts[0].x, sp.verts[0].y)
	} else {
		return NewCanonicalSpline2d(sp.knots)
	}
}

func (sp *HermiteSpline2d) uniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: segmCnt >= 1, bs.knots.IsUniform()
	segmCnt := sp.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.verts[i], sp.verts[i+1]
		avs = append(avs, vstart.x, vend.x, vstart.exitTanx, vend.entryTanx)
		avs = append(avs, vstart.y, vend.y, vstart.exitTany, vend.entryTany)
	}
	a := mat.NewDense(dim*segmCnt, 4, avs)

	var b = mat.NewDense(4, 4, []float64{
		1, 0, -3, 2,
		0, 0, 3, -2,
		0, 1, -2, 1,
		0, 0, -1, 1,
	})

	var coefs mat.Dense
	coefs.Mul(a, b)

	return NewCanonicalSpline2dByMatrix(sp.knots, coefs)
}

func (sp *HermiteSpline2d) nonUniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(entryTansx) == len(entryTansy) == len(exitTansx) == len(exitTansy) == len(knots)
	segmCnt := sp.SegmentCnt()
	cubics := make([]Cubic2d, segmCnt)

	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.verts[i], sp.verts[i+1]
		a := mat.NewDense(dim, 4, []float64{
			vstart.x, vend.x, vstart.exitTanx, vend.entryTanx,
			vstart.y, vend.y, vstart.exitTany, vend.entryTany,
		})

		sgl := sp.knots.SegmentLength(i)
		b := mat.NewDense(4, 4, []float64{
			1, 0, -3, 2,
			0, 0, 3, -2,
			0, sgl, -2 * sgl, sgl,
			0, 0, -sgl, sgl,
		})

		/*a := coefs.NewDense(4, 4, []float64{
			1, 0, 0, 0,
			0, 0, sgl, 0,
			-3, 3, -2 * sgl, -sgl,
			2, -2, sgl, sgl,
		})

		vstart, vend := sp.verts[i], sp.verts[i+1]
		b := coefs.NewDense(4, dim, []float64{
			vstart.x, vstart.y,
			vend.x, vend.y,
			vstart.exitTanx, vstart.exitTany,
			vend.exitTanx, vend.exitTany,
		})*/

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, 0), coefs.At(0, 1), coefs.At(0, 2), coefs.At(0, 3)),
			NewCubicPoly(coefs.At(1, 0), coefs.At(1, 1), coefs.At(1, 2), coefs.At(1, 3)))
	}

	return NewCanonicalSpline2d(sp.knots, cubics...)
}

func (sp *HermiteSpline2d) At(t float64) (x, y float64) {
	if sp.canon != nil {
		return sp.canon.At(t)
	} else {
		return 0, 0
	}
}

func (sp *HermiteSpline2d) Fn() bendit.Fn2d {
	if sp.canon != nil {
		return sp.canon.Fn()
	} else {
		return NewCanonicalSpline2d(bendit.NewUniformKnots()).Fn()
	}
}

func (sp *HermiteSpline2d) Approx(maxDist float64, collector bendit.LineCollector2d) {
	panic("implement me")
}

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
