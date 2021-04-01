package cubspl

import (
	"fmt"
	"gonum.org/v1/gonum/mat"
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
		cubx, cuby := hs.createCubics()
		return func(t float64) (x, y float64) {
			i, u, err := hs.mapToSegm(t)
			if err != nil {
				return 0, 0 // TODO or panic? or error?
			} else {
				return cubx[i].At(u), cuby[i].At(u)
			}
		}
	} else if n == 1 {
		return func(p float64) (x, y float64) {
			if p == 0 {
				x, y = hs.vertices[0].Point()
				return
			} else {
				return 0, 0
			}
		}
	} else if n == 0 {
		return func(p float64) (x, y float64) {
			return 0, 0
		}
	} else {
		panic("internal error: negative number of vertices")
	}
}

func (hs *HermiteSpline2d) createCubics() (cubx, cuby []cubic) {
	const dim = 2
	segmCnt := hs.SegmentCnt() // Precondition: segmCnt >= 1

	cubx = make([]cubic, segmCnt)
	cuby = make([]cubic, segmCnt)

	for i := 0; i < segmCnt; i++ {
		tlen := hs.knots[i+1] - hs.knots[i]
		a := mat.NewDense(4, 4, []float64{
			1, 0, 0, 0,
			0, 0, tlen, 0,
			-3, 3, -2 * tlen, -tlen,
			2, -2, tlen, tlen,
		})

		startv, endv := hs.vertices[i], hs.vertices[i+1]
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
func (hs *HermiteSpline2d) mapToSegm(t float64) (i int, u float64, err error) {
	segmCnt := hs.SegmentCnt() // Precondition: segmCnt >= 1
	if t < hs.knots[0] {
		err = fmt.Errorf("%v smaller than first knot %v", t, hs.knots[0])
		return
	}

	for i := 0; i < segmCnt; i++ {
		if t <= hs.knots[i+1] {
			return i, (t - hs.knots[i]) / (hs.knots[i+1] - hs.knots[i]), nil
		}
	}
	err = fmt.Errorf("%v greater than upper limit %v", t, hs.knots[segmCnt+1])
	return
}
