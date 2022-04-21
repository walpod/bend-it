package cubic

import (
	"github.com/walpod/bendigo"
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

/*func (cb *CubicPoly) Fn() func(float64) float64 {
	return func(u float64) float64 {
		return cb.At(u)
	}
}*/

type CubicPolies struct {
	// TODO maybe use 2x4 matrix and matrix multiplication instead
	cubs []CubicPoly
}

func NewCubicPolies(cubs ...CubicPoly) CubicPolies {
	return CubicPolies{cubs: cubs}
}

func (cb CubicPolies) Dim() int {
	return len(cb.cubs)
}

func (cb *CubicPolies) At(u float64) bendigo.Vec {
	dim := len(cb.cubs)
	p := make(bendigo.Vec, dim)
	for d := 0; d < dim; d++ {
		p[d] = cb.cubs[d].At(u)
	}
	return p
}

type CanonicalSpline struct {
	knots  bendigo.Knots
	cubics []CubicPolies
}

// tknots: nil = uniform else non-uniform
func NewCanonicalSpline(tknots []float64, cubics ...CubicPolies) *CanonicalSpline {
	var knots bendigo.Knots
	if tknots == nil {
		// uniform
		knotCnt := len(cubics) + 1
		if len(cubics) == 0 {
			knotCnt = 0
		}
		knots = bendigo.NewUniformKnots(knotCnt)
	} else {
		// non-uniform
		if len(cubics) == 0 && len(tknots) != 0 {
			panic("knots must be empty if no cubics specified")
		}
		if len(cubics) > 0 && len(tknots) != len(cubics)+1 {
			panic("there must be one more knot than cubics")
		}
		knots = bendigo.NewNonUniformKnots(tknots)
	}

	canon := &CanonicalSpline{knots, cubics}
	return canon
}

func NewSingleVertexCanonicalSpline(v bendigo.Vec) *CanonicalSpline {
	// domain with value 0 only, knots '0,0'
	dim := len(v)
	cubs := make([]CubicPoly, dim)
	for d := 0; d < dim; d++ {
		cubs[d] = NewCubicPoly(v[d], 0, 0, 0)
	}
	return NewCanonicalSpline([]float64{0, 0}, NewCubicPolies(cubs...))
}

// matrix: (segmCnt*2) x 4
func NewCanonicalSplineByMatrix(tknots []float64, dim int, mat mat.Dense) *CanonicalSpline {
	r, _ := mat.Dims()
	segmCnt := r / dim
	if tknots != nil && len(tknots) != segmCnt+1 {
		panic("non-uniform knots must have length matrix-rows/dim + 1")
	}

	cubics := make([]CubicPolies, segmCnt)
	rowno := 0
	for i := 0; i < segmCnt; i++ {
		cubs := make([]CubicPoly, dim)
		for j := 0; j < dim; j++ {
			cubs[j] = NewCubicPoly(mat.At(rowno, 0), mat.At(rowno, 1), mat.At(rowno, 2), mat.At(rowno, 3))
			rowno++
		}
		cubics[i] = NewCubicPolies(cubs...)
	}
	return NewCanonicalSpline(tknots, cubics...)
}

func (sp *CanonicalSpline) Knots() bendigo.Knots {
	return sp.knots
}

func (sp *CanonicalSpline) At(t float64) bendigo.Vec {
	if len(sp.cubics) == 0 {
		return nil //return make(bendigo.Vector, sp.dim) ... point (0,0,...0)
	}

	segmentNo, u, err := sp.knots.MapToSegment(t)
	if err != nil {
		return nil
	} else {
		return sp.cubics[segmentNo].At(u)
	}
}
