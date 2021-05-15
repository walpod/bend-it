package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"testing"
)

func createCanonicalParabola00to11() *CanonicalSpline2d {
	return NewCanonicalSpline2d(bendit.NewUniformKnots(), NewCubic2d(
		NewCubicPoly(0, 1, 0, 0),
		NewCubicPoly(0, 0, 1, 0)))
}

func TestCanonicalSpline2d_At(t *testing.T) {
	cp := NewCubicPoly(0, 0, 1, 0)
	canon := NewCanonicalSpline2d(bendit.NewUniformKnots(), NewCubic2d(cp, cp))
	AssertSplineAt(t, canon, 0, 0, 0)
	AssertSplineAt(t, canon, 0.25, 0.25*0.25, 0.25*0.25)
	AssertSplineAt(t, canon, 0.5, 0.5*0.5, 0.5*0.5)
	AssertSplineAt(t, canon, 0.75, 0.75*0.75, 0.75*0.75)
	AssertSplineAt(t, canon, 1, 1, 1)

	// one vertex only
	canon = NewOneVertexCanonicalSpline2d(1, 2)
	AssertSplineAt(t, canon, 0, 1, 2)

	// empty domain
	canon = NewCanonicalSpline2d(bendit.NewUniformKnots())
	canon = NewCanonicalSpline2d(bendit.NewKnots([]float64{}))
}

func TestCanonicalSpline2d_Bezier(t *testing.T) {
	canon := createCanonicalParabola00to11()
	bezier := canon.Bezier()
	AssertSplinesEqual(t, canon, bezier, 0, 1, 100)
}

func TestCanonicalSpline2d_Approx(t *testing.T) {
	canon := createCanonicalParabola00to11()
	lc := NewLineToSliceCollector2d()
	canon.Approx(0.01, lc)
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	AssertApproxStartPointsMatchSpline(t, lc.Lines, canon)
}
