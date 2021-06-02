package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"math/rand"
	"testing"
)

func createNaturalDiag00to11() *NaturalHermiteSpline2d {
	return NewNaturalHermiteSpline2d(bendit.NewUniformKnots(),
		NewHermiteVx2Raw(0, 0),
		NewHermiteVx2Raw(1, 1))
}

func createNaturalVase() *NaturalHermiteSpline2d {
	return NewNaturalHermiteSpline2d(
		bendit.NewUniformKnots(),
		NewHermiteVx2Raw(-1, 1),
		NewHermiteVx2Raw(0, 0),
		NewHermiteVx2Raw(1, 1))
}

func TestNaturalHermiteSpline_At(t *testing.T) {
	nat := createNaturalDiag00to11()
	AssertSplineAt(t, nat, 0, 0, 0)
	AssertSplineAt(t, nat, 0.25, 0.25, 0.25)
	AssertSplineAt(t, nat, 0.5, 0.5, 0.5)
	AssertSplineAt(t, nat, 0.75, 0.75, 0.75)
	AssertSplineAt(t, nat, 1, 1, 1)

	nat = createNaturalVase()
	AssertSplineAt(t, nat, 0, -1, 1)
	AssertSplineAt(t, nat, 1, 0, 0)
	AssertSplineAt(t, nat, 2, 1, 1)

	domain := nat.Knots().Domain()
	for i := 0; i < 100; i++ {
		atT := domain.Start + rand.Float64()*(domain.End-domain.Start)
		x, y := nat.At(atT)
		assert.True(t, x >= -1 && x <= 1, "natural point.x must be in range -1..1")
		assert.True(t, y >= 0 && y <= 1, "natural point.x must be in range -1..1")
	}
}
