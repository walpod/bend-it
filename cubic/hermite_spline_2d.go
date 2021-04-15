package cubic

import (
	"github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

// build hermite spline, if knots are empty then spline is uniform
func BuildHermiteSpline2d(vertsx, vertsy []float64,
	entryTansx, entryTansy []float64,
	exitTansx, exitTansy []float64,
	knots []float64) bendit.Fn2d {

	n := len(vertsx)
	if len(vertsy) != n || len(entryTansx) != n || len(entryTansy) != n || len(exitTansx) != n || len(exitTansy) != n ||
		(len(knots) > 0 && len(knots) != n) {
		panic("versv, vertsy, all tangents and (optional) knots must have the same length")
	}

	if n >= 2 {
		var cubics []Cubic2d
		if len(knots) == 0 {
			// uniform spline
			cubics = createUniCubics(vertsx, vertsy, entryTansx, entryTansy, exitTansx, exitTansy)
		} else {
			// non-uniform spline
			cubics = createNonUniCubics(vertsx, vertsy, entryTansx, entryTansy, exitTansx, exitTansy, knots)
		}
		return NewCanonicalSpline2d(cubics, knots).Fn()
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

// create cubics for uniform spline
func createUniCubics(vertsx, vertsy []float64,
	entryTansx, entryTansy []float64, exitTansx, exitTansy []float64) (cubics []Cubic2d) {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(tangents)
	segmCnt := len(vertsx) - 1
	if segmCnt < 1 {
		return []Cubic2d{}
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

	cubics = make([]Cubic2d, segmCnt)

	colno := 0
	for i := 0; i < segmCnt; i++ {
		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, colno), coefs.At(1, colno), coefs.At(2, colno), coefs.At(3, colno)),
			NewCubicPoly(coefs.At(0, colno+1), coefs.At(1, colno+1), coefs.At(2, colno+1), coefs.At(3, colno+1)))
		colno += 2
	}
	return
}

// create cubics for non-uniform spline
func createNonUniCubics(vertsx, vertsy []float64,
	entryTansx, entryTansy []float64, exitTansx, exitTansy []float64,
	knots []float64) (cubics []Cubic2d) {

	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(entryTansx) == len(entryTansy) == len(exitTansx) == len(exitTansy) == len(knots)
	segmCnt := len(vertsx) - 1
	cubics = make([]Cubic2d, segmCnt)

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

		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, 0), coefs.At(1, 0), coefs.At(2, 0), coefs.At(3, 0)),
			NewCubicPoly(coefs.At(0, 1), coefs.At(1, 1), coefs.At(2, 1), coefs.At(3, 1)))
	}

	return
}
