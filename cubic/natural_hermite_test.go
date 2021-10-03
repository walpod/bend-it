package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"math/rand"
	"testing"
)

func createNaturalDiag00to11() *NaturalHermiteSpline2d {
	return NewNaturalHermiteSpline2d(nil,
		NewRawHermiteVertex(bendit.NewVec(0, 0)),
		NewRawHermiteVertex(bendit.NewVec(1, 1)))
}

func createNaturalVase() *NaturalHermiteSpline2d {
	return NewNaturalHermiteSpline2d(nil,
		NewRawHermiteVertex(bendit.NewVec(-1, 1)),
		NewRawHermiteVertex(bendit.NewVec(0, 0)),
		NewRawHermiteVertex(bendit.NewVec(1, 1)))
}

func TestNaturalHermiteSpline_At(t *testing.T) {
	nat := createNaturalDiag00to11()
	nat.Prepare()
	AssertSplineAt(t, nat, 0, bendit.NewVec(0, 0))
	AssertSplineAt(t, nat, 0.25, bendit.NewVec(0.25, 0.25))
	AssertSplineAt(t, nat, 0.5, bendit.NewVec(0.5, 0.5))
	AssertSplineAt(t, nat, 0.75, bendit.NewVec(0.75, 0.75))
	AssertSplineAt(t, nat, 1, bendit.NewVec(1, 1))

	nat = createNaturalVase()
	nat.Prepare()
	AssertSplineAt(t, nat, 0, bendit.NewVec(-1, 1))
	AssertSplineAt(t, nat, 1, bendit.NewVec(0, 0))
	AssertSplineAt(t, nat, 2, bendit.NewVec(1, 1))

	ts, te := nat.knots.Tstart(), nat.knots.Tend()
	for i := 0; i < 100; i++ {
		atT := ts + rand.Float64()*(te-ts)
		v := nat.At(atT)
		assert.True(t, v[0] >= -1 && v[0] <= 1, "natural point[0] must be in range -1..1")
		assert.True(t, v[1] >= 0 && v[1] <= 1, "natural point[1] must be in range -1..1")
	}
}
