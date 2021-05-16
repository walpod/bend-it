package cubic

import (
	bendit "github.com/walpod/bend-it"
	"testing"
)

func createHermDiag00to11() *HermiteSpline2d {
	return NewHermiteSpline2d(bendit.NewUniformKnots(),
		[]float64{0, 1}, []float64{0, 1},
		[]float64{0, 1}, []float64{0, 1},
		[]float64{1, 0}, []float64{1, 0},
	)
}

func createHermParabola00to11() *HermiteSpline2d {
	return NewHermiteSpline2d(bendit.NewUniformKnots(),
		[]float64{0, 1}, []float64{0, 1},
		[]float64{0, 1}, []float64{0, 2},
		[]float64{1, 0}, []float64{0, 0},
	)
}

func createDoubleHermParabola00to11to22() *HermiteSpline2d {
	return NewHermiteSpline2d(bendit.NewUniformKnots(),
		[]float64{0, 1, 2}, []float64{0, 1, 2},
		[]float64{0, 1, 1}, []float64{0, 2, 2},
		[]float64{1, 1, 0}, []float64{0, 0, 0},
	)
}

func TestHermiteSpline2d_At(t *testing.T) {
	herm := createHermDiag00to11()
	AssertSplineAt(t, herm, 0, 0, 0)
	AssertSplineAt(t, herm, 0.25, 0.25, 0.25)
	AssertSplineAt(t, herm, .5, .5, .5)
	AssertSplineAt(t, herm, 0.75, 0.75, 0.75)
	AssertSplineAt(t, herm, 1, 1, 1)

	herm = createHermParabola00to11()
	AssertSplineAt(t, herm, 0, 0, 0)
	AssertSplineAt(t, herm, 0.25, 0.25, 0.25*0.25)
	AssertSplineAt(t, herm, 0.5, 0.5, 0.25)
	AssertSplineAt(t, herm, 0.75, 0.75, 0.75*0.75)
	AssertSplineAt(t, herm, 1, 1, 1)

	herm = createDoubleHermParabola00to11to22()
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
	herm = NewHermiteSpline2d(bendit.NewUniformKnots(), []float64{1}, []float64{2}, []float64{0}, []float64{0}, []float64{0}, []float64{0})
	AssertSplineAt(t, herm, 0, 1, 2)

	// empty domain
	herm = NewHermiteSpline2d(bendit.NewUniformKnots(), []float64{}, []float64{}, []float64{}, []float64{}, []float64{}, []float64{})
}
