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
	knots  bendit.Knots
	cubics []Cubic2d
}

// tknots: nil = uniform else non-uniform
func NewCanonicalSpline2d(tknots []float64, cubics ...Cubic2d) *CanonicalSpline2d {
	var knots bendit.Knots
	if tknots == nil {
		knots = bendit.NewUniformKnots(len(cubics) + 1)
	} else {
		if len(cubics) == 0 && len(tknots) != 0 {
			panic("knots must be empty (not nil) if no cubics specified")
		}
		if len(cubics) > 0 && len(tknots) != len(cubics)+1 {
			panic("there must be one more knot than cubics")
		}
		knots = bendit.NewNonUniformKnots(tknots)
	}

	canon := &CanonicalSpline2d{knots, cubics}
	return canon
}

func NewSingleVxCanonicalSpline2d(x, y float64) *CanonicalSpline2d {
	// domain with value 0 only, knots '0,0'
	return NewCanonicalSpline2d([]float64{0, 0}, NewCubic2d(
		NewCubicPoly(x, 0, 0, 0),
		NewCubicPoly(y, 0, 0, 0)))
}

// matrix: (segmCnt*2) x 4
func NewCanonicalSpline2dByMatrix(tknots []float64, mat mat.Dense) *CanonicalSpline2d {
	r, _ := mat.Dims()
	segmCnt := r / 2
	if tknots != nil && len(tknots) != segmCnt+1 {
		panic("non-uniform knots must have length matrix-rows/2 + 1")
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
	return NewCanonicalSpline2d(tknots, cubics...)
}

func (sp *CanonicalSpline2d) Knots() bendit.Knots {
	return sp.knots
}

func (sp *CanonicalSpline2d) Vertex(knotNo int) bendit.Vertex2d {
	if knotNo > len(sp.cubics) {
		return nil
	} else if knotNo == len(sp.cubics) {
		x, y := sp.cubics[knotNo-1].At(1)
		return NewHermiteVx2Raw(x, y)
	} else {
		x, y := sp.cubics[knotNo].At(0)
		return NewHermiteVx2Raw(x, y)
	}
}

func (sp *CanonicalSpline2d) At(t float64) (x, y float64) {
	if len(sp.cubics) == 0 {
		return 0, 0
	}

	segmNo, u, err := sp.knots.MapToSegment(t)
	if err != nil {
		return 0, 0
	} else {
		return sp.cubics[segmNo].At(u)
	}
}

func (sp *CanonicalSpline2d) Fn() bendit.Fn2d {
	return func(t float64) (x, y float64) {
		return sp.At(t)
	}
}

func (sp *CanonicalSpline2d) Bezier() *BezierSpline2d {
	if len(sp.cubics) >= 1 {
		if sp.knots.IsUniform() {
			return sp.uniBezier()
		} else {
			panic("not yet implemented")
		}
	} else {
		return NewBezierSpline2d(nil)
	}
}

func (sp *CanonicalSpline2d) uniBezier() *BezierSpline2d {
	const dim = 2
	// precondition: len(cubics) >= 1, bs.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		cubx := sp.cubics[i].cubx
		avs = append(avs, cubx.a, cubx.b, cubx.c, cubx.d)
		cuby := sp.cubics[i].cuby
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

	return NewBezierSpline2dByMatrix(sp.knots.External(), coefs)
}

func (sp *CanonicalSpline2d) Approx(maxDist float64, collector bendit.LineCollector2d) {
	sp.Bezier().Approx(maxDist, collector)
}

func (sp *CanonicalSpline2d) ApproxSegments(fromSegmentNo, toSegmentNo int, maxDist float64, collector bendit.LineCollector2d) {
	sp.Bezier().ApproxSegments(fromSegmentNo, toSegmentNo, maxDist, collector)
}
