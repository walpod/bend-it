package cubic

import (
	"fmt"
	"github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

// HermiteTanFinder finds tangents based on given vertices and knots
type HermiteTanFinder interface {
	Find(knots bendit.Knots, vertices []*HermiteVertex)
}

type HermiteVertBuilder struct {
	knots     bendit.Knots
	vertices  []*HermiteVertex
	tanFinder HermiteTanFinder
	/*// internal cache of prepare
	canon    *CanonicalSpline
	bezier   *BezierVertBuilder
	tanFound bool*/
}

func NewHermiteVertBuilder(tknots []float64, vertices ...*HermiteVertex) *HermiteVertBuilder {
	return NewHermiteVertBuilderTanFinder(tknots, nil, vertices...)
}

func NewHermiteVertBuilderTanFinder(tknots []float64, tanFinder HermiteTanFinder, vertices ...*HermiteVertex) *HermiteVertBuilder {
	var knots bendit.Knots
	if tknots == nil {
		knots = bendit.NewUniformKnots(len(vertices))
	} else {
		if len(tknots) != len(vertices) {
			panic("tknots and vertices must have same length")
		}
		knots = bendit.NewNonUniformKnots(tknots)
	}

	herm := &HermiteVertBuilder{knots: knots, vertices: vertices, tanFinder: tanFinder}
	return herm
}

func (sp *HermiteVertBuilder) Knots() bendit.Knots {
	return sp.knots
}

func (sp *HermiteVertBuilder) Dim() int {
	if len(sp.vertices) == 0 {
		return 0
	} else {
		return sp.vertices[0].loc.Dim()
	}
}

func (sp *HermiteVertBuilder) Vertex(knotNo int) bendit.Vertex {
	if knotNo >= len(sp.vertices) {
		return nil
	} else {
		return sp.vertices[knotNo]
	}
}

func (sp *HermiteVertBuilder) AddVertex(knotNo int, vertex bendit.Vertex) (err error) {
	err = sp.knots.AddKnot(knotNo)
	if err != nil {
		return err
	}
	hvt := vertex.(*HermiteVertex)
	if knotNo == len(sp.vertices) {
		sp.vertices = append(sp.vertices, hvt)
	} else {
		sp.vertices = append(sp.vertices, nil)
		copy(sp.vertices[knotNo+1:], sp.vertices[knotNo:])
		sp.vertices[knotNo] = hvt
	}
	return nil
}

func (sp *HermiteVertBuilder) UpdateVertex(knotNo int, vertex bendit.Vertex) (err error) {
	if !sp.knots.KnotExists(knotNo) {
		return fmt.Errorf("knotNo %v does not exist", knotNo)
	}
	sp.vertices[knotNo] = vertex.(*HermiteVertex)
	return nil
}

func (sp *HermiteVertBuilder) DeleteVertex(knotNo int) (err error) {
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

func (sp *HermiteVertBuilder) Canonical() *CanonicalSpline {
	if sp.tanFinder != nil {
		sp.tanFinder.Find(sp.knots, sp.vertices)
	}

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

func (sp *HermiteVertBuilder) uniCanonical() *CanonicalSpline {
	// precondition: segmCnt >= 1, bs.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()
	dim := sp.Dim()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
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

	return NewCanonicalSplineByMatrix(sp.knots.External(), dim, coefs)
}

func (sp *HermiteVertBuilder) nonUniCanonical() *CanonicalSpline {
	segmCnt := sp.knots.SegmentCnt()
	cubics := make([]CubicPolies, segmCnt)
	dim := sp.Dim()

	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		avs := make([]float64, 0, dim*4)
		for d := 0; d < dim; d++ {
			avs = append(avs, vstart.loc[d], vend.loc[d], vstart.exit[d], vend.entry[d])
		}
		a := mat.NewDense(dim, 4, avs)
		/*a := mat.NewDense(dim, 4, []float64{
			vstart.x, vend.x, vstart.exitTan.x, vend.entryTan.x,
			vstart.y, vend.y, vstart.exitTan.y, vend.entryTan.y,
		})*/

		sgl, _ := sp.knots.SegmentLen(i)
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

	return NewCanonicalSpline(sp.knots.External(), cubics...)
}

func (sp *HermiteVertBuilder) Build() bendit.Spline {
	return sp.Canonical()
}

func (sp *HermiteVertBuilder) Bezier() *BezierVertBuilder {
	if sp.tanFinder != nil {
		sp.tanFinder.Find(sp.knots, sp.vertices)
	}

	n := len(sp.vertices)
	if n >= 2 {
		if sp.knots.IsUniform() {
			return sp.uniBezier()
		} else {
			panic("not yet implemented")
		}
	} else if n == 1 {
		// TODO or instead nil ? zv := bendit.NewZeroVec(sp.Dim())
		return NewBezierVertBuilder(sp.knots.External(),
			NewBezierVertex(sp.vertices[0].loc, nil, nil))
	} else {
		return NewBezierVertBuilder(sp.knots.External())
	}
}

func (sp *HermiteVertBuilder) uniBezier() *BezierVertBuilder {
	// precondition: len(cubics) >= 1, bs.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()
	dim := sp.Dim()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
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

	return NewBezierVertBuilderByMatrix(sp.knots.External(), dim, coefs)
}

func (sp *HermiteVertBuilder) BezierApproxer() *BezierApproxer {
	return sp.Bezier().BezierApproxer()
}

func (sp *HermiteVertBuilder) BuildApproxer() bendit.SplineApproxer {
	return sp.BezierApproxer()
}
