package cubic

import (
	"github.com/stretchr/testify/assert"
	bendit "github.com/walpod/bend-it"
	"math"
	"math/rand"
	"testing"
)

func createCardinalDiag00to11() *CardinalHermiteSpline2d {
	return NewCardinalHermiteSpline2d(bendit.NewUniformKnots(), 0,
		NewRawHermiteVertex2d(0, 0),
		NewRawHermiteVertex2d(1, 1))
}

func createCardinalVase() *CardinalHermiteSpline2d {
	return NewCardinalHermiteSpline2d(
		bendit.NewUniformKnots(), 0,
		NewRawHermiteVertex2d(-1, 1),
		NewRawHermiteVertex2d(0, 0),
		NewRawHermiteVertex2d(1, 1))
}

func TestCardinalHermiteSpline_At(t *testing.T) {
	card := createCardinalDiag00to11()
	for i := 0; i < 100; i++ {
		card.SetTension(rand.Float64()*4 - 2)
		AssertSplineAt(t, card, 0, 0, 0)
		AssertSplineAt(t, card, 0.5, 0.5, 0.5)
		AssertSplineAt(t, card, 1, 1, 1)
	}
	card.SetTension(-1)
	AssertSplineAt(t, card, 0.25, 0.25, 0.25)
	AssertSplineAt(t, card, 0.75, 0.75, 0.75)

	card = createCardinalVase()
	for i := 0; i < 100; i++ {
		card.SetTension(rand.Float64()*4 - 2)
		AssertSplineAt(t, card, 0, -1, 1)
		AssertSplineAt(t, card, 1, 0, 0)
		AssertSplineAt(t, card, 2, 1, 1)
	}
	card.SetTension(1) // high tension: stretched to line segments
	domain := card.Knots().Domain(2)
	for i := 0; i < 100; i++ {
		atT := domain.Start + rand.Float64()*(domain.End-domain.Start)
		x, y := card.At(atT)
		assert.InDelta(t, math.Abs(x), math.Abs(y), delta, "cardinal point must be on line segment between Vase points")
	}
}
