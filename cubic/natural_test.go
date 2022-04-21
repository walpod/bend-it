package cubic

import (
	"github.com/stretchr/testify/assert"
	"github.com/walpod/bendigo"
	"math/rand"
	"testing"
)

func createNaturalDiag00to11() *NaturalVertBuilder {
	return NewNaturalVertBuilder(nil,
		NewRawHermiteVertex(bendigo.NewVec(0, 0)),
		NewRawHermiteVertex(bendigo.NewVec(1, 1)))
}

func createNaturalVase() *NaturalVertBuilder {
	return NewNaturalVertBuilder(nil,
		NewRawHermiteVertex(bendigo.NewVec(-1, 1)),
		NewRawHermiteVertex(bendigo.NewVec(0, 0)),
		NewRawHermiteVertex(bendigo.NewVec(1, 1)))
}

func TestNaturalHermiteSpline(t *testing.T) {
	nat := createNaturalDiag00to11().Spline()
	AssertSplineAt(t, nat, 0, bendigo.NewVec(0, 0))
	AssertSplineAt(t, nat, 0.25, bendigo.NewVec(0.25, 0.25))
	AssertSplineAt(t, nat, 0.5, bendigo.NewVec(0.5, 0.5))
	AssertSplineAt(t, nat, 0.75, bendigo.NewVec(0.75, 0.75))
	AssertSplineAt(t, nat, 1, bendigo.NewVec(1, 1))

	nat = createNaturalVase().Spline()
	AssertSplineAt(t, nat, 0, bendigo.NewVec(-1, 1))
	AssertSplineAt(t, nat, 1, bendigo.NewVec(0, 0))
	AssertSplineAt(t, nat, 2, bendigo.NewVec(1, 1))

	ts, te := nat.Knots().Tstart(), nat.Knots().Tend()
	for i := 0; i < 100; i++ {
		atT := ts + rand.Float64()*(te-ts)
		v := nat.At(atT)
		assert.True(t, v[0] >= -1 && v[0] <= 1, "natural point[0] must be in range -1..1")
		assert.True(t, v[1] >= 0 && v[1] <= 1, "natural point[1] must be in range -1..1")
	}
}
