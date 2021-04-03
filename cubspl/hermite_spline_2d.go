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
	vertsx, vertsy []float64
	tangents       []VertexTan2d
	// TODO slopeEstimator
	knots []float64 // TODO uniform - non-uniform
}

/*
func NewHermiteSpline2d(vertsx, vertsy []float64, tangents []VertexTan2d, knots []float64) *HermiteSpline2d {
	n := len(vertsx)
	if len(vertsy) != n || len(tangents) != n || len(knots) != n {
		panic("versv, vertsy, tangents and knots must all have the same length")
	}
	return &HermiteSpline2d{vertsx: vertsx, vertsy: vertsy, tangents: tangents, knots: knots}
}
*/

func (hs *HermiteSpline2d) VertexCnt() int {
	return len(hs.vertsx)
}

func (hs *HermiteSpline2d) SegmentCnt() int {
	if len(hs.vertsx) > 0 {
		return len(hs.vertsx) - 1
	} else {
		return 0
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

func (hs *HermiteSpline2d) Add(vertx, verty float64, tangent VertexTan2d) {
	hs.vertsx = append(hs.vertsx, vertx)
	hs.vertsy = append(hs.vertsy, verty)
	hs.tangents = append(hs.tangents, tangent)
	hs.knots = append(hs.knots, hs.KnotN()+1) // TODO currently for uniform splines
}

func (hs *HermiteSpline2d) Fn() SplineFn2d {
	return BuildNuHermiteSplineFn2d(hs.vertsx, hs.vertsy, hs.tangents, hs.knots)
}

// build non-uniform hermite spline
func BuildNuHermiteSplineFn2d(vertsx, vertsy []float64, tangents []VertexTan2d, knots []float64) SplineFn2d {
	n := len(vertsx)
	if len(vertsy) != n || len(tangents) != n || len(knots) != n {
		panic("versv, vertsy, tangents and knots must all have the same length")
	}

	if n >= 2 {
		cubx, cuby := createNuCubics(vertsx, vertsy, tangents, knots)
		return func(t float64) (x, y float64) {
			segmNo, u, err := mapNuToSegm(t, knots)
			if err != nil {
				return 0, 0 // TODO or panic? or error?
			} else {
				return cubx[segmNo].At(u), cuby[segmNo].At(u)
			}
		}
	} else {
		return func(t float64) (x, y float64) {
			if n == 1 && t == knots[0] { // TODO delta around first knot
				return vertsx[0], vertsy[0]
			} else {
				return 0, 0
			}
		}
	}
}

// create cubics for non-uniform spline
func createNuCubics(vertsx, vertsy []float64, tangents []VertexTan2d, knots []float64) (cubx, cuby []cubic) {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(tangents) == len(knots)
	segmCnt := len(vertsx) - 1
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

		stat, endt := tangents[i], tangents[i+1]
		smx, smy := stat.ExitTan()
		elx, ely := endt.EntryTan()
		b := mat.NewDense(4, dim, []float64{
			vertsx[i], vertsy[i],
			vertsx[i+1], vertsy[i+1],
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

func mapNuToSegm(t float64, knots []float64) (segmNo int, u float64, err error) {
	segmCnt := len(knots) - 1
	if segmCnt < 1 {
		err = errors.New("at least one segment having 2 knots required")
		return
	}
	if t < knots[0] {
		err = fmt.Errorf("%v smaller than first knot %v", t, knots[0])
		return
	}

	// TODO speed up mapping
	for i := 0; i < segmCnt; i++ {
		if t <= knots[i+1] {
			return i, (t - knots[i]) / (knots[i+1] - knots[i]), nil
		}
	}
	err = fmt.Errorf("%v greater than upper limit %v", t, knots[segmCnt+1])
	return
}

// build non-uniform hermite spline
func BuildUniHermiteSplineFn2d(vertsx, vertsy []float64, tangents []VertexTan2d) SplineFn2d {
	n := len(vertsx)
	if len(vertsy) != n || len(tangents) != n {
		panic("versv, vertsy and tangents must all have the same length")
	}

	if n >= 2 {
		cubx, cuby := createUniCubics(vertsx, vertsy, tangents)
		return func(t float64) (x, y float64) {
			segmNo, u, err := mapUniToSegm(t, n-1)
			if err != nil {
				return 0, 0 // TODO or panic? or error?
			} else {
				return cubx[segmNo].At(u), cuby[segmNo].At(u)
			}
		}
	} else {
		return func(t float64) (x, y float64) {
			if n == 1 && t == 0 { // TODO delta around 0
				return vertsx[0], vertsy[0]
			} else {
				return 0, 0
			}
		}
	}
}

// create cubics for uniform spline
func createUniCubics(vertsx, vertsy []float64, tangents []VertexTan2d) (cubx, cuby []cubic) {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(tangents)
	segmCnt := len(vertsx) - 1
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
		stat, endt := tangents[i], tangents[i+1]
		smx, smy := stat.ExitTan()
		elx, ely := endt.EntryTan()
		bvs = append(bvs, vertsx[i], vertsx[i+1], smx, elx)
		bvs = append(bvs, vertsy[i], vertsy[i+1], smy, ely)
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

func mapUniToSegm(t float64, segmCnt int) (segmNo int, u float64, err error) {
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
