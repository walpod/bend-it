package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"math"
	"math/rand"
	"testing"
)

// START some general bend-it spline Asserts
const delta = 0.0000000001

func AssertSplineAt(t *testing.T, spline bendit.Spline2d, atT float64, expx, expy float64) {
	x, y := spline.At(atT)
	assert.InDeltaf(t, expx, x, delta, "spline.At(%v).x = %v != expected %v", atT, x, expx)
	assert.InDeltaf(t, expy, y, delta, "spline.At(%v).y = %v != expected %v", atT, y, expy)
}

func AssertSplinesEqualInRange(t *testing.T, spline0 bendit.Spline2d, spline1 bendit.Spline2d, tstart, tend float64, sampleCnt int) {
	for i := 0; i < sampleCnt; i++ {
		atT := rand.Float64()*(tend-tstart) + tstart
		x0, y0 := spline0.At(atT)
		x1, y1 := spline1.At(atT)
		assert.InDeltaf(t, x0, x1, delta, "spline0.At(%v).x = %v != spline1.At(%v).x = %v", atT, x0, atT, x1)
		assert.InDeltaf(t, y0, y1, delta, "spline0.At(%v).y = %v != spline1.At(%v).y = %v", atT, y0, atT, y1)
	}
}

func AssertSplinesEqual(t *testing.T, spline0 bendit.Spline2d, spline1 bendit.Spline2d, sampleCnt int) {
	// assert over full domain
	AssertSplinesEqualInRange(t, spline0, spline1, spline0.Knots().Tstart(), spline0.Knots().Tend(), sampleCnt)
}

func AssertApproxStartPointsMatchSpline(t *testing.T, lines []bendit.LineParams, spline bendit.Spline2d) {
	for _, lin := range lines {
		x, y := spline.At(lin.Tstart)
		assert.InDeltaf(t, x, lin.Pstartx, delta, "spline.At(%v).x = %v != start-point.x = %v of approximated line", lin.Tstart, x, lin.Pstartx)
		assert.InDeltaf(t, y, lin.Pstarty, delta, "spline.At(%v).y = %v != start-point.y = %v of approximated line", lin.Tstart, y, lin.Pstarty)
	}
}

func AssertRandSplinePointProperty(t *testing.T, spline bendit.Spline2d, hasProp func(x, y float64) bool, msg string) {
	ts, te := spline.Knots().Tstart(), spline.Knots().Tend()
	atT := ts + rand.Float64()*(te-ts)
	x, y := spline.At(atT)
	assert.True(t, hasProp(x, y), msg)
}

// END some general bend-it spline Asserts

func AssertBezierAtDeCasteljau(t *testing.T, bezier *BezierSpline2d, atT float64) {
	x, y := bezier.At(atT)
	xdc, ydc := bezier.AtDeCasteljau(atT)
	//fmt.Printf("bezier.AtDeCasteljau(%v) = (%v, %v) \n", atT, xdc, ydc)
	assert.InDeltaf(t, xdc, x, delta, "spline.At(%v).x = %v != spline.AtDeCasteljau(%v).x = %v", atT, x, atT, xdc)
	assert.InDeltaf(t, ydc, y, delta, "spline.At(%v).y = %v != spline.AtDeCasteljau(%v).y = %v", atT, y, atT, ydc)
}

func AssertControlsAreEqual(t *testing.T, expected *Control, actual *Control, isEntry bool) {
	var side string
	if isEntry {
		side = "entry"
	} else {
		side = "exit"
	}
	assert.InDeltaf(t, expected.x, actual.x, delta, "expected %v-control.x = %v != actual.x = %v", side, expected.x, actual.x)
	assert.InDeltaf(t, expected.y, actual.y, delta, "expected %v-control.y = %v != actual.y = %v", side, expected.y, actual.y)
}

func AssertBezierVxAreEqual(t *testing.T, expected *BezierVx2, expectedDependent bool, actual *BezierVx2) {
	assert.InDeltaf(t, expected.x, actual.x, delta, "expected bezier.x = %v != actual bezier.x = %v", expected.x, actual.x)
	assert.InDeltaf(t, expected.y, actual.y, delta, "expected bezier.y = %v != actual bezier.y = %v", expected.y, actual.y)
	AssertControlsAreEqual(t, expected.entry, actual.entry, true)
	AssertControlsAreEqual(t, expected.exit, actual.exit, false)
	assert.Equal(t, expectedDependent, actual.dependent, "expected dependent = %v != actual dependent = %v", expectedDependent, actual.dependent)
}

// createBezierDiag00to11 creates a bezier representing a straight line from (0,0) to (1,1)
func createBezierDiag00to11() *BezierSpline2d {
	return NewBezierSpline2d(nil,
		NewBezierVx2(0, 0, nil, NewControl(1./3, 1./3)),
		NewBezierVx2(1, 1, NewControl(2./3, 2./3), nil),
	)
}

// createBezierDiag00to11 creates a bezier representing an S-formed slope from (0,0) to (1,1)
func createBezierS00to11() *BezierSpline2d {
	return NewBezierSpline2d(nil,
		NewBezierVx2(0, 0, nil, NewControl(1, 0)),
		NewBezierVx2(1, 1, NewControl(0, 1), nil),
	)
}

// createBezierDiag00to11 creates two consecutive beziers representing an S-formed slope from (0,0) to (1,1) or (1,1) to (2,2), resp.
func createDoubleBezierS00to11to22() *BezierSpline2d {
	return NewBezierSpline2d(nil,
		NewBezierVx2(0, 0, nil, NewControl(1, 0)),
		NewBezierVx2(1, 1 /*NewControl(0, 1)*/, nil, NewControl(2, 1)),
		NewBezierVx2(2, 2, NewControl(1, 2), nil),
	)
}

func TestBezierSpline2d_At(t *testing.T) {
	bezier := createBezierDiag00to11()
	AssertSplineAt(t, bezier, 0, 0, 0)
	AssertSplineAt(t, bezier, 0.25, 0.25, 0.25)
	AssertSplineAt(t, bezier, .5, .5, .5)
	AssertSplineAt(t, bezier, 0.75, 0.75, 0.75)
	AssertSplineAt(t, bezier, 1, 1, 1)

	bezier = createDoubleBezierS00to11to22()
	AssertSplineAt(t, bezier, 0, 0, 0)
	AssertSplineAt(t, bezier, 0.5, 0.5, 0.5)
	AssertSplineAt(t, bezier, 1, 1, 1)
	AssertSplineAt(t, bezier, 1.5, 1.5, 1.5)
	AssertSplineAt(t, bezier, 2, 2, 2)

	// single vertex, domain with value 0 only
	bezier = NewBezierSpline2d(nil,
		NewBezierVx2(1, 2, nil, nil))
	AssertSplineAt(t, bezier, 0, 1, 2)

	bezier = NewBezierSpline2d(
		[]float64{0},
		NewBezierVx2(1, 2, nil, nil))
	AssertSplineAt(t, bezier, 0, 1, 2)

	// empty domain
	bezier = NewBezierSpline2d(nil)
	bezier = NewBezierSpline2d([]float64{})
}

func TestBezierSpline2d_AtDeCasteljau(t *testing.T) {
	bezier := createBezierS00to11()
	AssertBezierAtDeCasteljau(t, bezier, 0)
	AssertBezierAtDeCasteljau(t, bezier, 0.1)
	AssertBezierAtDeCasteljau(t, bezier, 0.25)
	AssertBezierAtDeCasteljau(t, bezier, 0.5)
	AssertBezierAtDeCasteljau(t, bezier, 0.75)
	AssertBezierAtDeCasteljau(t, bezier, 0.9)
	AssertBezierAtDeCasteljau(t, bezier, 1)
}

func TestBezierSpline2d_Canonical(t *testing.T) {
	bezier := createBezierS00to11()
	AssertSplinesEqual(t, bezier, bezier.Canonical(), 100)

	bezier = createDoubleBezierS00to11to22()
	AssertSplinesEqual(t, bezier, bezier.Canonical(), 100)
}

func TestBezierSpline2d_Approx(t *testing.T) {
	bezier := createBezierDiag00to11()
	lc := bendit.NewLineToSliceCollector2d()
	bendit.ApproxAll(bezier, 0.1, lc)
	assert.Len(t, lc.Lines, 1, "approximated with one line")
	assert.InDeltaf(t, 0., lc.Lines[0].Pstartx, delta, "start point x=0")
	assert.InDeltaf(t, 0., lc.Lines[0].Pstarty, delta, "start point y=0")
	assert.InDeltaf(t, 1., lc.Lines[0].Pendx, delta, "end point x=0")
	assert.InDeltaf(t, 1., lc.Lines[0].Pendy, delta, "end point y=0")

	// start points of approximated lines must be on bezier curve and match bezier.At
	bezier = createBezierS00to11()
	lc = bendit.NewLineToSliceCollector2d()
	bendit.ApproxAll(bezier, 0.02, lc)
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	AssertApproxStartPointsMatchSpline(t, lc.Lines, bezier)
}

func TestBezierVx2Dependent(t *testing.T) {
	bvx := NewBezierVx2(0, 0, NewControl(1, 2), nil)
	AssertControlsAreEqual(t, NewControl(-bvx.Entry().x, -bvx.Entry().y), bvx.exit, false)
	bvx = NewBezierVx2(0, 0, nil, NewControl(3, -5))
	AssertControlsAreEqual(t, bvx.entry, NewControl(-bvx.Exit().x, -bvx.Exit().y), true)
}

func TestBezierVx2_Move(t *testing.T) {
	bvx := NewBezierVx2(0, 0, NewControl(0, 1), nil).
		Move(2, 0)
	AssertBezierVxAreEqual(t, NewBezierVx2(2, 0, NewControl(2, 1), NewControl(2, -1)), true, bvx)
	bvx = NewBezierVx2(0, 0, NewControl(0, 2), NewControl(3, 0)).
		Move(1, 1)
	AssertBezierVxAreEqual(t, NewBezierVx2(1, 1, NewControl(1, 3), NewControl(4, 1)), false, bvx)
}

func TestBezierVx2_WithEntry(t *testing.T) {
	bvx := NewBezierVx2(0, 0, nil, NewControl(0, 1)).
		WithEntry(NewControl(2, 2))
	AssertBezierVxAreEqual(t, NewBezierVx2(0, 0, NewControl(2, 2), NewControl(-2, -2)), true, bvx)
	bvx = NewBezierVx2(0, 0, NewControl(0, 1), NewControl(0, 1)).
		WithEntry(NewControl(2, 2))
	AssertBezierVxAreEqual(t, NewBezierVx2(0, 0, NewControl(2, 2), NewControl(0, 1)), false, bvx)
}

func TestBezierVx2_WithExit(t *testing.T) {
	bvx := NewBezierVx2(0, 0, NewControl(0, 1), nil).
		WithExit(NewControl(-2, -2))
	AssertBezierVxAreEqual(t, NewBezierVx2(0, 0, NewControl(2, 2), NewControl(-2, -2)), true, bvx)
	bvx = NewBezierVx2(0, 0, NewControl(0, 1), NewControl(0, 1)).
		WithExit(NewControl(2, 2))
	AssertBezierVxAreEqual(t, NewBezierVx2(0, 0, NewControl(0, 1), NewControl(2, 2)), false, bvx)
}

func TestProjectedVectorDist(t *testing.T) {
	assert.InDeltaf(t, 1., ProjectedVectorDist(0, 1, 1, 0), delta, "unit square")
	assert.InDeltaf(t, 1., ProjectedVectorDist(1./math.Sqrt2, 1./math.Sqrt2, 1./math.Sqrt2, -1./math.Sqrt2), delta, "unit square, rotated")
	assert.InDeltaf(t, 2., ProjectedVectorDist(0, 2, 2, 0), delta, "square - side length 2")

	assert.InDeltaf(t, 1., ProjectedVectorDist(0, 1, 2, 0), delta, "rectangle")
	assert.InDeltaf(t, 1., ProjectedVectorDist(1./math.Sqrt2, 1./math.Sqrt2, math.Sqrt2, -math.Sqrt2), delta, "rectangle, rotated")
	assert.InDeltaf(t, 1., ProjectedVectorDist(1./math.Sqrt2, 1./math.Sqrt2, 10, -10), delta, "rectangle, rotated, enlarged")

	assert.InDeltaf(t, 1., ProjectedVectorDist(1, 1, 1, 0), delta, "45 degree")
	assert.InDeltaf(t, 1., ProjectedVectorDist(1, 1, -10, 0), delta, "45 degree, other direction")
	assert.InDeltaf(t, 1., ProjectedVectorDist(math.Sqrt2, 0, -3, 3), delta, "45 degree")
}
