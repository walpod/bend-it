package cubic

import (
	bendit "github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

// cubic polynomial
type CubicPoly struct {
	a, b, c, d float64
}

func NewCubicPoly(a float64, b float64, c float64, d float64) CubicPoly {
	return CubicPoly{a: a, b: b, c: c, d: d}
}

func (cb *CubicPoly) At(u float64) float64 {
	return cb.a + u*(cb.b+u*(cb.c+cb.d*u))
}

func (cb *CubicPoly) Fn() func(float64) float64 {
	return func(u float64) float64 {
		return cb.At(u)
	}
}

type Cubic2d struct {
	// TODO maybe use instead 2x4 matrix and matrix multiplication
	cubx, cuby CubicPoly
}

func NewCubic2d(cubx CubicPoly, cuby CubicPoly) Cubic2d {
	return Cubic2d{cubx: cubx, cuby: cuby}
}

func (cb *Cubic2d) At(u float64) (x, y float64) {
	return cb.cubx.At(u), cb.cuby.At(u)
}

func (cb *Cubic2d) Fn() bendit.Fn2d {
	return func(u float64) (x, y float64) {
		return cb.At(u)
	}
}

type CanonicalSpline2d struct {
	cubics []Cubic2d
	knots  *bendit.Knots
}

func NewCanonicalSpline2d(knots *bendit.Knots, cubics ...Cubic2d) *CanonicalSpline2d {
	if knots.Count() > 0 && knots.Count() != len(cubics)+1 {
		panic("knots must be empty or having length of cubics + 1")
	}
	return &CanonicalSpline2d{cubics: cubics, knots: knots}
}

func NewOneVertexCanonicalSpline2d(x, y float64) *CanonicalSpline2d {
	// domain with value 0 only, knots '0,0'
	return NewCanonicalSpline2d(bendit.NewKnots([]float64{0, 0}), NewCubic2d(
		NewCubicPoly(x, 0, 0, 0),
		NewCubicPoly(y, 0, 0, 0)))
}

// matrix: (segmCnt*2) x 4
func NewCanonicalSpline2dByMatrix(mat mat.Dense, knots *bendit.Knots) *CanonicalSpline2d {
	r, _ := mat.Dims()
	segmCnt := r / 2
	if knots.Count() > 0 && knots.Count() != segmCnt+1 {
		panic("knots must be empty or having length or matrix-rows/2 + 1")
	}
	cubics := make([]Cubic2d, segmCnt)
	rowno := 0
	for i := 0; i < segmCnt; i++ {
		cubx := NewCubicPoly(mat.At(rowno, 0), mat.At(rowno, 1), mat.At(rowno, 2), mat.At(rowno, 3))
		rowno++
		cuby := NewCubicPoly(mat.At(rowno, 0), mat.At(rowno, 1), mat.At(rowno, 2), mat.At(rowno, 3))
		rowno++
		cubics[i] = NewCubic2d(cubx, cuby)
	}
	return &CanonicalSpline2d{cubics: cubics, knots: knots}
}

func (cs *CanonicalSpline2d) SegmentCnt() int {
	return len(cs.cubics)
}

func (cs *CanonicalSpline2d) Knots() *bendit.Knots {
	return cs.knots
}

func (cs *CanonicalSpline2d) At(t float64) (x, y float64) {
	if len(cs.cubics) == 0 {
		return 0, 0
	}

	segmNo, u, err := cs.knots.MapToSegment(t, cs.SegmentCnt())
	if err != nil {
		return 0, 0
	} else {
		return cs.cubics[segmNo].At(u)
	}
}

func (cs *CanonicalSpline2d) Fn() bendit.Fn2d {
	return func(t float64) (x, y float64) {
		return cs.At(t)
	}
}

func (cs *CanonicalSpline2d) Bezier() *BezierSpline2d {
	if len(cs.cubics) >= 1 {
		if cs.knots.IsUniform() {
			return cs.uniBezier()
		} else {
			panic("not yet implemented")
		}
	} else {
		return NewBezierSpline2d(bendit.NewUniformKnots())
	}
}

func (cs *CanonicalSpline2d) uniBezier() *BezierSpline2d {
	const dim = 2
	// precondition: len(verts) >= 2, bs.knots.IsUniform()
	segmCnt := cs.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		cubx := cs.cubics[i].cubx
		avs = append(avs, cubx.a, cubx.b, cubx.c, cubx.d)
		cuby := cs.cubics[i].cuby
		avs = append(avs, cuby.a, cuby.b, cuby.c, cuby.d)
	}
	a := mat.NewDense(dim*segmCnt, 4, avs)

	b := mat.NewDense(4, 4, []float64{
		1, 1, 1, 1,
		0, 1. / 3, 2. / 3, 1,
		0, 0, 1. / 3, 1,
		0, 0, 0, 1,
	})

	var coefs mat.Dense
	coefs.Mul(a, b)

	return NewBezierSpline2dByMatrix(cs.knots, coefs)
}

func (cs *CanonicalSpline2d) Approx(maxDist float64, collector bendit.LineCollector2d) {
	cs.Bezier().Approx(maxDist, collector)
}
