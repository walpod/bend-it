package cubic

import (
	"github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

type HermiteSpline2d struct {
	vertsx, vertsy         []float64
	tanFinder              HermiteTanFinder2d
	entryTansx, entryTansy []float64
	exitTansx, exitTansy   []float64
	knots                  []float64
	canon                  *CanonicalSpline2d
}

func NewHermiteSpline2d(vertsx []float64, vertsy []float64,
	entryTansx []float64, entryTansy []float64, exitTansx []float64, exitTansy []float64,
	knots []float64) *HermiteSpline2d {

	herm := &HermiteSpline2d{vertsx: vertsx, vertsy: vertsy,
		entryTansx: entryTansx, entryTansy: entryTansy, exitTansx: exitTansx, exitTansy: exitTansy, knots: knots}
	herm.Build()
	return herm
}

func NewHermiteSplineTanFinder2d(vertsx []float64, vertsy []float64, tanFinder HermiteTanFinder2d, knots []float64) *HermiteSpline2d {
	herm := &HermiteSpline2d{vertsx: vertsx, vertsy: vertsy, tanFinder: tanFinder, knots: knots}
	herm.Build()
	return herm
}

func (hs *HermiteSpline2d) SegmentCnt() int {
	return len(hs.vertsx) - 1
}

func (hs *HermiteSpline2d) Domain() bendit.SplineDomain {
	var to float64
	if hs.knots == nil {
		to = float64(hs.SegmentCnt())
	} else {
		to = hs.knots[len(hs.knots)-1]
	}
	return bendit.SplineDomain{From: 0, To: to}
}

// build hermite spline, if knots are empty then spline is uniform
func (hs *HermiteSpline2d) Build() {
	n := len(hs.vertsx)
	/*if len(vertsy) != n || len(entryTansx) != n || len(entryTansy) != n || len(exitTansx) != n || len(exitTansy) != n ||
		(len(knots) > 0 && len(knots) != n) {
		panic("versv, vertsy, all tangents and (optional) knots must have the same length")
	}*/
	if n >= 2 {
		if hs.tanFinder != nil {
			hs.entryTansx, hs.entryTansy, hs.exitTansx, hs.exitTansy = hs.tanFinder.Find(hs.vertsx, hs.vertsy, hs.knots)
		}

		var cubics []Cubic2d
		if len(hs.knots) == 0 {
			// uniform spline
			cubics = hs.createUniCubics()
		} else {
			// non-uniform spline
			cubics = hs.createNonUniCubics()
		}
		hs.canon = NewCanonicalSpline2d(cubics, hs.knots)
	} else {
		hs.canon = nil
	}
}

/*func (bs *HermiteSpline2d) BuildHermiteSpline2d() {
	n := len(bs.vertsx)

	if n >= 2 {
		var cubics []Cubic2d
		if len(bs.knots) == 0 {
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
}*/

// create cubics for uniform spline
func (hs *HermiteSpline2d) createUniCubics() []Cubic2d {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(tangents)
	segmCnt := len(hs.vertsx) - 1
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
		bvs = append(bvs, hs.vertsx[i], hs.vertsx[i+1], hs.exitTansx[i], hs.entryTansx[i+1])
		bvs = append(bvs, hs.vertsy[i], hs.vertsy[i+1], hs.exitTansy[i], hs.entryTansy[i+1])
	}
	b := mat.NewDense(dim*segmCnt, 4, bvs).T()

	var coefs mat.Dense
	coefs.Mul(a, b)

	cubics := make([]Cubic2d, segmCnt)

	colno := 0
	for i := 0; i < segmCnt; i++ {
		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, colno), coefs.At(1, colno), coefs.At(2, colno), coefs.At(3, colno)),
			NewCubicPoly(coefs.At(0, colno+1), coefs.At(1, colno+1), coefs.At(2, colno+1), coefs.At(3, colno+1)))
		colno += 2
	}
	return cubics
}

// create cubics for non-uniform spline
func (hs *HermiteSpline2d) createNonUniCubics() []Cubic2d {

	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(entryTansx) == len(entryTansy) == len(exitTansx) == len(exitTansy) == len(knots)
	segmCnt := len(hs.vertsx) - 1
	cubics := make([]Cubic2d, segmCnt)

	for i := 0; i < segmCnt; i++ {
		tlen := hs.knots[i+1] - hs.knots[i]
		a := mat.NewDense(4, 4, []float64{
			1, 0, 0, 0,
			0, 0, tlen, 0,
			-3, 3, -2 * tlen, -tlen,
			2, -2, tlen, tlen,
		})

		b := mat.NewDense(4, dim, []float64{
			hs.vertsx[i], hs.vertsy[i],
			hs.vertsx[i+1], hs.vertsy[i+1],
			hs.exitTansx[i], hs.exitTansy[i],
			hs.entryTansx[i+1], hs.entryTansy[i+1],
		})

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, 0), coefs.At(1, 0), coefs.At(2, 0), coefs.At(3, 0)),
			NewCubicPoly(coefs.At(0, 1), coefs.At(1, 1), coefs.At(2, 1), coefs.At(3, 1)))
	}

	return cubics
}

func (hs *HermiteSpline2d) At(t float64) (x, y float64) {
	if hs.canon != nil {
		return hs.canon.At(t)
	} else {
		return 0, 0
	}
}

func (hs *HermiteSpline2d) Fn() bendit.Fn2d {
	if hs.canon != nil {
		return hs.canon.Fn()
	} else {
		return NewCanonicalSpline2d(nil, nil).Fn()
	}
}
