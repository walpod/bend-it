package cubic

import (
	"fmt"
	bendit "github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
	"math"
)

type BezierVx2 struct {
	x, y  float64
	entry Controller
	exit  Controller
}

func NewBezierVx2(x float64, y float64, entry Controller, exit Controller) *BezierVx2 {
	// exchange Controller
	switch entry.(type) {
	case *Reflective:
		if exit == nil || exit.IsForExchange() {
			panic("reflective entry must have concrete exit (!IsForExchange)")
		}
		entry = NewPointReflection(x, y, exit.ControlX(), exit.ControlY())
	}

	switch exit.(type) {
	case *Reflective:
		if entry == nil || entry.IsForExchange() {
			panic("reflective exit must have concrete entry (!IsForExchange)")
		}
		exit = NewPointReflection(x, y, entry.ControlX(), entry.ControlY())
	}

	return &BezierVx2{x: x, y: y, entry: entry, exit: exit}
}

/*func NewBezierVx2(x float64, y float64, entryCtrlx float64, entryCtrly float64, exitCtrlx float64, exitCtrly float64) *BezierVx2 {
	return &BezierVx2{x: x, y: y,
		entryCtrlx: entryCtrlx, entryCtrly: entryCtrly, exitCtrlx: exitCtrlx, exitCtrly: exitCtrly}
}*/

func (vx BezierVx2) Coord() (x, y float64) {
	return vx.x, vx.y
}

func (vx BezierVx2) Entry() Controller {
	return vx.entry
}

func (vx BezierVx2) Exit() Controller {
	return vx.exit
}

type BezierSpline2d struct {
	knots    bendit.Knots
	vertices []*BezierVx2
	annex    *bendit.Annex
	canon    *CanonicalSpline2d // map to canonical, cubic spline
}

func NewBezierSpline2d(tknots []float64, vertices ...*BezierVx2) *BezierSpline2d {
	var knots bendit.Knots
	if tknots == nil {
		knots = bendit.NewUniformKnots(len(vertices))
	} else {
		if len(tknots) != len(vertices) {
			panic("knots and vertices must have same length")
		}
		knots = bendit.NewNonUniformKnots(tknots)
	}

	bez := &BezierSpline2d{knots: knots, vertices: vertices, annex: bendit.NewAnnex(knots)}
	return bez
}

func NewBezierSpline2dByMatrix(tknots []float64, mat mat.Dense) *BezierSpline2d {
	const dim = 2
	rows, _ := mat.Dims()
	segmCnt := rows / 2
	vertices := make([]*BezierVx2, 0, segmCnt)
	vertices = append(vertices, NewBezierVx2(
		mat.At(0, 0), mat.At(1, 0),
		nil, //NewControl(0, 0,),
		NewControl(mat.At(0, 1), mat.At(1, 1))))
	for i := 1; i < segmCnt; i++ {
		vertices = append(vertices, NewBezierVx2(
			mat.At(i*dim, 0), mat.At(i*dim+1, 0),
			NewControl(mat.At(i*dim-2, 2), mat.At(i*dim-1, 2)),
			NewControl(mat.At(i*dim, 1), mat.At(i*dim+1, 1))))
	}
	vertices = append(vertices, NewBezierVx2(
		mat.At(segmCnt*dim-2, 3), mat.At(segmCnt*dim-1, 3),
		NewControl(mat.At(segmCnt*dim-2, 2), mat.At(segmCnt*dim-1, 2)),
		nil)) //NewControl(0, 0)

	return NewBezierSpline2d(tknots, vertices...)
}

func (sp *BezierSpline2d) Knots() bendit.Knots {
	return sp.knots
}

func (sp *BezierSpline2d) BezierVertex(knotNo int) *BezierVx2 {
	if knotNo >= len(sp.vertices) {
		return nil
	} else {
		return sp.vertices[knotNo]
	}
}

func (sp *BezierSpline2d) Vertex(knotNo int) bendit.Vertex2d {
	return sp.BezierVertex(knotNo)
}

func (sp *BezierSpline2d) Update(knotNo int, x float64, y float64, entry Controller, exit Controller) (err error) {
	if knotNo >= len(sp.vertices) {
		err = fmt.Errorf("knotNo %v does not exist", knotNo)
		return
	}
	sp.vertices[knotNo] = NewBezierVx2(x, y, entry, exit)
	return
}

func (sp *BezierSpline2d) Prepare() {
	sp.prepareCanon()
}

func (sp *BezierSpline2d) ResetPrepare() {
	sp.canon = nil
}

func (sp *BezierSpline2d) prepareCanon() {
	sp.canon = sp.Canonical()
}

func (sp *BezierSpline2d) Canonical() *CanonicalSpline2d {
	n := len(sp.vertices)
	if n >= 2 {
		if sp.knots.IsUniform() {
			return sp.uniCanonical()
		} else {
			return sp.nonUniCanonical()
		}
	} else if n == 1 {
		return NewSingleVxCanonicalSpline2d(sp.vertices[0].x, sp.vertices[0].y)
	} else {
		return NewCanonicalSpline2d(sp.knots.External())
	}
}

func (sp *BezierSpline2d) uniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: segmCnt >= 1, sp.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		avs = append(avs, vstart.x, vstart.exit.ControlX(), vend.entry.ControlX(), vend.x)
		avs = append(avs, vstart.y, vstart.exit.ControlY(), vend.entry.ControlY(), vend.y)
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

	return NewCanonicalSpline2dByMatrix(sp.knots.External(), coefs)
}

func (sp *BezierSpline2d) nonUniCanonical() *CanonicalSpline2d {
	// TODO implement non-uniform
	panic("not yet implemented")
}

// At evaluates point on bezier spline for given parameter t
func (sp *BezierSpline2d) At(t float64) (x, y float64) {
	if sp.canon == nil {
		sp.prepareCanon()
	}
	return sp.canon.At(t)
}

// AtDeCasteljau is an alternative to 'At' using De Casteljau algorithm.
func (sp *BezierSpline2d) AtDeCasteljau(t float64) (x, y float64) {
	segmNo, u, err := sp.knots.MapToSegment(t)
	if err != nil {
		return 0, 0
	} else {
		// TODO prepare u for non-uniform
		linip := func(a, b float64) float64 { // linear interpolation
			return a + u*(b-a)
		}
		start := sp.vertices[segmNo]
		end := sp.vertices[segmNo+1]
		x01, y01 := linip(start.x, start.exit.ControlX()), linip(start.y, start.exit.ControlY())
		x11, y11 := linip(start.exit.ControlX(), end.entry.ControlX()), linip(start.exit.ControlY(), end.entry.ControlY())
		x21, y21 := linip(end.entry.ControlX(), end.x), linip(end.entry.ControlY(), end.y)
		x02, y02 := linip(x01, x11), linip(y01, y11)
		x12, y12 := linip(x11, x21), linip(y11, y21)
		return linip(x02, x12), linip(y02, y12)
	}
}

func (sp *BezierSpline2d) Fn() bendit.Fn2d {
	if sp.canon == nil {
		sp.prepareCanon()
	}
	return sp.canon.Fn()
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
	for i := 0; i < sp.knots.SegmentCnt(); i++ {
		tstart, _ := sp.knots.Knot(i)
		tend, _ := sp.knots.Knot(i + 1)
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		subdivide(
			tstart, tend,
			vstart.x, vstart.y,
			vstart.exit.ControlX(), vstart.exit.ControlY(),
			vend.entry.ControlX(), vend.entry.ControlY(),
			vend.x, vend.y)
	}
}

// calculate distance of vector v to projected vector v on w
func ProjectedVectorDist(vx, vy, wx, wy float64) float64 {
	// distance = area of parallelogram(v, w) / length(w)
	return math.Abs(wx*vy-wy*vx) / math.Sqrt(wx*wx+wy*wy)
}

func (sp *BezierSpline2d) Annex() *bendit.Annex {
	return sp.annex
}
