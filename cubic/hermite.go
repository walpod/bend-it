package cubic

import (
	"fmt"
	"github.com/walpod/bendigo"
	"gonum.org/v1/gonum/mat"
)

/*// HermiteTanFinder finds tangents based on given vertices and knots
type HermiteTanFinder interface {
	Find(knots bendigo.Knots, vertices []*HermiteVertex)
}*/

type HermiteVertBuilder struct {
	knots    bendigo.Knots
	vertices []*HermiteVertex
}

func NewHermiteVertBuilder(tknots []float64, vertices ...*HermiteVertex) *HermiteVertBuilder {
	var knots bendigo.Knots
	if tknots == nil {
		knots = bendigo.NewUniformKnots(len(vertices))
	} else {
		if len(tknots) != len(vertices) {
			panic("tknots and vertices must have same length")
		}
		knots = bendigo.NewNonUniformKnots(tknots)
	}

	herm := &HermiteVertBuilder{knots: knots, vertices: vertices}
	return herm
}

func (sb *HermiteVertBuilder) Knots() bendigo.Knots {
	return sb.knots
}

func (sb *HermiteVertBuilder) Dim() int {
	if len(sb.vertices) == 0 {
		return 0
	} else {
		return sb.vertices[0].loc.Dim()
	}
}

func (sb *HermiteVertBuilder) Vertex(knotNo int) bendigo.Vertex {
	if knotNo >= len(sb.vertices) {
		return nil
	} else {
		return sb.vertices[knotNo]
	}
}

func (sb *HermiteVertBuilder) AddVertex(knotNo int, vertex bendigo.Vertex) (err error) {
	err = sb.knots.AddKnot(knotNo)
	if err != nil {
		return err
	}
	hvt := vertex.(*HermiteVertex)
	if knotNo == len(sb.vertices) {
		sb.vertices = append(sb.vertices, hvt)
	} else {
		sb.vertices = append(sb.vertices, nil)
		copy(sb.vertices[knotNo+1:], sb.vertices[knotNo:])
		sb.vertices[knotNo] = hvt
	}
	return nil
}

func (sb *HermiteVertBuilder) UpdateVertex(knotNo int, vertex bendigo.Vertex) (err error) {
	if !sb.knots.KnotExists(knotNo) {
		return fmt.Errorf("knotNo %v does not exist", knotNo)
	}
	sb.vertices[knotNo] = vertex.(*HermiteVertex)
	return nil
}

func (sb *HermiteVertBuilder) DeleteVertex(knotNo int) (err error) {
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

func (sb *HermiteVertBuilder) Canonical() *CanonicalSpline {
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

func (sb *HermiteVertBuilder) uniCanonical() *CanonicalSpline {
	// precondition: segmCnt >= 1, bs.knots.IsUniform()
	segmCnt := sb.knots.SegmentCnt()
	dim := sb.Dim()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sb.vertices[i], sb.vertices[i+1]
		for d := 0; d < dim; d++ {
			avs = append(avs, vstart.loc[d], vend.loc[d], vstart.exit[d], vend.entry[d])
		}
	}
	a := mat.NewDense(dim*segmCnt, 4, avs)

	b := mat.NewDense(4, 4, []float64{
		1, 0, -3, 2,
		0, 0, 3, -2,
		0, 1, -2, 1,
		0, 0, -1, 1,
	})

	var coefs mat.Dense
	coefs.Mul(a, b)

	return NewCanonicalSplineByMatrix(sb.knots.External(), dim, coefs)
}

func (sb *HermiteVertBuilder) nonUniCanonical() *CanonicalSpline {
	segmCnt := sb.knots.SegmentCnt()
	cubics := make([]CubicPolies, segmCnt)
	dim := sb.Dim()

	for i := 0; i < segmCnt; i++ {
		vstart, vend := sb.vertices[i], sb.vertices[i+1]
		avs := make([]float64, 0, dim*4)
		for d := 0; d < dim; d++ {
			avs = append(avs, vstart.loc[d], vend.loc[d], vstart.exit[d], vend.entry[d])
		}
		a := mat.NewDense(dim, 4, avs)
		/*a := mat.NewDense(dim, 4, []float64{
			vstart.x, vend.x, vstart.exitTan.x, vend.entryTan.x,
			vstart.y, vend.y, vstart.exitTan.y, vend.entryTan.y,
		})*/

		sgl, _ := sb.knots.SegmentLen(i)
		b := mat.NewDense(4, 4, []float64{
			1, 0, -3, 2,
			0, 0, 3, -2,
			0, sgl, -2 * sgl, sgl,
			0, 0, -sgl, sgl,
		})

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubs := make([]CubicPoly, dim)
		for d := 0; d < dim; d++ {
			cubs[d] = NewCubicPoly(coefs.At(d, 0), coefs.At(d, 1), coefs.At(d, 2), coefs.At(d, 3))
		}
		cubics[i] = NewCubicPolies(cubs...)
		/*cubics[i] = NewCubicPolies(
		NewCubicPoly(coefs.At(0, 0), coefs.At(0, 1), coefs.At(0, 2), coefs.At(0, 3)),
		NewCubicPoly(coefs.At(1, 0), coefs.At(1, 1), coefs.At(1, 2), coefs.At(1, 3)))*/
	}

	return NewCanonicalSpline(sb.knots.External(), cubics...)
}

func (sb *HermiteVertBuilder) Spline() bendigo.Spline {
	return sb.Canonical()
}

func (sb *HermiteVertBuilder) Bezier() *BezierVertBuilder {
	n := len(sb.vertices)
	if n >= 2 {
		if sb.knots.IsUniform() {
			return sb.uniBezier()
		} else {
			panic("not yet implemented") // TODO
		}
	} else if n == 1 {
		// TODO or instead nil ? zv := bendigo.NewZeroVec(sb.Dim())
		return NewBezierVertBuilder(sb.knots.External(),
			NewBezierVertex(sb.vertices[0].loc, nil, nil))
	} else {
		return NewBezierVertBuilder(sb.knots.External())
	}
}

func (sb *HermiteVertBuilder) uniBezier() *BezierVertBuilder {
	// precondition: len(cubics) >= 1, bs.knots.IsUniform()
	segmCnt := sb.knots.SegmentCnt()
	dim := sb.Dim()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sb.vertices[i], sb.vertices[i+1]
		for d := 0; d < dim; d++ {
			avs = append(avs, vstart.loc[d], vend.loc[d], vstart.exit[d], vend.entry[d])
		}
	}
	a := mat.NewDense(dim*segmCnt, 4, avs)

	b := mat.NewDense(4, 4, []float64{
		1, 1, 0, 0,
		0, 0, 1, 1,
		0, 1. / 3, 0, 0,
		0, 0, -1. / 3, 0,
	})

	var coefs mat.Dense
	coefs.Mul(a, b)

	return NewBezierVertBuilderByMatrix(sb.knots.External(), dim, coefs)
}

func (sb *HermiteVertBuilder) LinApproximate(fromSegmentNo, toSegmentNo int, consumer bendigo.LineConsumer, linaxParams *bendigo.LinaxParams) {
	sb.Bezier().LinApproximate(fromSegmentNo, toSegmentNo, consumer, linaxParams)
}

func (sb *HermiteVertBuilder) LinaxSpline(linaxParams *bendigo.LinaxParams) *bendigo.LinaxSpline {
	return bendigo.BuildLinaxSpline(sb, linaxParams)
}
