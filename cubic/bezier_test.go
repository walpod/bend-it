package cubic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/walpod/bendigo"
	"math/rand"
	"testing"
)

// START some general bendigo spline Asserts ... TODO to be moved ...
const delta = 0.0000000001

func AssertVecInDelta(t *testing.T, expected bendigo.Vec, actual bendigo.Vec, msg string) {
	assert.Equal(t, expected.Dim(), actual.Dim(), "dimension of expected = %v != dimension of actual %v", expected.Dim(), actual.Dim())
	for d := 0; d < expected.Dim(); d++ {
		assert.InDeltaf(t, expected[d], actual[d], delta, msg+", at dim = %v", d)
	}
}

func AssertSplineAt(t *testing.T, spline bendigo.Spline, atT float64, expected bendigo.Vec) {
	actual := spline.At(atT)
	AssertVecInDelta(t, expected, actual, fmt.Sprintf("spline0.At(%v) = %v != spline1.At(%v) = %v", atT, expected, atT, actual))
}

func AssertSplinesEqualInRange(t *testing.T, spline0 bendigo.Spline, spline1 bendigo.Spline, tstart, tend float64, sampleCnt int) {
	for i := 0; i < sampleCnt; i++ {
		atT := rand.Float64()*(tend-tstart) + tstart
		v0 := spline0.At(atT)
		v1 := spline1.At(atT)
		AssertVecInDelta(t, v0, v1, fmt.Sprintf("spline0.At(%v).x = %v != spline1.At(%v).x = %v", atT, v0, atT, v1))
	}
}

func AssertSplinesEqual(t *testing.T, spline0 bendigo.Spline, spline1 bendigo.Spline, sampleCnt int) {
	// assert over full domain
	AssertSplinesEqualInRange(t, spline0, spline1, spline0.Knots().Tstart(), spline0.Knots().Tend(), sampleCnt)
}

func AssertApproxStartPointsMatchSpline(t *testing.T, lines []bendigo.LineParams, spline bendigo.Spline) {
	for _, lin := range lines {
		v := spline.At(lin.Tstart)
		AssertVecInDelta(t, v, lin.Pstart, fmt.Sprintf("spline.At(%v) = %v != start-point = %v of approximated line", lin.Tstart, v, lin.Pstart))
		//assert.InDeltaf(t, x, lin.Pstartx, delta, "spline.At(%v).x = %v != start-point.x = %v of approximated line", lin.Tstart, x, lin.Pstartx)
	}
}

func AssertRandSplinePointProperty(t *testing.T, spline bendigo.Spline, hasProp func(v bendigo.Vec) bool, msg string) {
	ts, te := spline.Knots().Tstart(), spline.Knots().Tend()
	atT := ts + rand.Float64()*(te-ts)
	v := spline.At(atT)
	assert.True(t, hasProp(v), msg)
}

// END some general bendigo spline Asserts

// createBezierDiag00to11 creates a bezier representing a straight line from (0,0) to (1,1)
func createBezierDiag00to11() *BezierVertBuilder {
	return NewBezierVertBuilder(nil,
		NewBezierVertex(bendigo.NewVec(0, 0), nil, bendigo.NewVec(1./3, 1./3)),
		NewBezierVertex(bendigo.NewVec(1, 1), bendigo.NewVec(2./3, 2./3), nil),
	)
}

// createBezierDiag00to11 creates a bezier representing an S-formed slope from (0,0) to (1,1)
func createBezierS00to11() *BezierVertBuilder {
	return NewBezierVertBuilder(nil,
		NewBezierVertex(bendigo.NewVec(0, 0), nil, bendigo.NewVec(1, 0)),
		NewBezierVertex(bendigo.NewVec(1, 1), bendigo.NewVec(0, 1), nil),
	)
}

// createBezierDiag00to11 creates two consecutive beziers representing an S-formed slope from (0,0) to (1,1) or (1,1) to (2,2), resp.
func createDoubleBezierS00to11to22() *BezierVertBuilder {
	return NewBezierVertBuilder(nil,
		NewBezierVertex(bendigo.NewVec(0, 0), nil, bendigo.NewVec(1, 0)),
		NewBezierVertex(bendigo.NewVec(1, 1) /*bendigo.NewVec(0, 1)*/, nil, bendigo.NewVec(2, 1)),
		NewBezierVertex(bendigo.NewVec(2, 2), bendigo.NewVec(1, 2), nil),
	)
}

func TestBezierSpline_At(t *testing.T) {
	bezier := createBezierDiag00to11().Build()
	AssertSplineAt(t, bezier, 0, bendigo.NewVec(0, 0))
	AssertSplineAt(t, bezier, 0.25, bendigo.NewVec(0.25, 0.25))
	AssertSplineAt(t, bezier, .5, bendigo.NewVec(.5, .5))
	AssertSplineAt(t, bezier, 0.75, bendigo.NewVec(0.75, 0.75))
	AssertSplineAt(t, bezier, 1, bendigo.NewVec(1, 1))

	bezier = createDoubleBezierS00to11to22().Build()
	AssertSplineAt(t, bezier, 0, bendigo.NewVec(0, 0))
	AssertSplineAt(t, bezier, 0.5, bendigo.NewVec(0.5, 0.5))
	AssertSplineAt(t, bezier, 1, bendigo.NewVec(1, 1))
	AssertSplineAt(t, bezier, 1.5, bendigo.NewVec(1.5, 1.5))
	AssertSplineAt(t, bezier, 2, bendigo.NewVec(2, 2))

	// single vertex, domain with value 0 only
	bezier = NewBezierVertBuilder(nil,
		NewBezierVertex(bendigo.NewVec(1, 2), nil, nil)).
		Build()
	AssertSplineAt(t, bezier, 0, bendigo.NewVec(1, 2))

	bezier = NewBezierVertBuilder(
		[]float64{0},
		NewBezierVertex(bendigo.NewVec(1, 2), nil, nil)).
		Build()
	AssertSplineAt(t, bezier, 0, bendigo.NewVec(1, 2))

	// empty domain
	bezier = NewBezierVertBuilder(nil).Build()
	bezier = NewBezierVertBuilder([]float64{}).Build()
}

func TestDeCasteljauSpline_At(t *testing.T) {
	bezierBuilder := createBezierS00to11()
	AssertSplinesEqual(t, bezierBuilder.Build(), bezierBuilder.BuildDeCasteljau(), 100)
}

func TestBezierApproxer_Approx(t *testing.T) {
	bezierApproxer := createBezierDiag00to11().BuildApproxer()
	lc := bendigo.NewLineToSliceCollector()
	bendigo.ApproxAll(bezierApproxer, 0.1, lc)
	assert.Len(t, lc.Lines, 1, "approximated with one line")
	AssertVecInDelta(t, bendigo.NewVec(0, 0), lc.Lines[0].Pstart, "start point = [0,0]")
	AssertVecInDelta(t, bendigo.NewVec(1, 1), lc.Lines[0].Pend, "end point = [1,1]")

	// start points of approximated lines must be on bezier curve and match bezier.At
	bezierBuilder := createBezierS00to11()
	lc = bendigo.NewLineToSliceCollector()
	bendigo.ApproxAll(bezierBuilder.BuildApproxer(), 0.02, lc)
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	AssertApproxStartPointsMatchSpline(t, lc.Lines, bezierBuilder.Build())
}

func TestBezierVertBuilder_AddVertex(t *testing.T) {
	bezierBuilder := createBezierDiag00to11()
	err := bezierBuilder.AddVertex(3, nil)
	assert.NotNil(t, err, "knot-no. too large")
	err = bezierBuilder.AddVertex(2, NewBezierVertex(bendigo.NewVec(2, 2), bendigo.NewVec(1.5, 1.5), nil))
	assert.Equal(t, bezierBuilder.knots.KnotCnt(), 3, "knot-cnt %v wrong", bezierBuilder.knots.KnotCnt())
	err = bezierBuilder.AddVertex(0, NewBezierVertex(bendigo.NewVec(-1, -1), bendigo.NewVec(-2, -2), nil))
	assert.Equal(t, bezierBuilder.knots.KnotCnt(), 4, "knot-cnt %v wrong", bezierBuilder.knots.KnotCnt())
	assert.Equal(t, bezierBuilder.Vertex(1), createBezierDiag00to11().Vertex(0), "vertices don't match")
	assert.Equal(t, bezierBuilder.Vertex(2), createBezierDiag00to11().Vertex(1), "vertices don't match")
}

func TestBezierVertBuilder_DeleteVertex(t *testing.T) {
	bezierBuilder := createBezierDiag00to11()
	err := bezierBuilder.DeleteVertex(2)
	assert.NotNil(t, err, "knot-no. doesn't exist")
	err = bezierBuilder.DeleteVertex(1)
	assert.Equal(t, bezierBuilder.knots.KnotCnt(), 1, "knot-cnt %v wrong", bezierBuilder.knots.KnotCnt())
	assert.Equal(t, bezierBuilder.Vertex(0), createBezierDiag00to11().Vertex(0), "vertices don't match")
	err = bezierBuilder.DeleteVertex(0)
	assert.Equal(t, bezierBuilder.knots.KnotCnt(), 0, "knot-cnt %v wrong", bezierBuilder.knots.KnotCnt())
}
