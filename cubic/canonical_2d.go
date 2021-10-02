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
	// TODO maybe use 2x4 matrix and matrix multiplication instead
	cubs []CubicPoly
}

func NewCubic2d(cubs ...CubicPoly) Cubic2d {
	return Cubic2d{cubs: cubs}
}

func (cb Cubic2d) Dim() int {
	return len(cb.cubs)
}

func (cb *Cubic2d) At(u float64) bendit.Vec {
	dim := len(cb.cubs)
	p := make(bendit.Vec, dim)
	for d := 0; d < dim; d++ {
		p[d] = cb.cubs[d].At(u)
	}
	return p
}

func (cb *Cubic2d) Fn() bendit.Fn2d {
	return func(u float64) bendit.Vec {
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

func NewSingleVertexCanonicalSpline2d(v bendit.Vec) *CanonicalSpline2d {
	// domain with value 0 only, knots '0,0'
	dim := len(v)
	cubs := make([]CubicPoly, dim)
	for d := 0; d < dim; d++ {
		cubs[d] = NewCubicPoly(v[d], 0, 0, 0)
	}
	return NewCanonicalSpline2d([]float64{0, 0}, NewCubic2d(cubs...))
}

// matrix: (segmCnt*2) x 4
func NewCanonicalSpline2dByMatrix(tknots []float64, dim int, mat mat.Dense) *CanonicalSpline2d {
	r, _ := mat.Dims()
	segmCnt := r / dim
	if tknots != nil && len(tknots) != segmCnt+1 {
		panic("non-uniform knots must have length matrix-rows/dim + 1")
	}

	cubics := make([]Cubic2d, segmCnt)
	rowno := 0
	for i := 0; i < segmCnt; i++ {
		cubs := make([]CubicPoly, dim)
		for j := 0; j < dim; j++ {
			cubs[j] = NewCubicPoly(mat.At(rowno, 0), mat.At(rowno, 1), mat.At(rowno, 2), mat.At(rowno, 3))
			rowno++
		}
		cubics[i] = NewCubic2d(cubs...)
	}
	return NewCanonicalSpline2d(tknots, cubics...)
}

func (sp *CanonicalSpline2d) Knots() bendit.Knots {
	return sp.knots
}

func (sp *CanonicalSpline2d) At(t float64) bendit.Vec {
	if len(sp.cubics) == 0 {
		return nil //return make(bendit.Vector, sp.dim) ... point (0,0,...0)
	}

	segmNo, u, err := sp.knots.MapToSegment(t)
	if err != nil {
		return nil
	} else {
		return sp.cubics[segmNo].At(u)
	}
}

func (sp *CanonicalSpline2d) Fn() bendit.Fn2d {
	return func(t float64) bendit.Vec {
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
	// precondition: len(cubics) >= 1, bs.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()
	dim := sp.cubics[0].Dim()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		for j := 0; j < dim; j++ {
			cub := sp.cubics[i].cubs[j]
			avs = append(avs, cub.a, cub.b, cub.c, cub.d)
		}
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

	return NewBezierSpline2dByMatrix(sp.knots.External(), dim, coefs)
}

func (sp *CanonicalSpline2d) Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector bendit.LineCollector2d) {
	// TODO Prepare bezier?
	sp.Bezier().Approx(fromSegmentNo, toSegmentNo, maxDist, collector)
}
