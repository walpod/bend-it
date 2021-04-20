package cubic

import (
	bendit "github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
	"math"
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
			cubics = bs.createUniCubics() // uniform
		} else {
			cubics = bs.createNonUniCubics() // non-uniform
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

// approximate bezier-spline with polygon using subdivision
func (bs *BezierSpline2d) Approximate(
	isFlat func(x0, y0, x1, y1, x2, y2, x3, y3 float64) bool,
	line func(x0, y0, x1, y1 float64)) {

	myIsFlat := func(x0, y0, x1, y1, x2, y2, x3, y3 float64) bool {
		if isFlat != nil {
			return isFlat(x0, y0, x1, y1, x2, y2, x3, y3)
		} else {
			const delta = 0.1 // TODO
			lx, ly := x3-x0, y3-y0
			return ProjectedVectorDist(x1-x0, y1-y0, lx, ly) <= delta && ProjectedVectorDist(x2-x0, y2-y0, lx, ly) <= delta
		}
	}

	var subdivide func(x0, y0, x1, y1, x2, y2, x3, y3 float64)
	subdivide = func(x0, y0, x1, y1, x2, y2, x3, y3 float64) {
		if myIsFlat(x0, y0, x1, y1, x2, y2, x3, y3) {
			line(x0, y0, x3, y3)
		} else {
			m := 0.5
			x01, y01 := m*x0+m*x1, m*y0+m*y1
			x11, y11 := m*x1+m*x2, m*y1+m*y2
			x21, y21 := m*x2+m*x3, m*y2+m*y3
			x02, y02 := m*x01+m*x11, m*y01+m*y11
			x12, y12 := m*x11+m*x21, m*y11+m*y21
			xm, ym := m*x02+m*x12, m*y02+m*y12
			subdivide(x0, y0, x01, y01, x02, y02, xm, ym)
			subdivide(xm, ym, x12, y12, x21, y21, x3, y3)
		}
	}

	for i := 0; i < len(bs.vertsx)-1; i++ {
		subdivide(bs.vertsx[i], bs.vertsy[i],
			bs.ctrlx[2*i], bs.ctrly[2*i],
			bs.ctrlx[2*i+1], bs.ctrly[2*i+1],
			bs.vertsx[i+1], bs.vertsy[i+1])
	}
}

// calculate distance of vector v to projected vector v on w
func ProjectedVectorDist(vx, vy, wx, wy float64) float64 {
	// distance = area of parallelogram(v, w) / length(w)
	return math.Abs(wx*vy-wy*vx) / math.Sqrt(wx*wx+wy*wy)
}
