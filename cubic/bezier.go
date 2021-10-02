package cubic

import (
	"fmt"
	bendit "github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

type BezierVx2 struct {
	v         bendit.Vec
	entry     bendit.Vec
	exit      bendit.Vec
	dependent bool // are the two controls dependent on each other?
}

// one of entry or exit control can be nil, is handled as dependent control (on other side of the vertex)
func NewBezierVx2(v, entry, exit bendit.Vec) *BezierVx2 {
	dependent := false

	// handle dependent controls
	if entry == nil && exit != nil {
		entry = v.InvertInPoint(exit)
		dependent = true
	} else if entry != nil && exit == nil {
		exit = v.InvertInPoint(entry)
		dependent = true
	}

	return &BezierVx2{v: v, entry: entry, exit: exit, dependent: dependent}
}

func (vt BezierVx2) Coord() bendit.Vec {
	return vt.v
}

func (vt BezierVx2) Entry() bendit.Vec {
	return vt.entry
}

func (vt BezierVx2) Exit() bendit.Vec {
	return vt.exit
}

func (vt BezierVx2) Control(isEntry bool) bendit.Vec {
	if isEntry {
		return vt.entry
	} else {
		return vt.exit
	}
}

func (vt BezierVx2) Dependent() bool {
	return vt.dependent
}

func (vt BezierVx2) Translate(d bendit.Vec) bendit.Vertex2d {
	var exit bendit.Vec
	if !vt.dependent {
		exit = vt.exit.Add(d)
	}
	return NewBezierVx2(vt.v.Add(d), vt.entry.Add(d), exit)
}

func (vt BezierVx2) WithEntry(entry bendit.Vec) *BezierVx2 {
	exit := vt.exit
	if vt.dependent {
		exit = nil
	}
	return NewBezierVx2(vt.v, entry, exit)
}

func (vt BezierVx2) WithExit(exit bendit.Vec) *BezierVx2 {
	entry := vt.entry
	if vt.dependent {
		entry = nil
	}
	return NewBezierVx2(vt.v, entry, exit)
}

type BezierSpline2d struct {
	knots    bendit.Knots
	vertices []*BezierVx2
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

	bez := &BezierSpline2d{knots: knots, vertices: vertices}
	return bez
}

func NewBezierSpline2dByMatrix(tknots []float64, dim int, mat mat.Dense) *BezierSpline2d {
	rows, _ := mat.Dims()
	segmCnt := rows / dim
	vertices := make([]*BezierVx2, 0, segmCnt)
	var v, entry, exit bendit.Vec

	// start vertex
	row := 0
	v = bendit.NewZeroVec(dim)
	exit = bendit.NewZeroVec(dim)
	for d := 0; d < dim; d, row = d+1, row+1 {
		v[d] = mat.At(row, 0)
		exit[d] = mat.At(row, 1)
	}
	vertices = append(vertices, NewBezierVx2(v, bendit.NewZeroVec(dim), exit))

	// intermediate vertices
	v = bendit.NewZeroVec(dim)
	entry = bendit.NewZeroVec(dim)
	exit = bendit.NewZeroVec(dim)
	for i := 1; i < segmCnt; i++ {
		for d := 0; d < dim; d, row = d+1, row+1 {
			v[d] = mat.At(row, 0)
			entry[d] = mat.At(row-dim, 2)
			exit[d] = mat.At(row, 1)
		}
		vertices = append(vertices, NewBezierVx2(v, entry, exit))
	}

	// end vertex
	row -= dim
	v = bendit.NewZeroVec(dim)
	entry = bendit.NewZeroVec(dim)
	for d := 0; d < dim; d, row = d+1, row+1 {
		v[d] = mat.At(row, 3)
		entry[d] = mat.At(row, 2)
	}
	vertices = append(vertices, NewBezierVx2(v, entry, bendit.NewZeroVec(dim)))
	/*
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
			nil)) //NewControl(0, 0)*/
	return NewBezierSpline2d(tknots, vertices...)
}

func (sp *BezierSpline2d) Knots() bendit.Knots {
	return sp.knots
}

func (sp *BezierSpline2d) Dim() int {
	if len(sp.vertices) == 0 {
		return 0
	} else {
		return sp.vertices[0].v.Dim()
	}
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

func (sp *BezierSpline2d) AddVertex(knotNo int, vertex bendit.Vertex2d) (err error) {
	err = sp.knots.AddKnot(knotNo)
	if err != nil {
		return err
	}
	bvt := vertex.(*BezierVx2)
	if knotNo == len(sp.vertices) {
		sp.vertices = append(sp.vertices, bvt)
	} else {
		sp.vertices = append(sp.vertices, nil)
		copy(sp.vertices[knotNo+1:], sp.vertices[knotNo:])
		sp.vertices[knotNo] = bvt
	}
	return nil
}

func (sp *BezierSpline2d) UpdateVertex(knotNo int, vertex bendit.Vertex2d) (err error) {
	if !sp.knots.KnotExists(knotNo) {
		return fmt.Errorf("knotNo %v does not exist", knotNo)
	}
	sp.vertices[knotNo] = vertex.(*BezierVx2)
	return nil
}

func (sp *BezierSpline2d) DeleteVertex(knotNo int) (err error) {
	err = sp.knots.DeleteKnot(knotNo)
	if err != nil {
		return err
	}
	if knotNo == len(sp.vertices)-1 {
		sp.vertices = sp.vertices[:knotNo]
	} else {
		sp.vertices = append(sp.vertices[:knotNo], sp.vertices[knotNo+1:]...)
	}
	return nil
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
		return NewSingleVertexCanonicalSpline2d(sp.vertices[0].v)
	} else {
		return NewCanonicalSpline2d(sp.knots.External())
	}
}

func (sp *BezierSpline2d) uniCanonical() *CanonicalSpline2d {
	// precondition: segmCnt >= 1, sp.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()
	dim := sp.Dim()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		for d := 0; d < dim; d++ {
			avs = append(avs, vstart.v[d], vstart.exit[d], vend.entry[d], vend.v[d])
		}
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

	return NewCanonicalSpline2dByMatrix(sp.knots.External(), dim, coefs)
}

func (sp *BezierSpline2d) nonUniCanonical() *CanonicalSpline2d {
	// TODO implement non-uniform
	panic("not yet implemented")
}

// At evaluates point on bezier spline for given parameter t
// Prepare must be called before
func (sp *BezierSpline2d) At(t float64) bendit.Vec {
	return sp.canon.At(t)
}

// AtDeCasteljau is an alternative to 'At' using De Casteljau algorithm.
func (sp *BezierSpline2d) AtDeCasteljau(t float64) bendit.Vec {
	segmNo, u, err := sp.knots.MapToSegment(t)
	if err != nil {
		return nil
	} else {
		dim := sp.Dim()
		// TODO prepare u for non-uniform
		linip := func(a, b float64) float64 { // linear interpolation
			return a + u*(b-a)
		}
		start := sp.vertices[segmNo]
		end := sp.vertices[segmNo+1]
		p := bendit.NewZeroVec(dim)
		for d := 0; d < dim; d++ {
			b01 := linip(start.v[d], start.exit[d])
			b11 := linip(start.exit[d], end.entry[d])
			b21 := linip(end.entry[d], end.v[d])
			b02 := linip(b01, b11)
			b12 := linip(b11, b21)
			p[d] = linip(b02, b12)
		}
		return p
		/*x01, y01 := linip(start.x, start.exit.X()), linip(start.y, start.exit.Y())
		x11, y11 := linip(start.exit.X(), end.entry.X()), linip(start.exit.Y(), end.entry.Y())
		x21, y21 := linip(end.entry.X(), end.x), linip(end.entry.Y(), end.y)
		x02, y02 := linip(x01, x11), linip(y01, y11)
		x12, y12 := linip(x11, x21), linip(y11, y21)
		return linip(x02, x12), linip(y02, y12)*/
	}
}

func (sp *BezierSpline2d) Fn() bendit.Fn2d {
	sp.prepareCanon()
	return sp.canon.Fn()
}

// Approx -imate bezier-spline with line-segments using subdivision
func (sp *BezierSpline2d) Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector bendit.LineCollector2d) {
	dim := sp.Dim()

	/*isFlat := func(x0, y0, x1, y1, x2, y2, x3, y3 float64) bool {
		lx, ly := x3-x0, y3-y0
		return ProjectedVectorDist(x1-x0, y1-y0, lx, ly) <= maxDist &&
			ProjectedVectorDist(x2-x0, y2-y0, lx, ly) <= maxDist
	}*/
	isFlat := func(v0, v1, v2, v3 bendit.Vec) bool {
		v03 := v3.Sub(v0)
		return v1.Sub(v0).ProjectedVecDist(v03) <= maxDist && v2.Sub(v0).ProjectedVecDist(v03) <= maxDist
	}

	/*var subdivide func(segmNo int, ts, te, x0, y0, x1, y1, x2, y2, x3, y3 float64)
	subdivide = func(segmNo int, ts, te, x0, y0, x1, y1, x2, y2, x3, y3 float64) {
		if isFlat(x0, y0, x1, y1, x2, y2, x3, y3) {
			collector.CollectLine(segmNo, ts, te, x0, y0, x3, y3)
		} else {
			m := 0.5
			tm := ts*m + te*m
			x01, y01 := m*x0+m*x1, m*y0+m*y1
			x11, y11 := m*x1+m*x2, m*y1+m*y2
			x21, y21 := m*x2+m*x3, m*y2+m*y3
			x02, y02 := m*x01+m*x11, m*y01+m*y11
			x12, y12 := m*x11+m*x21, m*y11+m*y21
			x03, y03 := m*x02+m*x12, m*y02+m*y12
			subdivide(segmNo, ts, tm, x0, y0, x01, y01, x02, y02, x03, y03)
			subdivide(segmNo, tm, te, x03, y03, x12, y12, x21, y21, x3, y3)
		}
	}*/
	var subdivide func(segmNo int, ts, te float64, v0, v1, v2, v3 bendit.Vec)
	subdivide = func(segmNo int, ts, te float64, v0, v1, v2, v3 bendit.Vec) {
		if isFlat(v0, v1, v2, v3) {
			collector.CollectLine(segmNo, ts, te, v0, v3)
		} else {
			m := 0.5
			tm := ts*m + te*m
			v01 := bendit.NewZeroVec(dim)
			v11 := bendit.NewZeroVec(dim)
			v21 := bendit.NewZeroVec(dim)
			v02 := bendit.NewZeroVec(dim)
			v12 := bendit.NewZeroVec(dim)
			v03 := bendit.NewZeroVec(dim)
			for d := 0; d < dim; d++ {
				v01[d] = m*v0[d] + m*v1[d]
				v11[d] = m*v1[d] + m*v2[d]
				v21[d] = m*v2[d] + m*v3[d]
				v02[d] = m*v01[d] + m*v11[d]
				v12[d] = m*v11[d] + m*v21[d]
				v03[d] = m*v02[d] + m*v12[d]
			}
			subdivide(segmNo, ts, tm, v0, v01, v02, v03)
			subdivide(segmNo, tm, te, v03, v12, v21, v3)
			/*x01, y01 := m*x0+m*x1, m*y0+m*y1
			x11, y11 := m*x1+m*x2, m*y1+m*y2
			x21, y21 := m*x2+m*x3, m*y2+m*y3
			x02, y02 := m*x01+m*x11, m*y01+m*y11
			x12, y12 := m*x11+m*x21, m*y11+m*y21
			x03, y03 := m*x02+m*x12, m*y02+m*y12
			subdivide(ts, tm, x0, y0, x01, y01, x02, y02, x03, y03)
			subdivide(tm, te, x03, y03, x12, y12, x21, y21, x3, y3)*/
		}
	}

	// subdivide each segment
	for segmentNo := fromSegmentNo; segmentNo <= toSegmentNo; segmentNo++ {
		tstart, _ := sp.knots.Knot(segmentNo)
		tend, _ := sp.knots.Knot(segmentNo + 1)
		vstart, vend := sp.vertices[segmentNo], sp.vertices[segmentNo+1]
		subdivide(segmentNo, tstart, tend, vstart.v, vstart.exit, vend.entry, vend.v)
	}
}

// calculate distance of vector v to projected vector v on w
/*func ProjectedVectorDist(vx, vy, wx, wy float64) float64 {
	// distance = area of parallelogram(v, w) / length(w)
	return math.Abs(wx*vy-wy*vx) / math.Sqrt(wx*wx+wy*wy)
}*/
