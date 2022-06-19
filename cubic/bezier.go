package cubic

import (
	"fmt"
	"github.com/walpod/bendigo"
	"gonum.org/v1/gonum/mat"
)

func NewBezierVertex(loc, entry, exit bendigo.Vec) *EnexVertex {
	return NewEnexVertex(loc, entry, exit, false)
}

type BezierVertBuilder struct {
	knots    bendigo.Knots
	vertices []*EnexVertex
}

func NewBezierVertBuilder(tknots []float64, vertices ...*EnexVertex) *BezierVertBuilder {
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
	vertices := make([]*EnexVertex, 0, segmCnt)
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

func (sb *BezierVertBuilder) Knots() bendigo.Knots {
	return sb.knots
}

func (sb *BezierVertBuilder) Dim() int {
	if len(sb.vertices) == 0 {
		return 0
	} else {
		return sb.vertices[0].loc.Dim()
	}
}

func (sb *BezierVertBuilder) BezierVertex(knotNo int) *EnexVertex {
	if knotNo >= len(sb.vertices) {
		return nil
	} else {
		return sb.vertices[knotNo]
	}
}

func (sb *BezierVertBuilder) Vertex(knotNo int) bendigo.Vertex {
	return sb.BezierVertex(knotNo)
}

func (sb *BezierVertBuilder) AddVertex(knotNo int, vertex bendigo.Vertex) (err error) {
	err = sb.knots.AddKnot(knotNo)
	if err != nil {
		return err
	}
	bvt := vertex.(*EnexVertex)
	if knotNo == len(sb.vertices) {
		sb.vertices = append(sb.vertices, bvt)
	} else {
		sb.vertices = append(sb.vertices, nil)
		copy(sb.vertices[knotNo+1:], sb.vertices[knotNo:])
		sb.vertices[knotNo] = bvt
	}
	return nil
}

func (sb *BezierVertBuilder) UpdateVertex(knotNo int, vertex bendigo.Vertex) (err error) {
	if !sb.knots.KnotExists(knotNo) {
		return fmt.Errorf("knotNo %v does not exist", knotNo)
	}
	sb.vertices[knotNo] = vertex.(*EnexVertex)
	return nil
}

func (sb *BezierVertBuilder) DeleteVertex(knotNo int) (err error) {
	err = sb.knots.DeleteKnot(knotNo)
	if err != nil {
		return err
	}
	if knotNo == len(sb.vertices)-1 {
		sb.vertices = sb.vertices[:knotNo]
	} else {
		sb.vertices = append(sb.vertices[:knotNo], sb.vertices[knotNo+1:]...)
	}
	return nil
}

func (sb *BezierVertBuilder) Canonical() *CanonicalSpline {
	n := len(sb.vertices)
	if n >= 2 {
		if sb.knots.IsUniform() {
			return sb.uniCanonical()
		} else {
			return sb.nonUniCanonical()
		}
	} else if n == 1 {
		return NewSingleVertexCanonicalSpline(sb.vertices[0].loc)
	} else {
		return NewCanonicalSpline(sb.knots.External())
	}
}

func (sb *BezierVertBuilder) uniCanonical() *CanonicalSpline {
	// precondition: segmCnt >= 1, sb.knots.IsUniform()
	segmCnt := sb.knots.SegmentCnt()
	dim := sb.Dim()

	avs := make([]float64, 0, segmCnt*dim*4)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sb.vertices[i], sb.vertices[i+1]
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

	return NewCanonicalSplineByMatrix(sb.knots.External(), dim, coefs)
}

func (sb *BezierVertBuilder) nonUniCanonical() *CanonicalSpline {
	// TODO implement non-uniform
	panic("not yet implemented")
}

func (sb *BezierVertBuilder) Spline() bendigo.Spline {
	return sb.Canonical()
}

func (sb *BezierVertBuilder) DeCasteljauSpline() *DeCasteljauSpline {
	segmentCnt := sb.knots.SegmentCnt()
	if segmentCnt == 0 {
		return nil
	}

	controls := make([]bendigo.Vec, 0, segmentCnt*4)
	for s := 0; s < segmentCnt; s++ {
		vtstart, vtend := sb.vertices[s], sb.vertices[s+1]
		controls = append(controls, vtstart.loc, vtstart.exit, vtend.entry, vtend.loc)
	}

	return NewDeCasteljauSpline(sb.knots, controls)
}

func (sb *BezierVertBuilder) LinApproximate(fromSegmentNo, toSegmentNo int, consumer bendigo.LineConsumer, linaxParams *bendigo.LinaxParams) {
	dim := sb.Dim()

	isFlat := func(v0, v1, v2, v3 bendigo.Vec) bool {
		v03 := v3.Sub(v0)
		return v1.Sub(v0).ProjectedVecDist(v03) <= linaxParams.MaxDist && v2.Sub(v0).ProjectedVecDist(v03) <= linaxParams.MaxDist
	}

	var subdivide func(segmentNo int, ts, te float64, v0, v1, v2, v3 bendigo.Vec)
	subdivide = func(segmentNo int, ts, te float64, v0, v1, v2, v3 bendigo.Vec) {
		if isFlat(v0, v1, v2, v3) {
			consumer.ConsumeLine(segmentNo, ts, te, v0, v3)
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
		tstart, tend, err := bendigo.SegmentTrange(sb.knots, segmentNo)
		if err == nil { // ignore nonexistent segments
			vtstart, vtend := sb.vertices[segmentNo], sb.vertices[segmentNo+1]
			subdivide(segmentNo, tstart, tend, vtstart.loc, vtstart.exit, vtend.entry, vtend.loc)
		}
	}
}

func (sb *BezierVertBuilder) LinaxSpline(linaxParams *bendigo.LinaxParams) *bendigo.LinaxSpline {
	return bendigo.BuildLinaxSpline(sb, linaxParams)
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
