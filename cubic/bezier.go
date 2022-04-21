package cubic

import (
	"fmt"
	"github.com/walpod/bendigo"
	"gonum.org/v1/gonum/mat"
)

type BezierVertBuilder struct {
	knots    bendigo.Knots
	vertices []*BezierVertex
}

func NewBezierVertBuilder(tknots []float64, vertices ...*BezierVertex) *BezierVertBuilder {
	var knots bendigo.Knots
	if tknots == nil {
		knots = bendigo.NewUniformKnots(len(vertices))
	} else {
		if len(tknots) != len(vertices) {
			panic("knots and vertices must have same length")
		}
		knots = bendigo.NewNonUniformKnots(tknots)
	}

	bez := &BezierVertBuilder{knots: knots, vertices: vertices}
	return bez
}

func NewBezierVertBuilderByMatrix(tknots []float64, dim int, mat mat.Dense) *BezierVertBuilder {
	rows, _ := mat.Dims()
	segmCnt := rows / dim
	vertices := make([]*BezierVertex, 0, segmCnt)
	var v, entry, exit bendigo.Vec

	// start vertex
	row := 0
	v = bendigo.NewZeroVec(dim)
	exit = bendigo.NewZeroVec(dim)
	for d := 0; d < dim; d, row = d+1, row+1 {
		v[d] = mat.At(row, 0)
		exit[d] = mat.At(row, 1)
	}
	vertices = append(vertices, NewBezierVertex(v, bendigo.NewZeroVec(dim), exit))

	// intermediate vertices
	v = bendigo.NewZeroVec(dim)
	entry = bendigo.NewZeroVec(dim)
	exit = bendigo.NewZeroVec(dim)
	for i := 1; i < segmCnt; i++ {
		for d := 0; d < dim; d, row = d+1, row+1 {
			v[d] = mat.At(row, 0)
			entry[d] = mat.At(row-dim, 2)
			exit[d] = mat.At(row, 1)
		}
		vertices = append(vertices, NewBezierVertex(v, entry, exit))
	}

	// end vertex
	row -= dim
	v = bendigo.NewZeroVec(dim)
	entry = bendigo.NewZeroVec(dim)
	for d := 0; d < dim; d, row = d+1, row+1 {
		v[d] = mat.At(row, 3)
		entry[d] = mat.At(row, 2)
	}
	vertices = append(vertices, NewBezierVertex(v, entry, bendigo.NewZeroVec(dim)))

	return NewBezierVertBuilder(tknots, vertices...)
}

func (sp *BezierVertBuilder) Knots() bendigo.Knots {
	return sp.knots
}

func (sp *BezierVertBuilder) Dim() int {
	if len(sp.vertices) == 0 {
		return 0
	} else {
		return sp.vertices[0].loc.Dim()
	}
}

func (sp *BezierVertBuilder) BezierVertex(knotNo int) *BezierVertex {
	if knotNo >= len(sp.vertices) {
		return nil
	} else {
		return sp.vertices[knotNo]
	}
}

func (sp *BezierVertBuilder) Vertex(knotNo int) bendigo.Vertex {
	return sp.BezierVertex(knotNo)
}

func (sp *BezierVertBuilder) AddVertex(knotNo int, vertex bendigo.Vertex) (err error) {
	err = sp.knots.AddKnot(knotNo)
	if err != nil {
		return err
	}
	bvt := vertex.(*BezierVertex)
	if knotNo == len(sp.vertices) {
		sp.vertices = append(sp.vertices, bvt)
	} else {
		sp.vertices = append(sp.vertices, nil)
		copy(sp.vertices[knotNo+1:], sp.vertices[knotNo:])
		sp.vertices[knotNo] = bvt
	}
	return nil
}

func (sp *BezierVertBuilder) UpdateVertex(knotNo int, vertex bendigo.Vertex) (err error) {
	if !sp.knots.KnotExists(knotNo) {
		return fmt.Errorf("knotNo %v does not exist", knotNo)
	}
	sp.vertices[knotNo] = vertex.(*BezierVertex)
	return nil
}

func (sp *BezierVertBuilder) DeleteVertex(knotNo int) (err error) {
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

func (sp *BezierVertBuilder) Canonical() *CanonicalSpline {
	n := len(sp.vertices)
	if n >= 2 {
		if sp.knots.IsUniform() {
			return sp.uniCanonical()
		} else {
			return sp.nonUniCanonical()
		}
	} else if n == 1 {
		return NewSingleVertexCanonicalSpline(sp.vertices[0].loc)
	} else {
		return NewCanonicalSpline(sp.knots.External())
	}
}

func (sp *BezierVertBuilder) uniCanonical() *CanonicalSpline {
	// precondition: segmCnt >= 1, sp.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()
	dim := sp.Dim()

	avs := make([]float64, 0, segmCnt*dim*4)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		for d := 0; d < dim; d++ {
			avs = append(avs, vstart.loc[d], vstart.exit[d], vend.entry[d], vend.loc[d])
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

	return NewCanonicalSplineByMatrix(sp.knots.External(), dim, coefs)
}

func (sp *BezierVertBuilder) nonUniCanonical() *CanonicalSpline {
	// TODO implement non-uniform
	panic("not yet implemented")
}

func (sp *BezierVertBuilder) Build() bendigo.Spline {
	return sp.Canonical()
}

func (sp *BezierVertBuilder) DeCasteljauSpline() *DeCasteljauSpline {
	segmentCnt := sp.knots.SegmentCnt()
	if segmentCnt == 0 {
		return nil
	}

	controls := make([]bendigo.Vec, 0, segmentCnt*4)
	for s := 0; s < segmentCnt; s++ {
		vtstart, vtend := sp.vertices[s], sp.vertices[s+1]
		controls = append(controls, vtstart.loc, vtstart.exit, vtend.entry, vtend.loc)
	}

	return NewDeCasteljauSpline(sp.knots, controls)
}

func (sp *BezierVertBuilder) Linax(fromSegmentNo, toSegmentNo int, collector bendigo.LineCollector, linaxParams *bendigo.LinaxParams) {
	dim := sp.Dim()

	isFlat := func(v0, v1, v2, v3 bendigo.Vec) bool {
		v03 := v3.Sub(v0)
		return v1.Sub(v0).ProjectedVecDist(v03) <= linaxParams.MaxDist && v2.Sub(v0).ProjectedVecDist(v03) <= linaxParams.MaxDist
	}

	var subdivide func(segmentNo int, ts, te float64, v0, v1, v2, v3 bendigo.Vec)
	subdivide = func(segmentNo int, ts, te float64, v0, v1, v2, v3 bendigo.Vec) {
		if isFlat(v0, v1, v2, v3) {
			collector.CollectLine(segmentNo, ts, te, v0, v3)
		} else {
			m := 0.5
			tm := ts*m + te*m
			v01 := bendigo.NewZeroVec(dim)
			v11 := bendigo.NewZeroVec(dim)
			v21 := bendigo.NewZeroVec(dim)
			v02 := bendigo.NewZeroVec(dim)
			v12 := bendigo.NewZeroVec(dim)
			v03 := bendigo.NewZeroVec(dim)
			for d := 0; d < dim; d++ {
				v01[d] = m*v0[d] + m*v1[d]
				v11[d] = m*v1[d] + m*v2[d]
				v21[d] = m*v2[d] + m*v3[d]
				v02[d] = m*v01[d] + m*v11[d]
				v12[d] = m*v11[d] + m*v21[d]
				v03[d] = m*v02[d] + m*v12[d]
			}
			subdivide(segmentNo, ts, tm, v0, v01, v02, v03)
			subdivide(segmentNo, tm, te, v03, v12, v21, v3)
		}
	}

	// subdivide each segment
	for segmentNo := fromSegmentNo; segmentNo <= toSegmentNo; segmentNo++ {
		tstart, tend, err := bendigo.SegmentTrange(sp.knots, segmentNo)
		if err == nil { // ignore nonexistent segments
			vtstart, vtend := sp.vertices[segmentNo], sp.vertices[segmentNo+1]
			subdivide(segmentNo, tstart, tend, vtstart.loc, vtstart.exit, vtend.entry, vtend.loc)
		}
	}
}

// DeCasteljauSpline is an alternative Bezier implementation using De Casteljau algorithm.
type DeCasteljauSpline struct {
	knots    bendigo.Knots
	controls []bendigo.Vec // bezier controls, 4 per segment in consecutive order
}

func NewDeCasteljauSpline(knots bendigo.Knots, controls []bendigo.Vec) *DeCasteljauSpline {
	return &DeCasteljauSpline{knots: knots, controls: controls}
}

func (sp DeCasteljauSpline) Knots() bendigo.Knots {
	return sp.knots
}

func (sp DeCasteljauSpline) At(t float64) bendigo.Vec {
	segmentNo, u, err := sp.knots.MapToSegment(t)
	if err != nil {
		return nil
	}

	dim := 0
	if len(sp.controls) >= 1 {
		dim = sp.controls[0].Dim()
	}

	// TODO prepare u for non-uniform
	linip := func(a, b float64) float64 { // linear interpolation
		return a + u*(b-a)
	}
	idx := segmentNo * 4
	start, exit, entry, end := sp.controls[idx], sp.controls[idx+1], sp.controls[idx+2], sp.controls[idx+3]
	p := bendigo.NewZeroVec(dim)
	for d := 0; d < dim; d++ {
		b01 := linip(start[d], exit[d])
		b11 := linip(exit[d], entry[d])
		b21 := linip(entry[d], end[d])
		b02 := linip(b01, b11)
		b12 := linip(b11, b21)
		p[d] = linip(b02, b12)
	}
	return p
}
