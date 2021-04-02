package cubspl

import (
	"errors"
	"fmt"
	"gonum.org/v1/gonum/mat"
	"math"
)

type cubic struct {
	a, b, c, d float64
}

func (cb *cubic) At(u float64) float64 {
	return cb.a + u*(cb.b+u*(cb.c+cb.d*u))
}

func (cb *cubic) Fn() func(float64) float64 {
	return func(u float64) float64 {
		return cb.At(u)
	}
}

type SplineFn2d func(t float64) (x, y float64)

type HermiteSpline2d struct {
	vertices []HermiteVertex2d
	knots    []float64
}

func NewHermiteSpline2d(vertices []HermiteVertex2d, knots []float64) *HermiteSpline2d {
	return &HermiteSpline2d{vertices: vertices, knots: knots}
}

func (hs *HermiteSpline2d) VertexCnt() int {
	return len(hs.vertices)
}

func (hs *HermiteSpline2d) SegmentCnt() int {
	sc := len(hs.vertices) - 1
	if sc < 0 {
		return 0
	} else {
		return sc
	}
}

func (hs *HermiteSpline2d) Knot0() float64 {
	if len(hs.knots) == 0 {
		return 0 // TODO
	} else {
		return hs.knots[0]
	}
}

func (hs *HermiteSpline2d) KnotN() float64 {
	lk := len(hs.knots)
	if lk == 0 {
		return -1 // TODO
	} else {
		return hs.knots[lk-1]
	}
}

func (hs *HermiteSpline2d) Add(v HermiteVertex2d) {
	hs.vertices = append(hs.vertices, v)
	hs.knots = append(hs.knots, hs.KnotN()+1) // TODO currently only uniform splines
}

func (hs *HermiteSpline2d) Fn() SplineFn2d {
	n := hs.VertexCnt()
	if n >= 2 {
		cubx, cuby := createCubicsNonUni(hs.vertices, hs.knots)
		return func(t float64) (x, y float64) {
			segmNo, u, err := mapToSegmNonUni(t, hs.knots)
			if err != nil {
				return 0, 0 // TODO or panic? or error?
			} else {
				return cubx[segmNo].At(u), cuby[segmNo].At(u)
			}
		}
	} else if n == 1 {
		return func(t float64) (x, y float64) {
			if t == hs.knots[0] {
				x, y = hs.vertices[0].Point()
				return
			} else {
				return 0, 0
			}
		}
	} else if n == 0 {
		return func(t float64) (x, y float64) {
			return 0, 0
		}
	} else {
		panic("internal error: negative number of vertices")
	}
}

// create cubics for non-uniform spline
func createCubicsNonUni(vertices []HermiteVertex2d, knots []float64) (cubx, cuby []cubic) {
	const dim = 2
	segmCnt := len(vertices) - 1
	cubx = make([]cubic, segmCnt)
	cuby = make([]cubic, segmCnt)

	for i := 0; i < segmCnt; i++ {
		tlen := knots[i+1] - knots[i]
		a := mat.NewDense(4, 4, []float64{
			1, 0, 0, 0,
			0, 0, tlen, 0,
			-3, 3, -2 * tlen, -tlen,
			2, -2, tlen, tlen,
		})

		startv, endv := vertices[i], vertices[i+1]
		spx, spy := startv.Point()
		epx, epy := endv.Point()
		smx, smy := startv.ExitTan()
		elx, ely := endv.EntryTan()
		b := mat.NewDense(4, dim, []float64{
			spx, spy,
			epx, epy,
			smx, smy,
			elx, ely,
		})

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubx[i] = cubic{coefs.At(0, 0), coefs.At(1, 0), coefs.At(2, 0), coefs.At(3, 0)}
		cuby[i] = cubic{coefs.At(0, 1), coefs.At(1, 1), coefs.At(2, 1), coefs.At(3, 1)}
	}

	return
}

// TODO speed up mapping
func mapToSegmNonUni(t float64, knots []float64) (segmNo int, u float64, err error) {
	segmCnt := len(knots) - 1
	if segmCnt < 1 {
		err = errors.New("at least one segment having 2 knots required")
		return
	}
	if t < knots[0] {
		err = fmt.Errorf("%v smaller than first knot %v", t, knots[0])
		return
	}

	for i := 0; i < segmCnt; i++ {
		if t <= knots[i+1] {
			return i, (t - knots[i]) / (knots[i+1] - knots[i]), nil
		}
	}
	err = fmt.Errorf("%v greater than upper limit %v", t, knots[segmCnt+1])
	return
}

// create cubics for uniform spline
func createCubicsUni(vertices []HermiteVertex2d) (cubx, cuby []cubic) {
	const dim = 2
	segmCnt := len(vertices) - 1
	if segmCnt < 1 {
		return []cubic{}, []cubic{}
	}

	a := mat.NewDense(4, 4, []float64{
		1, 0, 0, 0,
		0, 0, 1, 0,
		-3, 3, -2, -1,
		2, -2, 1, 1,
	})

	bvs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		startv, endv := vertices[i], vertices[i+1]
		spx, spy := startv.Point()
		epx, epy := endv.Point()
		smx, smy := startv.ExitTan()
		elx, ely := endv.EntryTan()
		bvs = append(bvs, spx, epx, smx, elx)
		bvs = append(bvs, spy, epy, smy, ely)
	}
	b := mat.NewDense(dim*segmCnt, 4, bvs).T()

	var coefs mat.Dense
	coefs.Mul(a, b)

	cubx = make([]cubic, segmCnt)
	cuby = make([]cubic, segmCnt)

	colno := 0
	for i := 0; i < segmCnt; i++ {
		cubx[i] = cubic{coefs.At(0, colno), coefs.At(1, colno), coefs.At(2, colno), coefs.At(3, colno)}
		colno++
		cuby[i] = cubic{coefs.At(0, colno), coefs.At(1, colno), coefs.At(2, colno), coefs.At(3, colno)}
		colno++
	}
	return
}

func mapToSegmUni(t float64, segmCnt int) (segmNo int, u float64, err error) {
	upper := float64(segmCnt)
	if t < 0 {
		err = fmt.Errorf("%v smaller than 0", t)
		return
	}
	if t > upper {
		err = fmt.Errorf("%v greater than last knot %v", t, upper)
		return
	}

	var ifl float64
	ifl, u = math.Modf(t)
	if ifl == upper {
		// special case t == upper
		segmNo = segmCnt - 1
		u = 1
	} else {
		segmNo = int(ifl)
	}
	return
}
