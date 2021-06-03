package cubic

import (
	"github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

type HermiteVx2 struct {
	x, y                 float64
	entryTanx, entryTany float64
	exitTanx, exitTany   float64
}

func NewHermiteVx2(x float64, y float64, entryTanx float64, entryTany float64, exitTanx float64, exitTany float64) *HermiteVx2 {
	return &HermiteVx2{x, y, entryTanx, entryTany, exitTanx, exitTany}
}

func NewHermiteVx2Raw(x float64, y float64) *HermiteVx2 {
	return NewHermiteVx2(x, y, 0, 0, 0, 0)
}

func (vx HermiteVx2) Coord() (x, y float64) {
	return vx.x, vx.y
}

func (vx HermiteVx2) EntryTan() (lx, ly float64) {
	return vx.entryTanx, vx.entryTany
}

func (vx HermiteVx2) ExitTan() (mx, my float64) {
	return vx.exitTanx, vx.exitTany
}

// HermiteTanFinder2d finds tangents based on given vertices and knots
type HermiteTanFinder2d interface {
	Find(knots bendit.Knots, vertices []*HermiteVx2)
}

type HermiteSpline2d struct {
	knots     bendit.Knots
	vertices  []*HermiteVx2
	tanFinder HermiteTanFinder2d
	canon     *CanonicalSpline2d
}

func NewHermiteSpline2d(tknots []float64, vertices ...*HermiteVx2) *HermiteSpline2d {
	return NewHermiteSplineTanFinder2d(tknots, nil, vertices...)
}

func NewHermiteSplineTanFinder2d(tknots []float64, tanFinder HermiteTanFinder2d, vertices ...*HermiteVx2) *HermiteSpline2d {
	var knots bendit.Knots
	if tknots == nil {
		knots = bendit.NewUniformKnots(len(vertices))
	} else {
		if len(tknots) != len(vertices) {
			panic("tknots and vertices must have same length")
		}
		knots = bendit.NewNonUniformKnots(tknots)
	}

	herm := &HermiteSpline2d{knots: knots, vertices: vertices, tanFinder: tanFinder}
	herm.Build() // TODO don't build automatically
	return herm
}

func (sp *HermiteSpline2d) SegmentCnt() int {
	segmCnt := len(sp.vertices) - 1
	if segmCnt >= 0 {
		return segmCnt
	} else {
		return 0
	}
}

func (sp *HermiteSpline2d) Knots() bendit.Knots {
	return sp.knots
}

func (sp *HermiteSpline2d) Add(vertex *HermiteVx2) {
	if sp.knots.IsUniform() {
		sp.knots.(*bendit.UniformKnots).Add(1)
	} else {
		sp.knots.(*bendit.NonUniformKnots).Add(1)
	}
	sp.vertices = append(sp.vertices, vertex)
}

func (sp *HermiteSpline2d) AddL(segmentLen float64, vertex *HermiteVx2) {
	if sp.knots.IsUniform() {
		err := sp.knots.(*bendit.UniformKnots).Add(1)
		if err != nil {
			panic(err.Error())
		}
	} else {
		sp.knots.(*bendit.NonUniformKnots).Add(segmentLen)
	}
	sp.vertices = append(sp.vertices, vertex)
}

// Build hermite spline by mapping to canonical representation
func (sp *HermiteSpline2d) Build() {
	sp.canon = sp.Canonical()
}

func (sp *HermiteSpline2d) Canonical() *CanonicalSpline2d {
	n := len(sp.vertices)
	if n >= 2 {
		if sp.tanFinder != nil {
			sp.tanFinder.Find(sp.knots, sp.vertices)
		}

		if sp.knots.IsUniform() {
			return sp.uniCanonical()
		} else {
			return sp.nonUniCanonical()
		}
	} else if n == 1 {
		return NewSingleVxCanonicalSpline2d(sp.vertices[0].x, sp.vertices[0].y)
	} else {
		return NewCanonicalSpline2d(sp.knots.External())
	}
}

func (sp *HermiteSpline2d) uniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: segmCnt >= 1, bs.knots.IsUniform()
	segmCnt := sp.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		avs = append(avs, vstart.x, vend.x, vstart.exitTanx, vend.entryTanx)
		avs = append(avs, vstart.y, vend.y, vstart.exitTany, vend.entryTany)
	}
	a := mat.NewDense(dim*segmCnt, 4, avs)

	b := mat.NewDense(4, 4, []float64{
		1, 0, -3, 2,
		0, 0, 3, -2,
		0, 1, -2, 1,
		0, 0, -1, 1,
	})

	var coefs mat.Dense
	coefs.Mul(a, b)

	return NewCanonicalSpline2dByMatrix(sp.knots.External(), coefs)
}

func (sp *HermiteSpline2d) nonUniCanonical() *CanonicalSpline2d {
	const dim = 2
	segmCnt := sp.SegmentCnt()
	cubics := make([]Cubic2d, segmCnt)

	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		a := mat.NewDense(dim, 4, []float64{
			vstart.x, vend.x, vstart.exitTanx, vend.entryTanx,
			vstart.y, vend.y, vstart.exitTany, vend.entryTany,
		})

		sgl, _ := sp.knots.SegmentLen(i)
		b := mat.NewDense(4, 4, []float64{
			1, 0, -3, 2,
			0, 0, 3, -2,
			0, sgl, -2 * sgl, sgl,
			0, 0, -sgl, sgl,
		})

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, 0), coefs.At(0, 1), coefs.At(0, 2), coefs.At(0, 3)),
			NewCubicPoly(coefs.At(1, 0), coefs.At(1, 1), coefs.At(1, 2), coefs.At(1, 3)))
	}

	return NewCanonicalSpline2d(sp.knots.External(), cubics...)
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
		// TODO implicit build? return NewCanonicalSpline2d(bendit.NewUniformKnots()).Fn()
		return nil
	}
}

func (sp *HermiteSpline2d) Bezier() *BezierSpline2d {
	n := len(sp.vertices)
	if n >= 2 {
		// TODO when to call Find ...
		if sp.tanFinder != nil {
			sp.tanFinder.Find(sp.knots, sp.vertices)
		}

		if sp.knots.IsUniform() {
			return sp.uniBezier()
		} else {
			panic("not yet implemented")
		}
	} else if n == 1 {
		return NewBezierSpline2d(sp.knots.External(),
			NewBezierVx2(sp.vertices[0].x, sp.vertices[0].y, 0, 0, 0, 0))
	} else {
		return NewBezierSpline2d(sp.knots.External())
	}
}

func (sp *HermiteSpline2d) uniBezier() *BezierSpline2d {
	const dim = 2
	// precondition: len(cubics) >= 1, bs.knots.IsUniform()
	segmCnt := sp.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		avs = append(avs, vstart.x, vend.x, vstart.exitTanx, vend.entryTanx)
		avs = append(avs, vstart.y, vend.y, vstart.exitTany, vend.entryTany)
	}
	a := mat.NewDense(dim*segmCnt, 4, avs)

	b := mat.NewDense(4, 4, []float64{
		1, 1, 0, 0,
		0, 0, 1, 1,
		0, 1. / 3, 0, 0,
		0, 0, -1. / 3, 0,
	})

	var coefs mat.Dense
	coefs.Mul(a, b)

	return NewBezierSpline2dByMatrix(sp.knots.External(), coefs)
}

func (sp *HermiteSpline2d) Approx(maxDist float64, collector bendit.LineCollector2d) {
	sp.Bezier().Approx(maxDist, collector)
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
*/
