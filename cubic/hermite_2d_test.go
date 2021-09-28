package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"math"
	"testing"
)

func createHermDiag00to11() *HermiteSpline2d {
	return NewHermiteSpline2d(nil,
		NewHermiteVx2(0, 0, NewControl(0, 0), NewControl(1, 1)),
		NewHermiteVx2(1, 1, NewControl(1, 1), NewControl(0, 0)),
	)
}

func createNonUniHermDiag00to11() *HermiteSpline2d {
	return NewHermiteSpline2d([]float64{0, math.Sqrt2},
		NewHermiteVx2(0, 0, NewControl(0, 0), NewControl(1, 1)),
		NewHermiteVx2(1, 1, NewControl(1, 1), NewControl(0, 0)),
	)
}

func isOnDiag(x, y float64) bool {
	return math.Abs(x-y) < delta
}

func createHermParabola00to11(uniform bool) *HermiteSpline2d {
	var tknots []float64
	if uniform {
		tknots = nil
	} else {
		tknots = []float64{0, 1} // is in fact uniform but specified as non-uniform
	}
	return NewHermiteSpline2d(tknots,
		NewHermiteVx2(0, 0, NewControl(0, 0), NewControl(1, 0)),
		NewHermiteVx2(1, 1, NewControl(1, 2), NewControl(0, 0)),
	)
}

func createDoubleHermParabola00to11to22(uniform bool) *HermiteSpline2d {
	var tknots []float64
	if uniform {
		tknots = nil
	} else {
		tknots = []float64{0, 1, 2} // is in fact uniform but specified as non-uniform
	}
	return NewHermiteSpline2d(tknots,
		NewHermiteVx2(0, 0, NewControl(0, 0), NewControl(1, 0)),
		NewHermiteVx2(1, 1, NewControl(1, 2), NewControl(1, 0)),
		NewHermiteVx2(2, 2, NewControl(1, 2), NewControl(0, 0)),
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
	ts, te := herm.knots.Tstart(), herm.knots.Tend()
	AssertSplineAt(t, herm, ts, 0, 0)
	AssertSplineAt(t, herm, te/2, .5, .5)
	AssertSplineAt(t, herm, te, 1, 1)
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
	herm = NewHermiteSpline2d(nil,
		NewHermiteVx2(1, 2, NewControl(0, 0), NewControl(0, 0)))
	AssertSplineAt(t, herm, 0, 1, 2)

	// empty domain
	herm = NewHermiteSpline2d(nil)

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
	lc := bendit.NewLineToSliceCollector2d()
	bendit.ApproxAll(herm, 0.02, lc)
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	assert.InDeltaf(t, 0., lc.Lines[0].Pstartx, delta, "start point x=0")
	assert.InDeltaf(t, 0., lc.Lines[0].Pstarty, delta, "start point y=0")
	assert.InDeltaf(t, 2., lc.Lines[len(lc.Lines)-1].Pendx, delta, "end point x=0")
	assert.InDeltaf(t, 2., lc.Lines[len(lc.Lines)-1].Pendy, delta, "end point y=0")
	// start points of approximated lines must be on bezier curve and match bezier.At
	AssertApproxStartPointsMatchSpline(t, lc.Lines, herm)
}

func TestHermiteSpline2d_AddVertex(t *testing.T) {
	hermite := createHermDiag00to11()
	err := hermite.AddVertex(3, nil)
	assert.NotNil(t, err, "knot-no. too large")
	err = hermite.AddVertex(2, NewHermiteVx2(2, 2, NewControl(1.5, 1.5), nil))
	assert.Equal(t, hermite.knots.KnotCnt(), 3, "knot-cnt %v wrong", hermite.knots.KnotCnt())
	err = hermite.AddVertex(0, NewHermiteVx2(-1, -1, NewControl(-2, -2), nil))
	assert.Equal(t, hermite.knots.KnotCnt(), 4, "knot-cnt %v wrong", hermite.knots.KnotCnt())
	assert.Equal(t, hermite.Vertex(1), createHermDiag00to11().Vertex(0), "vertices don't match")
	assert.Equal(t, hermite.Vertex(2), createHermDiag00to11().Vertex(1), "vertices don't match")
}

func TestHermiteSpline2d_DeleteVertex(t *testing.T) {
	hermite := createHermDiag00to11()
	err := hermite.DeleteVertex(2)
	assert.NotNil(t, err, "knot-no. doesn't exist")
	err = hermite.DeleteVertex(1)
	assert.Equal(t, hermite.knots.KnotCnt(), 1, "knot-cnt %v wrong", hermite.knots.KnotCnt())
	assert.Equal(t, hermite.Vertex(0), createHermDiag00to11().Vertex(0), "vertices don't match")
	err = hermite.DeleteVertex(0)
	assert.Equal(t, hermite.knots.KnotCnt(), 0, "knot-cnt %v wrong", hermite.knots.KnotCnt())
}
