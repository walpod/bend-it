package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"testing"
)

func createCanonLineWithParabolicParam() *CanonicalSpline2d {
	cp := NewCubicPoly(0, 0, 1, 0)
	canon := NewCanonicalSpline2d(nil, NewCubic2d(cp, cp))
	return canon
}

func createCanonParabola00to11() *CanonicalSpline2d {
	return NewCanonicalSpline2d(nil,
		NewCubic2d(NewCubicPoly(0, 1, 0, 0), NewCubicPoly(0, 0, 1, 0)),
	)
}

func createDoubleCanonParabola00to11to22() *CanonicalSpline2d {
	return NewCanonicalSpline2d(nil,
		NewCubic2d(NewCubicPoly(0, 1, 0, 0), NewCubicPoly(0, 0, 1, 0)),
		NewCubic2d(NewCubicPoly(1, 1, 0, 0), NewCubicPoly(1, 0, 1, 0)),
	)
}

func TestCanonicalSpline2d_At(t *testing.T) {
	canon := createCanonLineWithParabolicParam()
	AssertSplineAt(t, canon, 0, 0, 0)
	AssertSplineAt(t, canon, 0.25, 0.25*0.25, 0.25*0.25)
	AssertSplineAt(t, canon, 0.5, 0.5*0.5, 0.5*0.5)
	AssertSplineAt(t, canon, 0.75, 0.75*0.75, 0.75*0.75)
	AssertSplineAt(t, canon, 1, 1, 1)

	canon = createCanonParabola00to11()
	AssertSplineAt(t, canon, 0, 0, 0)
	AssertSplineAt(t, canon, 0.5, 0.5, 0.5*0.5)
	AssertSplineAt(t, canon, 1, 1, 1)

	canon = createDoubleCanonParabola00to11to22()
	AssertSplineAt(t, canon, 0, 0, 0)
	AssertSplineAt(t, canon, 0.5, 0.5, 0.5*0.5)
	AssertSplineAt(t, canon, 1, 1, 1)
	AssertSplineAt(t, canon, 1.5, 1.5, 1+0.5*0.5)
	AssertSplineAt(t, canon, 2, 2, 2)

	// single vertex
	canon = NewSingleVxCanonicalSpline2d(1, 2)
	AssertSplineAt(t, canon, 0, 1, 2)

	// empty domain
	canon = NewCanonicalSpline2d(nil)
	canon = NewCanonicalSpline2d(nil)
	canon = NewCanonicalSpline2d(bendit.NewNonUniformKnots([]float64{}))
}

func TestCanonicalSpline2d_Bezier(t *testing.T) {
	canon := createCanonParabola00to11()
	AssertSplinesEqual(t, canon, canon.Bezier(), 100)

	canon = createDoubleCanonParabola00to11to22()
	AssertSplinesEqual(t, canon, canon.Bezier(), 100)
}

func TestCanonicalSpline2d_Approx(t *testing.T) {
	canon := createCanonParabola00to11()
	lc := NewLineToSliceCollector2d()
	canon.Approx(0.01, lc)
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	AssertApproxStartPointsMatchSpline(t, lc.Lines, canon)
}
