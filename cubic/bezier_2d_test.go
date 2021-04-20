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

func AssertSplineAtDeCasteljau(t *testing.T, bezier *BezierSpline2d, atT float64) {
	x, y := bezier.At(atT)
	xdc, ydc := bezier.AtDeCasteljau(atT)
	//fmt.Printf("bezier.AtDeCasteljau(%v) = (%v, %v) \n", atT, xdc, ydc)
	assert.InDeltaf(t, xdc, x, delta, "spline.At(%v).x = %v != spline.AtDeCasteljau(%v).x = %v", atT, x, atT, xdc)
	assert.InDeltaf(t, ydc, y, delta, "spline.At(%v).y = %v != spline.AtDeCasteljau(%v).y = %v", atT, y, atT, ydc)
}

func TestBezierSpline2d_At(t *testing.T) {
	spl := NewBezierSpline2d([]float64{0, 1}, []float64{0, 1},
		//[]float64{3, -2}, []float64{3, -2},
		[]float64{1, 0}, []float64{1, 0},
		nil)
	AssertSplineAt(t, spl, 0, 0, 0)
	//AssertSplineAt(t, spl, 0.25, 0.25, 0.25)
	AssertSplineAt(t, spl, .5, .5, .5)
	//AssertSplineAt(t, spl, 0.75, 0.75, 0.75)
	AssertSplineAt(t, spl, 1, 1, 1)
}

func TestBezierSpline2d_AtDeCasteljau(t *testing.T) {
	bezier := NewBezierSpline2d([]float64{0, 1}, []float64{0, 1},
		[]float64{1, 0}, []float64{0, 1},
		nil)
	AssertSplineAtDeCasteljau(t, bezier, 0)
	AssertSplineAtDeCasteljau(t, bezier, 0.1)
	AssertSplineAtDeCasteljau(t, bezier, 0.25)
	AssertSplineAtDeCasteljau(t, bezier, 0.5)
	AssertSplineAtDeCasteljau(t, bezier, 0.75)
	AssertSplineAtDeCasteljau(t, bezier, 0.9)
	AssertSplineAtDeCasteljau(t, bezier, 1)
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
