package cubic

import (
	bendit "github.com/walpod/bend-it"
	"testing"
)

func TestHermiteSpline2d_At(t *testing.T) {
	spl := NewHermiteSpline2d([]float64{0, 1}, []float64{0, 1},
		[]float64{0, 1}, []float64{0, 1},
		[]float64{1, 0}, []float64{1, 0},
		bendit.NewUniformKnots())
	AssertSplineAt(t, spl, 0, 0, 0)
	AssertSplineAt(t, spl, 0.25, 0.25, 0.25)
	AssertSplineAt(t, spl, .5, .5, .5)
	AssertSplineAt(t, spl, 0.75, 0.75, 0.75)
	AssertSplineAt(t, spl, 1, 1, 1)
}
