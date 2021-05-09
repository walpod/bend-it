package cubic

import (
	bendit "github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
	"math"
)

type BezierSpline2d struct {
	vertsx, vertsy []float64
	ctrlx, ctrly   []float64
	knots          *bendit.Knots
	canon          *CanonicalSpline2d // map to canonical, cubic spline
}

func NewBezierSpline2d(vertsx []float64, vertsy []float64, ctrlx []float64, ctrly []float64, knots *bendit.Knots) *BezierSpline2d {
	n := len(vertsx)
	ctrlCnt := (n - 1) * 2
	if ctrlCnt < 0 {
		ctrlCnt = 0
	}
	if len(vertsy) != n || len(ctrlx) != ctrlCnt || len(ctrly) != ctrlCnt || (knots.Count() > 0 && knots.Count() != n) {
		panic("vertsv, vertsy and (optional) knots must have the same length. ctrlx and ctrly must have twice the length of vertsx minus 2")
	}
	bs := &BezierSpline2d{vertsx: vertsx, vertsy: vertsy, ctrlx: ctrlx, ctrly: ctrly, knots: knots}
	bs.Build()
	return bs
}

func (bs *BezierSpline2d) SegmentCnt() int {
	return len(bs.vertsx) - 1
}

func (bs *BezierSpline2d) Knots() *bendit.Knots {
	return bs.knots
}

func (bs *BezierSpline2d) Build() {
	bs.canon = bs.Canonical()
}

func (bs *BezierSpline2d) Canonical() *CanonicalSpline2d {
	n := len(bs.vertsx)
	if n >= 2 {
		if bs.knots.IsUniform() {
			return bs.uniCanonical()
		} else {
			return bs.nonUniCanonical()
		}
	} else if n == 1 {
		// domain with value 0 only, knots '0,0'
		cubx := NewCubicPoly(bs.vertsx[0], 0, 0, 0)
		cuby := NewCubicPoly(bs.vertsy[0], 0, 0, 0)
		return NewCanonicalSpline2d([]Cubic2d{{cubx, cuby}}, bendit.NewKnots([]float64{0, 0}))
	} else {
		return NewCanonicalSpline2d([]Cubic2d{}, bs.knots)
	}
}

func (bs *BezierSpline2d) uniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: len(vertsx) >= 2, len(vertsx) == len(vertsy) == len(tangents), bs.knots.IsUniform()
	segmCnt := bs.SegmentCnt()

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

	return NewCanonicalSpline2dByMatrix(coefs, bs.knots)
}

func (bs *BezierSpline2d) nonUniCanonical() *CanonicalSpline2d {
	// TODO
	panic("not yet implemented")
}

// At evaluates point on bezier spline for given parameter t
func (bs *BezierSpline2d) At(t float64) (x, y float64) {
	if bs.canon != nil {
		return bs.canon.At(t)
	} else {
		return 0, 0
	}
}

// AtDeCasteljau is an alternative to 'At' using De Casteljau algorithm.
// As opposed to At calling Build beforehand is not required
func (bs *BezierSpline2d) AtDeCasteljau(t float64) (x, y float64) {
	segmNo, u, err := bs.knots.MapToSegment(t, bs.SegmentCnt())
	if err != nil {
		return 0, 0
	} else {
		// TODO prepare u for non-uniform
		linip := func(a, b float64) float64 { // linear interpolation
			return a + u*(b-a)
		}
		x01, y01 := linip(bs.vertsx[segmNo], bs.ctrlx[2*segmNo]), linip(bs.vertsy[segmNo], bs.ctrly[2*segmNo])
		x11, y11 := linip(bs.ctrlx[2*segmNo], bs.ctrlx[2*segmNo+1]), linip(bs.ctrly[2*segmNo], bs.ctrly[2*segmNo+1])
		x21, y21 := linip(bs.ctrlx[2*segmNo+1], bs.vertsx[segmNo+1]), linip(bs.ctrly[2*segmNo+1], bs.vertsy[segmNo+1])
		x02, y02 := linip(x01, x11), linip(y01, y11)
		x12, y12 := linip(x11, x21), linip(y11, y21)
		return linip(x02, x12), linip(y02, y12)
	}
}

func (bs *BezierSpline2d) Fn() bendit.Fn2d {
	if bs.canon != nil {
		return bs.canon.Fn()
	} else {
		return NewCanonicalSpline2d(nil, bendit.NewUniformKnots()).Fn()
	}
}

// approximate bezier-spline with line-segments using subdivision
func (bs *BezierSpline2d) Approximate(maxDist float64, collector bendit.LineCollector2d) {
	isFlat := func(x0, y0, x1, y1, x2, y2, x3, y3 float64) bool {
		lx, ly := x3-x0, y3-y0
		return ProjectedVectorDist(x1-x0, y1-y0, lx, ly) <= maxDist &&
			ProjectedVectorDist(x2-x0, y2-y0, lx, ly) <= maxDist
	}

	var subdivide func(x0, y0, x1, y1, x2, y2, x3, y3 float64)
	subdivide = func(x0, y0, x1, y1, x2, y2, x3, y3 float64) {
		if isFlat(x0, y0, x1, y1, x2, y2, x3, y3) {
			collector.CollectLine(x0, y0, x3, y3)
		} else {
			m := 0.5
			x01, y01 := m*x0+m*x1, m*y0+m*y1
			x11, y11 := m*x1+m*x2, m*y1+m*y2
			x21, y21 := m*x2+m*x3, m*y2+m*y3
			x02, y02 := m*x01+m*x11, m*y01+m*y11
			x12, y12 := m*x11+m*x21, m*y11+m*y21
			x03, y03 := m*x02+m*x12, m*y02+m*y12
			subdivide(x0, y0, x01, y01, x02, y02, x03, y03)
			subdivide(x03, y03, x12, y12, x21, y21, x3, y3)
		}
	}

	// subdivide each segment
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

type DirectCollector2d struct {
	line func(x0, y0, x3, y3 float64)
}

func NewDirectCollector2d(line func(x0, y0, x3, y3 float64)) *DirectCollector2d {
	return &DirectCollector2d{line: line}
}

func (lc DirectCollector2d) CollectLine(x0, y0, x3, y3 float64) {
	lc.line(x0, y0, x3, y3)
}
