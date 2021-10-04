package cubic

import (
	"fmt"
	"github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

// HermiteTanFinder2d finds tangents based on given vertices and knots
type HermiteTanFinder2d interface {
	Find(knots bendit.Knots, vertices []*HermiteVertex)
}

type HermiteSpline2d struct {
	knots     bendit.Knots
	vertices  []*HermiteVertex
	tanFinder HermiteTanFinder2d
	// internal cache of prepare
	canon    *CanonicalSpline
	bezier   *VertBezierBuilder
	tanFound bool
}

func NewHermiteSpline2d(tknots []float64, vertices ...*HermiteVertex) *HermiteSpline2d {
	return NewHermiteSplineTanFinder2d(tknots, nil, vertices...)
}

func NewHermiteSplineTanFinder2d(tknots []float64, tanFinder HermiteTanFinder2d, vertices ...*HermiteVertex) *HermiteSpline2d {
	var knots bendit.Knots
	if tknots == nil {
		knots = bendit.NewUniformKnots(len(vertices))
	} else {
		if len(tknots) != len(vertices) {
			panic("tknots and vertices must have same length")
		}
		knots = bendit.NewNonUniformKnots(tknots)
	}

	herm := &HermiteSpline2d{knots: knots, vertices: vertices, tanFinder: tanFinder, canon: nil, bezier: nil, tanFound: false}
	return herm
}

func (sp *HermiteSpline2d) Knots() bendit.Knots {
	return sp.knots
}

func (sp *HermiteSpline2d) Dim() int {
	if len(sp.vertices) == 0 {
		return 0
	} else {
		return sp.vertices[0].loc.Dim()
	}
}

func (sp *HermiteSpline2d) Vertex(knotNo int) bendit.Vertex {
	if knotNo >= len(sp.vertices) {
		return nil
	} else {
		return sp.vertices[knotNo]
	}
}

func (sp *HermiteSpline2d) AddVertex(knotNo int, vertex bendit.Vertex) (err error) {
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

func (sp *HermiteSpline2d) UpdateVertex(knotNo int, vertex bendit.Vertex) (err error) {
	if !sp.knots.KnotExists(knotNo) {
		return fmt.Errorf("knotNo %v does not exist", knotNo)
	}
	sp.vertices[knotNo] = vertex.(*HermiteVertex)
	return nil
}

func (sp *HermiteSpline2d) DeleteVertex(knotNo int) (err error) {
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

// Prepare execution of hermite spline by mapping to canonical and bezier representation
func (sp *HermiteSpline2d) Prepare() {
	sp.prepareCanon()
	// TODO sp.prepareBezier()
}

func (sp *HermiteSpline2d) ResetPrepare() {
	sp.tanFound = false
	sp.canon = nil
	sp.bezier = nil
}

func (sp *HermiteSpline2d) prepareTan() {
	if sp.tanFinder != nil {
		sp.tanFinder.Find(sp.knots, sp.vertices)
		sp.tanFound = true
	}
}

func (sp *HermiteSpline2d) prepareCanon() {
	sp.canon = sp.Canonical()
}

func (sp *HermiteSpline2d) Canonical() *CanonicalSpline {
	if sp.tanFinder != nil && !sp.tanFound {
		sp.prepareTan()
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

func (sp *HermiteSpline2d) uniCanonical() *CanonicalSpline {
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

func (sp *HermiteSpline2d) nonUniCanonical() *CanonicalSpline {
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
		cubics[i] = NewCubicPolyNd(cubs...)
		/*cubics[i] = NewCubicPolyNd(
		NewCubicPoly(coefs.At(0, 0), coefs.At(0, 1), coefs.At(0, 2), coefs.At(0, 3)),
		NewCubicPoly(coefs.At(1, 0), coefs.At(1, 1), coefs.At(1, 2), coefs.At(1, 3)))*/
	}

	return NewCanonicalSpline(sp.knots.External(), cubics...)
}

// At evaluates point on hermite spline for given parameter t
// Prepare must be called before
func (sp *HermiteSpline2d) At(t float64) bendit.Vec {
	return sp.canon.At(t)
}

func (sp *HermiteSpline2d) prepareBezier() {
	sp.bezier = sp.Bezier()
}

func (sp *HermiteSpline2d) Bezier() *VertBezierBuilder {
	if sp.tanFinder != nil && !sp.tanFound {
		sp.prepareTan()
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
		return NewVertBezierBuilder(sp.knots.External(),
			NewBezierVertex(sp.vertices[0].loc, nil, nil))
	} else {
		return NewVertBezierBuilder(sp.knots.External())
	}
}

func (sp *HermiteSpline2d) uniBezier() *VertBezierBuilder {
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

	return NewVertBezierBuilderdByMatrix(sp.knots.External(), dim, coefs)
}

/*func (sp *HermiteSpline2d) Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector bendit.LineCollector2d) {
	// TODO Prepare should be called before (as precondition) or leave it as it is?
	sp.Bezier().Approx(fromSegmentNo, toSegmentNo, maxDist, collector)
}*/
