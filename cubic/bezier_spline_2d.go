package cubic

import (
	bendit "github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

type BezierSpline2d struct {
	vertsx, vertsy []float64
	ctrlx, ctrly   []float64
	knots          []float64
	//cubics         []Cubic2d
	canon *CanonicalSpline2d
}

func NewBezierSpline2d(vertsx []float64, vertsy []float64, ctrlx []float64, ctrly []float64, knots []float64) *BezierSpline2d {
	n := len(vertsx)
	ctrlCnt := (n - 1) * 2
	if len(vertsy) != n || len(ctrlx) != ctrlCnt || len(ctrly) != ctrlCnt || (len(knots) > 0 && len(knots) != n) {
		panic("vertsv, vertsy and (optional) knots must have the same length. ctrlx and ctrly must have twice the length of vertsx minus 2")
	}
	bs := &BezierSpline2d{vertsx: vertsx, vertsy: vertsy, ctrlx: ctrlx, ctrly: ctrly, knots: knots}
	bs.Build()
	return bs
}

func (bs *BezierSpline2d) SegmentCnt() int {
	return len(bs.vertsx) - 1
}

func (bs *BezierSpline2d) Domain() bendit.SplineDomain {
	var to float64
	if bs.knots == nil {
		to = float64(bs.SegmentCnt())
	} else {
		to = bs.knots[len(bs.knots)-1]
	}
	return bendit.SplineDomain{From: 0, To: to}
}

func (bs *BezierSpline2d) Build() {
	n := len(bs.vertsx)
	if n >= 2 {
		var cubics []Cubic2d
		if len(bs.knots) == 0 {
			// uniform spline
			cubics = bs.createUniCubics()
		} else {
			// non-uniform spline
			cubics = bs.createNonUniCubics()
		}
		bs.canon = NewCanonicalSpline2d(cubics, bs.knots)
	} else {
		bs.canon = nil
	}
}

func (bs *BezierSpline2d) createUniCubics() []Cubic2d {
	const dim = 2
	// precondition: len(vertsx) == len(vertsy) == len(tangents)
	segmCnt := bs.SegmentCnt()
	if segmCnt < 1 {
		return []Cubic2d{}
	}

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		avs = append(avs, bs.vertsx[i], bs.ctrlx[2*i], bs.ctrlx[2*i+1], bs.vertsx[i+1])
		avs = append(avs, bs.vertsy[i], bs.ctrly[2*i], bs.ctrly[2*i+1], bs.vertsy[i+1])
	}
	a := mat.NewDense(dim*segmCnt, 4, avs)

	var b = mat.NewDense(4, 4, []float64{
		1, -3, 3, -1,
		0, 3, -6, 3,
		0, 0, 3, -3,
		0, 0, 0, 1,
	})

	var coefs mat.Dense
	coefs.Mul(a, b)

	cubics := make([]Cubic2d, segmCnt)

	rowno := 0
	for i := 0; i < segmCnt; i++ {
		cubx := NewCubicPoly(coefs.At(rowno, 0), coefs.At(rowno, 1), coefs.At(rowno, 2), coefs.At(rowno, 3))
		rowno++
		cuby := NewCubicPoly(coefs.At(rowno, 0), coefs.At(rowno, 1), coefs.At(rowno, 2), coefs.At(rowno, 3))
		rowno++
		cubics[i] = NewCubic2d(cubx, cuby)
	}

	return cubics
}

func (bs *BezierSpline2d) createNonUniCubics() []Cubic2d {
	// TODO
	panic("not yet implemented")
}

func (bs *BezierSpline2d) At(t float64) (x, y float64) {
	if bs.canon != nil {
		return bs.canon.At(t)
	} else {
		return 0, 0
	}
}

func (bs *BezierSpline2d) Fn() bendit.Fn2d {
	if bs.canon != nil {
		return bs.canon.Fn()
	} else {
		return NewCanonicalSpline2d(nil, nil).Fn()
	}
}
