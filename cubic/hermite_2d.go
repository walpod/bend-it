package cubic

import (
	"fmt"
	"github.com/walpod/bend-it"
	"gonum.org/v1/gonum/mat"
)

type HermiteVx2 struct {
	x, y      float64
	entryTan  *Control
	exitTan   *Control
	dependent bool // are the two tangents dependent on each other?
}

func NewHermiteVx2(x float64, y float64, entryTan *Control, exitTan *Control) *HermiteVx2 {
	dependent := false

	// handle dependent tangents
	if entryTan == nil && exitTan != nil {
		entryTan = NewControl(exitTan.x, exitTan.y)
		dependent = true
	} else if entryTan != nil && exitTan == nil {
		exitTan = NewControl(entryTan.x, entryTan.y)
		dependent = true
	}

	return &HermiteVx2{x, y, entryTan, exitTan, dependent}
}

func NewHermiteVx2Raw(x float64, y float64) *HermiteVx2 {
	return NewHermiteVx2(x, y, nil, nil)
}

func (vt HermiteVx2) Coord() (x, y float64) {
	return vt.x, vt.y
}

func (vt HermiteVx2) EntryTan() *Control {
	return vt.entryTan
}

func (vt HermiteVx2) ExitTan() *Control {
	return vt.exitTan
}

func (vt HermiteVx2) Tan(isEntry bool) *Control {
	if isEntry {
		return vt.entryTan
	} else {
		return vt.exitTan
	}
}

// absolute control point (as opposed to relative tangent)
func (vt HermiteVx2) Control(isEntry bool) *Control {
	if isEntry {
		return NewControl(vt.x-vt.entryTan.x, vt.y-vt.entryTan.y)
	} else {
		return NewControl(vt.x+vt.exitTan.x, vt.y+vt.exitTan.y)
	}
}

func (vt HermiteVx2) Dependent() bool {
	return vt.dependent
}

func (vt HermiteVx2) Translate(dx, dy float64) bendit.Vertex2d {
	return NewHermiteVx2(vt.x+dx, vt.y+dy, vt.entryTan, vt.exitTan)
}

func (vt HermiteVx2) WithEntryTan(entryTan *Control) *HermiteVx2 {
	var exitTan *Control
	if !vt.dependent {
		exitTan = vt.exitTan
	}
	return NewHermiteVx2(vt.x, vt.y, entryTan, exitTan)
}

func (vt HermiteVx2) WithExitTan(exitTan *Control) *HermiteVx2 {
	var entryTan *Control
	if !vt.dependent {
		entryTan = vt.entryTan
	}
	return NewHermiteVx2(vt.x, vt.y, entryTan, exitTan)
}

// HermiteTanFinder2d finds tangents based on given vertices and knots
type HermiteTanFinder2d interface {
	Find(knots bendit.Knots, vertices []*HermiteVx2)
}

type HermiteSpline2d struct {
	knots     bendit.Knots
	vertices  []*HermiteVx2
	tanFinder HermiteTanFinder2d
	// internal cache of prepare
	canon    *CanonicalSpline2d
	bezier   *BezierSpline2d
	tanFound bool
}

func NewHermiteSpline2d(tknots []float64, vertices ...*HermiteVx2) *HermiteSpline2d {
	return NewHermiteSplineTanFinder2d(tknots, nil, vertices...)
}

func NewHermiteSplineTanFinder2d(tknots []float64, tanFinder HermiteTanFinder2d, vertices ...*HermiteVx2) *HermiteSpline2d {
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

func (sp *HermiteSpline2d) Vertex(knotNo int) bendit.Vertex2d {
	if knotNo >= len(sp.vertices) {
		return nil
	} else {
		return sp.vertices[knotNo]
	}
}

func (sp *HermiteSpline2d) AddVertex(knotNo int, vertex bendit.Vertex2d) (err error) {
	err = sp.knots.AddKnot(knotNo)
	if err != nil {
		return err
	}
	hvt := vertex.(*HermiteVx2)
	if knotNo == len(sp.vertices) {
		sp.vertices = append(sp.vertices, hvt)
	} else {
		sp.vertices = append(sp.vertices, nil)
		copy(sp.vertices[knotNo+1:], sp.vertices[knotNo:])
		sp.vertices[knotNo] = hvt
	}
	return nil
}

func (sp *HermiteSpline2d) UpdateVertex(knotNo int, vertex bendit.Vertex2d) (err error) {
	if !sp.knots.KnotExists(knotNo) {
		return fmt.Errorf("knotNo %v does not exist", knotNo)
	}
	sp.vertices[knotNo] = vertex.(*HermiteVx2)
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

/*func (sp *HermiteSpline2d) Add(vertex *HermiteVx2) {
	if sp.knots.IsUniform() {
		sp.knots.(*bendit.UniformKnots).Add(1)
	} else {
		sp.knots.(*bendit.NonUniformKnots).Add(1)
	}
	sp.vertices = append(sp.vertices, vertex)
	sp.ResetPrepare()
}

func (sp *HermiteSpline2d) AddL(segmentLen float64, vertex *HermiteVx2) {
	if sp.knots.IsUniform() {
		err := sp.knots.(*bendit.UniformKnots).Add(1)
		if err != nil {
			panic(err.Error())
		}
	} else {
		sp.knots.(*bendit.NonUniformKnots).Add(segmentLen)
	}
	sp.vertices = append(sp.vertices, vertex)
}*/

// Prepare execution of hermite spline by mapping to canonical and bezier representation
func (sp *HermiteSpline2d) Prepare() {
	sp.prepareCanon()
	sp.prepareBezier()
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

func (sp *HermiteSpline2d) Canonical() *CanonicalSpline2d {
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
		return NewSingleVertexCanonicalSpline2d(sp.vertices[0].x, sp.vertices[0].y)
	} else {
		return NewCanonicalSpline2d(sp.knots.External())
	}
}

func (sp *HermiteSpline2d) uniCanonical() *CanonicalSpline2d {
	const dim = 2
	// precondition: segmCnt >= 1, bs.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		avs = append(avs, vstart.x, vend.x, vstart.exitTan.x, vend.entryTan.x)
		avs = append(avs, vstart.y, vend.y, vstart.exitTan.y, vend.entryTan.y)
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

	return NewCanonicalSpline2dByMatrix(sp.knots.External(), coefs)
}

func (sp *HermiteSpline2d) nonUniCanonical() *CanonicalSpline2d {
	const dim = 2
	segmCnt := sp.knots.SegmentCnt()
	cubics := make([]Cubic2d, segmCnt)

	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		a := mat.NewDense(dim, 4, []float64{
			vstart.x, vend.x, vstart.exitTan.x, vend.entryTan.x,
			vstart.y, vend.y, vstart.exitTan.y, vend.entryTan.y,
		})

		sgl, _ := sp.knots.SegmentLen(i)
		b := mat.NewDense(4, 4, []float64{
			1, 0, -3, 2,
			0, 0, 3, -2,
			0, sgl, -2 * sgl, sgl,
			0, 0, -sgl, sgl,
		})

		var coefs mat.Dense
		coefs.Mul(a, b)

		cubics[i] = NewCubic2d(
			NewCubicPoly(coefs.At(0, 0), coefs.At(0, 1), coefs.At(0, 2), coefs.At(0, 3)),
			NewCubicPoly(coefs.At(1, 0), coefs.At(1, 1), coefs.At(1, 2), coefs.At(1, 3)))
	}

	return NewCanonicalSpline2d(sp.knots.External(), cubics...)
}

func (sp *HermiteSpline2d) At(t float64) (x, y float64) {
	if sp.canon == nil {
		sp.prepareCanon()
	}
	return sp.canon.At(t)
}

func (sp *HermiteSpline2d) Fn() bendit.Fn2d {
	if sp.canon == nil {
		sp.prepareCanon()
	}
	return sp.canon.Fn()
}

func (sp *HermiteSpline2d) prepareBezier() {
	sp.bezier = sp.Bezier()
}

func (sp *HermiteSpline2d) Bezier() *BezierSpline2d {
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
		return NewBezierSpline2d(sp.knots.External(),
			NewBezierVx2(sp.vertices[0].x, sp.vertices[0].y, nil, nil))
	} else {
		return NewBezierSpline2d(sp.knots.External())
	}
}

func (sp *HermiteSpline2d) uniBezier() *BezierSpline2d {
	const dim = 2
	// precondition: len(cubics) >= 1, bs.knots.IsUniform()
	segmCnt := sp.knots.SegmentCnt()

	avs := make([]float64, 0, dim*4*segmCnt)
	for i := 0; i < segmCnt; i++ {
		vstart, vend := sp.vertices[i], sp.vertices[i+1]
		avs = append(avs, vstart.x, vend.x, vstart.exitTan.x, vend.entryTan.x)
		avs = append(avs, vstart.y, vend.y, vstart.exitTan.y, vend.entryTan.y)
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

	return NewBezierSpline2dByMatrix(sp.knots.External(), coefs)
}

func (sp *HermiteSpline2d) Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector bendit.LineCollector2d) {
	if sp.bezier == nil {
		sp.prepareBezier()
	}
	sp.Bezier().Approx(fromSegmentNo, toSegmentNo, maxDist, collector)
}

/*
   // TODO currently deactivated
   // entry and exit tangents for given vertex
   type VertexTan2d interface {
   	EntryTan() (lx, ly float64)
   	ExitTan() (mx, my float64)
   }

   type SingleTan2d struct {
   	Mx, My float64
   }

   func NewSingleTan2d(mx float64, my float64) *SingleTan2d {
   	return &SingleTan2d{Mx: mx, My: my}
   }

   func (st *SingleTan2d) EntryTan() (lx, ly float64) {
   	// entry = exit tangent
   	return st.Mx, st.My
   }

   func (st *SingleTan2d) ExitTan() (mx, my float64) {
   	return st.Mx, st.My
   }
*/
