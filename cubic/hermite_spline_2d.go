package cubic

import (
	"errors"
	"fmt"
	"gonum.org/v1/gonum/mat"
	"math"
)

// cubic polynomial
type cubicPoly struct {
	a, b, c, d float64
}

func (cb *cubicPoly) At(u float64) float64 {
	return cb.a + u*(cb.b+u*(cb.c+cb.d*u))
}

func (cb *cubicPoly) Fn() func(float64) float64 {
	return func(u float64) float64 {
		return cb.At(u)
	}
}

type Spline2d func(t float64) (x, y float64)

// build hermite spline, if knots are empty then spline is uniform
func BuildHermiteSpline2d(vertsx, vertsy []float64,
	entryTansx, entryTansy []float64,
	exitTansx, exitTansy []float64,
	knots []float64) Spline2d {

	n := len(vertsx)
	if len(vertsy) != n || len(entryTansx) != n || len(entryTansy) != n || len(exitTansx) != n || len(exitTansy) != n ||
		(len(knots) > 0 && len(knots) != n) {
		panic("versv, vertsy, all tangents and (optional) knots must have the same length")
	}

	if n >= 2 {
		if len(knots) == 0 {
			// uniform spline
			cubx, cuby := createUniCubics(vertsx, vertsy, entryTansx, entryTansy, exitTansx, exitTansy)
			return func(t float64) (x, y float64) {
				segmNo, u, err := mapUniToSegm(t, n-1)
				if err != nil {
					return 0, 0 // TODO or panic? or error?
				} else {
					return cubx[segmNo].At(u), cuby[segmNo].At(u)
				}
			}
		} else {
			// non-uniform spline
			cubx, cuby := createNuCubics(vertsx, vertsy, entryTansx, entryTansy, exitTansx, exitTansy, knots)
			return func(t float64) (x, y float64) {
				segmNo, u, err := mapNuToSegm(t, knots)
				if err != nil {
					return 0, 0 // TODO or panic? or error?
				} else {
					return cubx[segmNo].At(u), cuby[segmNo].At(u)
				}
			}
		}

	} else {
		return func(t float64) (x, y float64) {
			if n == 1 && ((len(knots) == 0 && t == 0) || (len(knots) == 1 && t == knots[0])) { // TODO delta around
				return vertsx[0], vertsy[0]
			} else {
				return 0, 0
			}
		}
	}
}

// create cubics for non-uniform spline
func createNuCubics(vertsx, vertsy []float64, entryTansx, entryTansy []float64, exitTansx, exitTansy []float64,
	knots []float64) (cubx, cuby []cubicPoly) {

	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(entryTansx) == len(entryTansy) == len(exitTansx) == len(exitTansy) == len(knots)
	segmCnt := len(vertsx) - 1
	cubx = make([]cubicPoly, segmCnt)
	cuby = make([]cubicPoly, segmCnt)

	for i := 0; i < segmCnt; i++ {
		tlen := knots[i+1] - knots[i]
		a := mat.NewDense(4, 4, []float64{
			1, 0, 0, 0,
			0, 0, tlen, 0,
			-3, 3, -2 * tlen, -tlen,
			2, -2, tlen, tlen,
		})

		b := mat.NewDense(4, dim, []float64{
			vertsx[i], vertsy[i],
			vertsx[i+1], vertsy[i+1],
			exitTansx[i], exitTansy[i],
			entryTansx[i+1], entryTansy[i+1],
		})

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubx[i] = cubicPoly{coefs.At(0, 0), coefs.At(1, 0), coefs.At(2, 0), coefs.At(3, 0)}
		cuby[i] = cubicPoly{coefs.At(0, 1), coefs.At(1, 1), coefs.At(2, 1), coefs.At(3, 1)}
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

// create cubics for uniform spline
func createUniCubics(vertsx, vertsy []float64, entryTansx, entryTansy []float64, exitTansx, exitTansy []float64) (cubx, cuby []cubicPoly) {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(tangents)
	segmCnt := len(vertsx) - 1
	if segmCnt < 1 {
		return []cubicPoly{}, []cubicPoly{}
	}

	a := mat.NewDense(4, 4, []float64{
		1, 0, 0, 0,
		0, 0, 1, 0,
		-3, 3, -2, -1,
		2, -2, 1, 1,
	})

	bvs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		bvs = append(bvs, vertsx[i], vertsx[i+1], exitTansx[i], entryTansx[i+1])
		bvs = append(bvs, vertsy[i], vertsy[i+1], exitTansy[i], entryTansy[i+1])
	}
	b := mat.NewDense(dim*segmCnt, 4, bvs).T()

	var coefs mat.Dense
	coefs.Mul(a, b)

	cubx = make([]cubicPoly, segmCnt)
	cuby = make([]cubicPoly, segmCnt)

	colno := 0
	for i := 0; i < segmCnt; i++ {
		cubx[i] = cubicPoly{coefs.At(0, colno), coefs.At(1, colno), coefs.At(2, colno), coefs.At(3, colno)}
		colno++
		cuby[i] = cubicPoly{coefs.At(0, colno), coefs.At(1, colno), coefs.At(2, colno), coefs.At(3, colno)}
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
