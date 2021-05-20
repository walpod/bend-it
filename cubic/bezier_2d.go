package cubic

import (
	bendit "github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
	"math"
)

type BezierVertex2d struct {
	x, y                   float64
	entryCtrlx, entryCtrly float64
	exitCtrlx, exitCtrly   float64
}

func NewBezierVertex2d(x float64, y float64, entryCtrlx float64, entryCtrly float64, exitCtrlx float64, exitCtrly float64) *BezierVertex2d {
	return &BezierVertex2d{x: x, y: y,
		entryCtrlx: entryCtrlx, entryCtrly: entryCtrly, exitCtrlx: exitCtrlx, exitCtrly: exitCtrly}
}

type BezierSpline2d struct {
	knots *bendit.Knots
	verts []*BezierVertex2d
	canon *CanonicalSpline2d // map to canonical, cubic spline
}

func NewBezierSpline2d(knots *bendit.Knots, verts ...*BezierVertex2d) *BezierSpline2d {
	if !knots.IsUniform() && len(verts) != knots.Count() {
		panic("verts and (optional) knots must have the same length")
	}
	bs := &BezierSpline2d{knots: knots, verts: verts}
	bs.Build() // TODO no automatic build
	return bs
}

func NewBezierSpline2dByMatrix(knots *bendit.Knots, mat mat.Dense) *BezierSpline2d {
	const dim = 2
	rows, _ := mat.Dims()
	segmCnt := rows / 2
	verts := make([]*BezierVertex2d, 0, segmCnt)
	verts = append(verts, NewBezierVertex2d(mat.At(0, 0), mat.At(1, 0),
		0, 0,
		mat.At(0, 1), mat.At(1, 1)))
	for i := 1; i < segmCnt; i++ {
		verts = append(verts, NewBezierVertex2d(mat.At(i*dim, 0), mat.At(i*dim+1, 0),
			mat.At(i*dim-2, 2), mat.At(i*dim-1, 2),
			mat.At(i*dim, 1), mat.At(i*dim+1, 1)))
	}
	verts = append(verts, NewBezierVertex2d(mat.At(segmCnt*dim-2, 3), mat.At(segmCnt*dim-1, 3),
		mat.At(segmCnt*dim-2, 2), mat.At(segmCnt*dim-1, 2),
		0, 0))
	return NewBezierSpline2d(knots, verts...)
}

func (sp *BezierSpline2d) SegmentCnt() int {
	segmCnt := len(sp.verts) - 1
	if segmCnt >= 0 {
		return segmCnt
	} else {
		return 0
	}
}

func (sp *BezierSpline2d) Knots() *bendit.Knots {
	return sp.knots
}

func (sp *BezierSpline2d) Build() {
	sp.canon = sp.Canonical()
}

func (sp *BezierSpline2d) Canonical() *CanonicalSpline2d {
	n := len(sp.verts)
	if n >= 2 {
		if sp.knots.IsUniform() {
			return sp.uniCanonical()
		} else {
			return sp.nonUniCanonical()
		}
	} else if n == 1 {
		return NewOneVertexCanonicalSpline2d(sp.verts[0].x, sp.verts[0].y)
	} else {
		return NewCanonicalSpline2d(sp.knots)
	}
}

func (sp *BezierSpline2d) uniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: segmCnt >= 1, sp.knots.IsUniform()
	segmCnt := sp.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.verts[i], sp.verts[i+1]
		avs = append(avs, vstart.x, vstart.exitCtrlx, vend.entryCtrlx, vend.x)
		avs = append(avs, vstart.y, vstart.exitCtrly, vend.entryCtrly, vend.y)
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

	return NewCanonicalSpline2dByMatrix(sp.knots, coefs)
}

func (sp *BezierSpline2d) nonUniCanonical() *CanonicalSpline2d {
	// TODO
	panic("not yet implemented")
}

// At evaluates point on bezier spline for given parameter t
func (sp *BezierSpline2d) At(t float64) (x, y float64) {
	if sp.canon != nil {
		return sp.canon.At(t)
	} else {
		return 0, 0
	}
}

// AtDeCasteljau is an alternative to 'At' using De Casteljau algorithm.
// As opposed to At calling Build beforehand is not required
func (sp *BezierSpline2d) AtDeCasteljau(t float64) (x, y float64) {
	segmNo, u, err := sp.knots.MapToSegment(t, sp.SegmentCnt())
	if err != nil {
		return 0, 0
	} else {
		// TODO prepare u for non-uniform
		linip := func(a, b float64) float64 { // linear interpolation
			return a + u*(b-a)
		}
		start := sp.verts[segmNo]
		end := sp.verts[segmNo+1]
		x01, y01 := linip(start.x, start.exitCtrlx), linip(start.y, start.exitCtrly)
		x11, y11 := linip(start.exitCtrlx, end.entryCtrlx), linip(start.exitCtrly, end.entryCtrly)
		x21, y21 := linip(end.entryCtrlx, end.x), linip(end.entryCtrly, end.y)
		x02, y02 := linip(x01, x11), linip(y01, y11)
		x12, y12 := linip(x11, x21), linip(y11, y21)
		return linip(x02, x12), linip(y02, y12)
	}
}

func (sp *BezierSpline2d) Fn() bendit.Fn2d {
	if sp.canon != nil {
		return sp.canon.Fn()
	} else {
		return NewCanonicalSpline2d(bendit.NewUniformKnots()).Fn()
	}
}

// Approx -imate bezier-spline with line-segments using subdivision
func (sp *BezierSpline2d) Approx(maxDist float64, collector bendit.LineCollector2d) {
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
	for i := 0; i < sp.SegmentCnt(); i++ {
		ts, te := sp.knots.SegmentRange(i)
		start := sp.verts[i]
		end := sp.verts[i+1]
		subdivide(
			ts, te,
			start.x, start.y,
			start.exitCtrlx, start.exitCtrly,
			end.entryCtrlx, end.entryCtrly,
			end.x, end.y)
	}
}

// calculate distance of vector v to projected vector v on w
func ProjectedVectorDist(vx, vy, wx, wy float64) float64 {
	// distance = area of parallelogram(v, w) / length(w)
	return math.Abs(wx*vy-wy*vx) / math.Sqrt(wx*wx+wy*wy)
}
