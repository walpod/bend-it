package cubic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"math/rand"
	"testing"
)

// START some general bend-it spline Asserts
const delta = 0.0000000001

func AssertVecInDelta(t *testing.T, expected bendit.Vec, actual bendit.Vec, msg string) {
	assert.Equal(t, expected.Dim(), actual.Dim(), "dimension of expected = %v != dimension of actual %v", expected.Dim(), actual.Dim())
	for d := 0; d < expected.Dim(); d++ {
		assert.InDeltaf(t, expected[d], actual[d], delta, msg+", at dim = %v", d)
	}
}

func AssertSplineAt(t *testing.T, spline bendit.Spline2d, atT float64, expected bendit.Vec) {
	actual := spline.At(atT)
	AssertVecInDelta(t, expected, actual, fmt.Sprintf("spline0.At(%v) = %v != spline1.At(%v) = %v", atT, expected, atT, actual))
}

func AssertSplinesEqualInRange(t *testing.T, spline0 bendit.Spline2d, spline1 bendit.Spline2d, tstart, tend float64, sampleCnt int) {
	for i := 0; i < sampleCnt; i++ {
		atT := rand.Float64()*(tend-tstart) + tstart
		v0 := spline0.At(atT)
		v1 := spline1.At(atT)
		AssertVecInDelta(t, v0, v1, fmt.Sprintf("spline0.At(%v).x = %v != spline1.At(%v).x = %v", atT, v0, atT, v1))
	}
}

func AssertSplinesEqual(t *testing.T, spline0 bendit.Spline2d, spline1 bendit.Spline2d, sampleCnt int) {
	// assert over full domain
	AssertSplinesEqualInRange(t, spline0, spline1, spline0.Knots().Tstart(), spline0.Knots().Tend(), sampleCnt)
}

func AssertApproxStartPointsMatchSpline(t *testing.T, lines []bendit.LineParams, spline bendit.Spline2d) {
	for _, lin := range lines {
		v := spline.At(lin.Tstart)
		AssertVecInDelta(t, v, lin.Pstart, fmt.Sprintf("spline.At(%v) = %v != start-point = %v of approximated line", lin.Tstart, v, lin.Pstart))
		//assert.InDeltaf(t, x, lin.Pstartx, delta, "spline.At(%v).x = %v != start-point.x = %v of approximated line", lin.Tstart, x, lin.Pstartx)
	}
}

func AssertRandSplinePointProperty(t *testing.T, spline bendit.Spline2d, hasProp func(v bendit.Vec) bool, msg string) {
	ts, te := spline.Knots().Tstart(), spline.Knots().Tend()
	atT := ts + rand.Float64()*(te-ts)
	v := spline.At(atT)
	assert.True(t, hasProp(v), msg)
}

// END some general bend-it spline Asserts

func AssertBezierAtInDeltaDeCasteljau(t *testing.T, bezier *BezierSpline2d, atT float64) {
	v := bezier.At(atT)
	vdc := bezier.AtDeCasteljau(atT)
	//fmt.Printf("bezier.AtDeCasteljau(%loc) = (%loc, %loc) \n", atT, xdc, ydc)
	AssertVecInDelta(t, vdc, v, fmt.Sprintf("spline.At(%v) = %v != spline.AtDeCasteljau(%v) = %v", atT, v, atT, vdc))
	//assert.InDeltaf(t, xdc, x, delta, "spline.At(%loc).x = %loc != spline.AtDeCasteljau(%loc).x = %loc", atT, x, atT, xdc)
}

/*func AssertControlsAreEqual(t *testing.T, expected *Control, actual *Control, isEntry bool) {
	var side string
	if isEntry {
		side = "entry"
	} else {
		side = "exit"
	}
	assert.InDeltaf(t, expected.x, actual.x, delta, "expected %v-control.x = %v != actual.x = %v", side, expected.x, actual.x)
	assert.InDeltaf(t, expected.y, actual.y, delta, "expected %v-control.y = %v != actual.y = %v", side, expected.y, actual.y)
}*/

func AssertControlVerticesAreEqual(t *testing.T, expected *ControlVertex, expectedDependent bool, actual *ControlVertex) {
	AssertVecInDelta(t, expected.loc, actual.loc, fmt.Sprintf("expected bezier = %v != actual bezier = %v", expected.loc, actual.loc))
	AssertVecInDelta(t, expected.entry, actual.entry, fmt.Sprintf("expected entry-control = %v != actual = %v", expected.entry, actual.entry))
	AssertVecInDelta(t, expected.exit, actual.exit, fmt.Sprintf("expected exit-control = %v != actual = %v", expected.entry, actual.entry))
	assert.Equal(t, expectedDependent, actual.dependent, "expected dependent = %v != actual dependent = %v", expectedDependent, actual.dependent)
}

// createBezierDiag00to11 creates a bezier representing a straight line from (0,0) to (1,1)
func createBezierDiag00to11() *BezierSpline2d {
	return NewBezierSpline2d(nil,
		NewControlVertex(bendit.NewVec(0, 0), nil, bendit.NewVec(1./3, 1./3)),
		NewControlVertex(bendit.NewVec(1, 1), bendit.NewVec(2./3, 2./3), nil),
	)
}

// createBezierDiag00to11 creates a bezier representing an S-formed slope from (0,0) to (1,1)
func createBezierS00to11() *BezierSpline2d {
	return NewBezierSpline2d(nil,
		NewControlVertex(bendit.NewVec(0, 0), nil, bendit.NewVec(1, 0)),
		NewControlVertex(bendit.NewVec(1, 1), bendit.NewVec(0, 1), nil),
	)
}

// createBezierDiag00to11 creates two consecutive beziers representing an S-formed slope from (0,0) to (1,1) or (1,1) to (2,2), resp.
func createDoubleBezierS00to11to22() *BezierSpline2d {
	return NewBezierSpline2d(nil,
		NewControlVertex(bendit.NewVec(0, 0), nil, bendit.NewVec(1, 0)),
		NewControlVertex(bendit.NewVec(1, 1) /*bendit.NewVec(0, 1)*/, nil, bendit.NewVec(2, 1)),
		NewControlVertex(bendit.NewVec(2, 2), bendit.NewVec(1, 2), nil),
	)
}

func TestBezierSpline2d_At(t *testing.T) {
	bezier := createBezierDiag00to11()
	bezier.Prepare()
	AssertSplineAt(t, bezier, 0, bendit.NewVec(0, 0))
	AssertSplineAt(t, bezier, 0.25, bendit.NewVec(0.25, 0.25))
	AssertSplineAt(t, bezier, .5, bendit.NewVec(.5, .5))
	AssertSplineAt(t, bezier, 0.75, bendit.NewVec(0.75, 0.75))
	AssertSplineAt(t, bezier, 1, bendit.NewVec(1, 1))

	bezier = createDoubleBezierS00to11to22()
	bezier.Prepare()
	AssertSplineAt(t, bezier, 0, bendit.NewVec(0, 0))
	AssertSplineAt(t, bezier, 0.5, bendit.NewVec(0.5, 0.5))
	AssertSplineAt(t, bezier, 1, bendit.NewVec(1, 1))
	AssertSplineAt(t, bezier, 1.5, bendit.NewVec(1.5, 1.5))
	AssertSplineAt(t, bezier, 2, bendit.NewVec(2, 2))

	// single vertex, domain with value 0 only
	bezier = NewBezierSpline2d(nil,
		NewControlVertex(bendit.NewVec(1, 2), nil, nil))
	bezier.Prepare()
	AssertSplineAt(t, bezier, 0, bendit.NewVec(1, 2))

	bezier = NewBezierSpline2d(
		[]float64{0},
		NewControlVertex(bendit.NewVec(1, 2), nil, nil))
	bezier.Prepare()
	AssertSplineAt(t, bezier, 0, bendit.NewVec(1, 2))

	// empty domain
	bezier = NewBezierSpline2d(nil)
	bezier = NewBezierSpline2d([]float64{})
}

func TestBezierSpline2d_AtDeCasteljau(t *testing.T) {
	bezier := createBezierS00to11()
	bezier.Prepare()
	AssertBezierAtInDeltaDeCasteljau(t, bezier, 0)
	AssertBezierAtInDeltaDeCasteljau(t, bezier, 0.1)
	AssertBezierAtInDeltaDeCasteljau(t, bezier, 0.25)
	AssertBezierAtInDeltaDeCasteljau(t, bezier, 0.5)
	AssertBezierAtInDeltaDeCasteljau(t, bezier, 0.75)
	AssertBezierAtInDeltaDeCasteljau(t, bezier, 0.9)
	AssertBezierAtInDeltaDeCasteljau(t, bezier, 1)
}

func TestBezierSpline2d_Canonical(t *testing.T) {
	bezier := createBezierS00to11()
	bezier.Prepare()
	AssertSplinesEqual(t, bezier, bezier.Canonical(), 100)

	bezier = createDoubleBezierS00to11to22()
	bezier.Prepare()
	AssertSplinesEqual(t, bezier, bezier.Canonical(), 100)
}

func TestBezierSpline2d_Approx(t *testing.T) {
	bezier := createBezierDiag00to11()
	lc := bendit.NewLineToSliceCollector2d()
	bendit.ApproxAll(bezier, 0.1, lc)
	assert.Len(t, lc.Lines, 1, "approximated with one line")
	AssertVecInDelta(t, bendit.NewVec(0, 0), lc.Lines[0].Pstart, "start point = [0,0]")
	AssertVecInDelta(t, bendit.NewVec(1, 1), lc.Lines[0].Pend, "end point = [1,1]")
	/*assert.InDeltaf(t, 0., lc.Lines[0].Pstartx, delta, "start point x=0")
	assert.InDeltaf(t, 0., lc.Lines[0].Pstarty, delta, "start point y=0")
	assert.InDeltaf(t, 1., lc.Lines[0].Pendx, delta, "end point x=0")
	assert.InDeltaf(t, 1., lc.Lines[0].Pendy, delta, "end point y=0")*/

	// start points of approximated lines must be on bezier curve and match bezier.At
	bezier = createBezierS00to11()
	bezier.Prepare()
	lc = bendit.NewLineToSliceCollector2d()
	bendit.ApproxAll(bezier, 0.02, lc)
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	AssertApproxStartPointsMatchSpline(t, lc.Lines, bezier)
}

func TestBezierSpline2d_AddVertex(t *testing.T) {
	bezier := createBezierDiag00to11()
	err := bezier.AddVertex(3, nil)
	assert.NotNil(t, err, "knot-no. too large")
	err = bezier.AddVertex(2, NewControlVertex(bendit.NewVec(2, 2), bendit.NewVec(1.5, 1.5), nil))
	assert.Equal(t, bezier.knots.KnotCnt(), 3, "knot-cnt %v wrong", bezier.knots.KnotCnt())
	err = bezier.AddVertex(0, NewControlVertex(bendit.NewVec(-1, -1), bendit.NewVec(-2, -2), nil))
	assert.Equal(t, bezier.knots.KnotCnt(), 4, "knot-cnt %v wrong", bezier.knots.KnotCnt())
	assert.Equal(t, bezier.Vertex(1), createBezierDiag00to11().Vertex(0), "vertices don't match")
	assert.Equal(t, bezier.Vertex(2), createBezierDiag00to11().Vertex(1), "vertices don't match")
}

func TestBezierSpline2d_DeleteVertex(t *testing.T) {
	bezier := createBezierDiag00to11()
	err := bezier.DeleteVertex(2)
	assert.NotNil(t, err, "knot-no. doesn't exist")
	err = bezier.DeleteVertex(1)
	assert.Equal(t, bezier.knots.KnotCnt(), 1, "knot-cnt %v wrong", bezier.knots.KnotCnt())
	assert.Equal(t, bezier.Vertex(0), createBezierDiag00to11().Vertex(0), "vertices don't match")
	err = bezier.DeleteVertex(0)
	assert.Equal(t, bezier.knots.KnotCnt(), 0, "knot-cnt %v wrong", bezier.knots.KnotCnt())
}

func TestBezierVx2Dependent(t *testing.T) {
	bvx := NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(1, 2), nil)
	AssertVecInDelta(t, bvx.entry.Negate(), bvx.exit, "dependent control must be reflected by origin [0,0]")
	bvx = NewControlVertex(bendit.NewVec(0, 0), nil, bendit.NewVec(3, -5))
	AssertVecInDelta(t, bvx.entry, bvx.exit.Negate(), "dependent control must be reflected by origin [0,0]")
}

func TestBezierVx2_Translate(t *testing.T) {
	bvx := NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), nil).Translate(bendit.NewVec(2, 0)).(*ControlVertex)
	AssertControlVerticesAreEqual(t, NewControlVertex(bendit.NewVec(2, 0), bendit.NewVec(2, 1), bendit.NewVec(2, -1)), true, bvx)
	bvx = NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 2), bendit.NewVec(3, 0)).Translate(bendit.NewVec(1, 1)).(*ControlVertex)
	AssertControlVerticesAreEqual(t, NewControlVertex(bendit.NewVec(1, 1), bendit.NewVec(1, 3), bendit.NewVec(4, 1)), false, bvx)
}

func TestBezierVx2_WithEntry(t *testing.T) {
	bvx := NewControlVertex(bendit.NewVec(0, 0), nil, bendit.NewVec(0, 1)).
		WithEntry(bendit.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(2, 2), bendit.NewVec(-2, -2)), true, bvx)
	bvx = NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), bendit.NewVec(0, 1)).
		WithEntry(bendit.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(2, 2), bendit.NewVec(0, 1)), false, bvx)
}

func TestBezierVx2_WithExit(t *testing.T) {
	bvx := NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), nil).
		WithExit(bendit.NewVec(-2, -2))
	AssertControlVerticesAreEqual(t, NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(2, 2), bendit.NewVec(-2, -2)), true, bvx)
	bvx = NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), bendit.NewVec(0, 1)).
		WithExit(bendit.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewControlVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), bendit.NewVec(2, 2)), false, bvx)
}
