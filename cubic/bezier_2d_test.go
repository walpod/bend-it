package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"math"
	"testing"
)

const delta = 0.0000000001

func AssertSplineAt(t *testing.T, spline bendit.Spline2d, atT float64, expx, expy float64) {
	x, y := spline.At(atT)
	assert.InDeltaf(t, expx, x, delta, "spline.At(%v).x = %v != expected %v", atT, x, expx)
	assert.InDeltaf(t, expy, y, delta, "spline.At(%v).y = %v != expected %v", atT, y, expy)
}

func AssertBezierAtDeCasteljau(t *testing.T, bezier *BezierSpline2d, atT float64) {
	x, y := bezier.At(atT)
	xdc, ydc := bezier.AtDeCasteljau(atT)
	//fmt.Printf("bezier.AtDeCasteljau(%v) = (%v, %v) \n", atT, xdc, ydc)
	assert.InDeltaf(t, xdc, x, delta, "spline.At(%v).x = %v != spline.AtDeCasteljau(%v).x = %v", atT, x, atT, xdc)
	assert.InDeltaf(t, ydc, y, delta, "spline.At(%v).y = %v != spline.AtDeCasteljau(%v).y = %v", atT, y, atT, ydc)
}

// createBezierDiag00to11 creates a bezier representing a straight line from 0,0 to 1,1
func createBezierDiag00to11() *BezierSpline2d {
	return NewBezierSpline2d(
		bendit.NewUniformKnots(),
		NewBezierVertex2d(0, 0, 0, 0, 1./3, 1./3),
		NewBezierVertex2d(1, 1, 2./3, 2./3, 0, 0))
}

// createBezierDiag00to11 creates a bezier representing an S-formed slope from 0,0 to 1,1
func createBezierS00to11() *BezierSpline2d {
	return NewBezierSpline2d(
		bendit.NewUniformKnots(),
		NewBezierVertex2d(0, 0, 0, 0, 1, 0),
		NewBezierVertex2d(1, 1, 0, 1, 0, 0))
}

func TestBezierSpline2d_At(t *testing.T) {
	bezier := createBezierDiag00to11()
	AssertSplineAt(t, bezier, 0, 0, 0)
	AssertSplineAt(t, bezier, 0.25, 0.25, 0.25)
	AssertSplineAt(t, bezier, .5, .5, .5)
	AssertSplineAt(t, bezier, 0.75, 0.75, 0.75)
	AssertSplineAt(t, bezier, 1, 1, 1)

	// one vertex, domain with value 0 only
	bezier = NewBezierSpline2d(
		bendit.NewUniformKnots(),
		NewBezierVertex2d(1, 2, 0, 0, 0, 0))
	AssertSplineAt(t, bezier, 0, 1, 2)

	bezier = NewBezierSpline2d(
		bendit.NewKnots([]float64{0}),
		NewBezierVertex2d(1, 2, 0, 0, 0, 0))
	AssertSplineAt(t, bezier, 0, 1, 2)

	// empty domain
	bezier = NewBezierSpline2d(bendit.NewUniformKnots())
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

func TestBezierSpline2d_Approx(t *testing.T) {
	bezier := createBezierDiag00to11()
	lines := make([][4]float64, 0)
	bezier.Approx(0.1, NewDirectCollector2d(func(x0, y0, x3, y3 float64) {
		lines = append(lines, [4]float64{x0, y0, x3, y3})
	}))
	assert.Len(t, lines, 1, "approximated with one line")
	assert.InDeltaf(t, 0., lines[0][0], delta, "start point x=0")
	assert.InDeltaf(t, 0., lines[0][1], delta, "start point y=0")
	assert.InDeltaf(t, 1., lines[0][2], delta, "end point x=0")
	assert.InDeltaf(t, 1., lines[0][3], delta, "end point y=0")

	bezier = createBezierS00to11()
	lines = make([][4]float64, 0)
	bezier.Approx(0.3, NewDirectCollector2d(func(x0, y0, x3, y3 float64) {
		lines = append(lines, [4]float64{x0, y0, x3, y3})
	}))
	assert.Greater(t, len(lines), 1, "approximated with more than one line")
	incT := 1. / float64(len(lines)) // TODO get ts, te from collector
	for i, lin := range lines {
		atT := float64(i) * incT
		x, y := bezier.At(atT)
		assert.InDeltaf(t, lin[0], x, delta, "x value of start points of approximated lines are on bezier curve")
		assert.InDeltaf(t, lin[1], y, delta, "y value of start points of approximated lines are on bezier curve")
	}
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
