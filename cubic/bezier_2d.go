package cubic

import (
	bendit "github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
	"math"
)

type BezierVertex2d struct {
	vertsx, vertsy         float64
	entryCtrlx, entryCtrly float64
	exitCtrlx, exitCtrly   float64
}

func NewBezierVertex2d(vertsx float64, vertsy float64, entryCtrlx float64, entryCtrly float64, exitCtrlx float64, exitCtrly float64) *BezierVertex2d {
	return &BezierVertex2d{vertsx: vertsx, vertsy: vertsy,
		entryCtrlx: entryCtrlx, entryCtrly: entryCtrly, exitCtrlx: exitCtrlx, exitCtrly: exitCtrly}
}

type BezierSpline2d struct {
	verts []*BezierVertex2d
	knots *bendit.Knots
	canon *CanonicalSpline2d // map to canonical, cubic spline
}

func NewBezierSpline2d(knots *bendit.Knots, verts ...*BezierVertex2d) *BezierSpline2d {
	if !knots.IsUniform() && len(verts) != knots.Count() {
		panic("verts and (optional) knots must have the same length")
	}
	bs := &BezierSpline2d{verts: verts, knots: knots}
	bs.Build()
	return bs
}

func (bs *BezierSpline2d) SegmentCnt() int {
	segmCnt := len(bs.verts) - 1
	if segmCnt >= 0 {
		return segmCnt
	} else {
		return 0
	}
}

func (bs *BezierSpline2d) Knots() *bendit.Knots {
	return bs.knots
}

func (bs *BezierSpline2d) Build() {
	bs.canon = bs.Canonical()
}

func (bs *BezierSpline2d) Canonical() *CanonicalSpline2d {
	n := len(bs.verts)
	if n >= 2 {
		if bs.knots.IsUniform() {
			return bs.uniCanonical()
		} else {
			return bs.nonUniCanonical()
		}
	} else if n == 1 {
		// domain with value 0 only, knots '0,0'
		cubx := NewCubicPoly(bs.verts[0].vertsx, 0, 0, 0)
		cuby := NewCubicPoly(bs.verts[0].vertsy, 0, 0, 0)
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
		start := bs.verts[i]
		end := bs.verts[i+1]
		avs = append(avs, start.vertsx, start.exitCtrlx, end.entryCtrlx, end.vertsx)
		avs = append(avs, start.vertsy, start.exitCtrly, end.entryCtrly, end.vertsy)
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
		start := bs.verts[segmNo]
		end := bs.verts[segmNo+1]
		x01, y01 := linip(start.vertsx, start.exitCtrlx), linip(start.vertsy, start.exitCtrly)
		x11, y11 := linip(start.exitCtrlx, end.entryCtrlx), linip(start.exitCtrly, end.entryCtrly)
		x21, y21 := linip(end.entryCtrlx, end.vertsx), linip(end.entryCtrly, end.vertsy)
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
func (bs *BezierSpline2d) Approx(maxDist float64, collector bendit.LineCollector2d) {
	isFlat := func(x0, y0, x1, y1, x2, y2, x3, y3 float64) bool {
		lx, ly := x3-x0, y3-y0
		return ProjectedVectorDist(x1-x0, y1-y0, lx, ly) <= maxDist &&
			ProjectedVectorDist(x2-x0, y2-y0, lx, ly) <= maxDist
	}

	var subdivide func(ts, te, x0, y0, x1, y1, x2, y2, x3, y3 float64)
	subdivide = func(ts, te, x0, y0, x1, y1, x2, y2, x3, y3 float64) {
		if isFlat(x0, y0, x1, y1, x2, y2, x3, y3) {
			collector.CollectLine(ts, te, x0, y0, x3, y3)
		} else {
			m := 0.5
			tm := ts*m + te*m
			x01, y01 := m*x0+m*x1, m*y0+m*y1
			x11, y11 := m*x1+m*x2, m*y1+m*y2
			x21, y21 := m*x2+m*x3, m*y2+m*y3
			x02, y02 := m*x01+m*x11, m*y01+m*y11
			x12, y12 := m*x11+m*x21, m*y11+m*y21
			x03, y03 := m*x02+m*x12, m*y02+m*y12
			subdivide(ts, tm, x0, y0, x01, y01, x02, y02, x03, y03)
			subdivide(tm, te, x03, y03, x12, y12, x21, y21, x3, y3)
		}
	}

	// subdivide each segment
	for i := 0; i < bs.SegmentCnt(); i++ {
		ts, te := bs.knots.SegmentRange(i)
		start := bs.verts[i]
		end := bs.verts[i+1]
		subdivide(
			ts, te,
			start.vertsx, start.vertsy,
			start.exitCtrlx, start.exitCtrly,
			end.entryCtrlx, end.entryCtrly,
			end.vertsx, end.vertsy)
	}
}

// calculate distance of vector v to projected vector v on w
func ProjectedVectorDist(vx, vy, wx, wy float64) float64 {
	// distance = area of parallelogram(v, w) / length(w)
	return math.Abs(wx*vy-wy*vx) / math.Sqrt(wx*wx+wy*wy)
}
