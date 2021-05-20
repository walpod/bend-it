package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"math"
	"testing"
)

func createHermDiag00to11() *HermiteSpline2d {
	return NewHermiteSpline2d(bendit.NewUniformKnots(),
		NewHermiteVertex2d(0, 0, 0, 0, 1, 1),
		NewHermiteVertex2d(1, 1, 1, 1, 0, 0),
	)
}

func createNonUniHermDiag00to11() *HermiteSpline2d {
	return NewHermiteSpline2d(bendit.NewKnots([]float64{0, math.Sqrt2}),
		NewHermiteVertex2d(0, 0, 0, 0, 1, 1),
		NewHermiteVertex2d(1, 1, 1, 1, 0, 0),
	)
}

func isOnDiag(x, y float64) bool {
	return math.Abs(x-y) < delta
}

func createHermParabola00to11(uniform bool) *HermiteSpline2d {
	var knots *bendit.Knots
	if uniform {
		knots = bendit.NewUniformKnots()
	} else {
		knots = bendit.NewKnots([]float64{0, 1}) // is in fact uniform but specified as non-uniform
	}
	return NewHermiteSpline2d(knots,
		NewHermiteVertex2d(0, 0, 0, 0, 1, 0),
		NewHermiteVertex2d(1, 1, 1, 2, 0, 0),
	)
}

func createDoubleHermParabola00to11to22(uniform bool) *HermiteSpline2d {
	var knots *bendit.Knots
	if uniform {
		knots = bendit.NewUniformKnots()
	} else {
		knots = bendit.NewKnots([]float64{0, 1, 2}) // is in fact uniform but specified as non-uniform
	}
	return NewHermiteSpline2d(knots,
		NewHermiteVertex2d(0, 0, 0, 0, 1, 0),
		NewHermiteVertex2d(1, 1, 1, 2, 1, 0),
		NewHermiteVertex2d(2, 2, 1, 2, 0, 0),
	)
}

func TestHermiteSpline2d_At(t *testing.T) {
	herm := createHermDiag00to11()
	AssertSplineAt(t, herm, 0, 0, 0)
	AssertSplineAt(t, herm, 0.25, 0.25, 0.25)
	AssertSplineAt(t, herm, .5, .5, .5)
	AssertSplineAt(t, herm, 0.75, 0.75, 0.75)
	AssertSplineAt(t, herm, 1, 1, 1)

	herm = createNonUniHermDiag00to11()
	domain := herm.knots.Domain(herm.SegmentCnt())
	AssertSplineAt(t, herm, domain.Start, 0, 0)
	AssertSplineAt(t, herm, domain.End/2, .5, .5)
	AssertSplineAt(t, herm, domain.End, 1, 1)
	for i := 0; i < 100; i++ {
		AssertRandSplinePointProperty(t, herm, isOnDiag, "hermite point must be on diagonal")
	}

	herm = createHermParabola00to11(true)
	AssertSplineAt(t, herm, 0, 0, 0)
	AssertSplineAt(t, herm, 0.25, 0.25, 0.25*0.25)
	AssertSplineAt(t, herm, 0.5, 0.5, 0.25)
	AssertSplineAt(t, herm, 0.75, 0.75, 0.75*0.75)
	AssertSplineAt(t, herm, 1, 1, 1)

	herm = createDoubleHermParabola00to11to22(true)
	AssertSplineAt(t, herm, 0, 0, 0)
	AssertSplineAt(t, herm, 0.25, 0.25, 0.25*0.25)
	AssertSplineAt(t, herm, 0.5, 0.5, 0.25)
	AssertSplineAt(t, herm, 0.75, 0.75, 0.75*0.75)
	AssertSplineAt(t, herm, 1, 1, 1)
	AssertSplineAt(t, herm, 1.25, 1.25, 1+0.25*0.25)
	AssertSplineAt(t, herm, 1.5, 1.5, 1.25)
	AssertSplineAt(t, herm, 1.75, 1.75, 1+0.75*0.75)
	AssertSplineAt(t, herm, 2, 2, 2)

	// domain with ony one value: 0
	herm = NewHermiteSpline2d(bendit.NewUniformKnots(), NewHermiteVertex2d(1, 2, 0, 0, 0, 0))
	AssertSplineAt(t, herm, 0, 1, 2)

	// empty domain
	herm = NewHermiteSpline2d(bendit.NewUniformKnots())

	// uniform and regular non-uniform must match
	herm = createHermParabola00to11(true)
	nuherm := createHermParabola00to11(false)
	AssertSplinesEqual(t, herm, nuherm, 100)

	herm = createDoubleHermParabola00to11to22(true)
	nuherm = createDoubleHermParabola00to11to22(false)
	AssertSplinesEqual(t, herm, nuherm, 100)
}

func TestHermiteSpline2d_Canonical(t *testing.T) {
	herm := createDoubleHermParabola00to11to22(true)
	AssertSplinesEqual(t, herm, herm.Canonical(), 100)

	herm = createDoubleHermParabola00to11to22(false)
	AssertSplinesEqual(t, herm, herm.Canonical(), 100)
}

func TestHermiteSpline2d_Approx(t *testing.T) {
	herm := createDoubleHermParabola00to11to22(true)
	lc := NewLineToSliceCollector2d()
	herm.Approx(0.02, lc)
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	assert.InDeltaf(t, 0., lc.Lines[0].Sx, delta, "start point x=0")
	assert.InDeltaf(t, 0., lc.Lines[0].Sy, delta, "start point y=0")
	assert.InDeltaf(t, 2., lc.Lines[len(lc.Lines)-1].Ex, delta, "end point x=0")
	assert.InDeltaf(t, 2., lc.Lines[len(lc.Lines)-1].Ey, delta, "end point y=0")
	// start points of approximated lines must be on bezier curve and match bezier.At
	AssertApproxStartPointsMatchSpline(t, lc.Lines, herm)
}
