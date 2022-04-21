package cubic

import (
	"github.com/walpod/bendigo"
	"math"
	"math/rand"
	"testing"
)

func createCardinalDiag00to11() *CardinalVertBuilder {
	return NewCardinalVertBuilder(nil, 0,
		NewRawHermiteVertex(bendigo.NewVec(0, 0)),
		NewRawHermiteVertex(bendigo.NewVec(1, 1)))
}

func createCardinalVase() *CardinalVertBuilder {
	return NewCardinalVertBuilder(
		nil, 0,
		NewRawHermiteVertex(bendigo.NewVec(-1, 1)),
		NewRawHermiteVertex(bendigo.NewVec(0, 0)),
		NewRawHermiteVertex(bendigo.NewVec(1, 1)))
}

func TestCardinalSpline(t *testing.T) {
	cardBuilder := createCardinalDiag00to11()
	for i := 0; i < 100; i++ {
		cardBuilder.SetTension(rand.Float64()*4 - 2)
		card := cardBuilder.Spline()
		AssertSplineAt(t, card, 0, bendigo.NewVec(0, 0))
		AssertSplineAt(t, card, 0.5, bendigo.NewVec(0.5, 0.5))
		AssertSplineAt(t, card, 1, bendigo.NewVec(1, 1))
	}
	cardBuilder.SetTension(-1)
	card := cardBuilder.Spline()
	AssertSplineAt(t, card, 0.25, bendigo.NewVec(0.25, 0.25))
	AssertSplineAt(t, card, 0.75, bendigo.NewVec(0.75, 0.75))

	cardBuilder = createCardinalVase()
	for i := 0; i < 100; i++ {
		cardBuilder.SetTension(rand.Float64()*4 - 2)
		card := cardBuilder.Spline()
		AssertSplineAt(t, card, 0, bendigo.NewVec(-1, 1))
		AssertSplineAt(t, card, 1, bendigo.NewVec(0, 0))
		AssertSplineAt(t, card, 2, bendigo.NewVec(1, 1))
	}

	// high tension: stretched to line segments
	cardBuilder.SetTension(1)
	isOnLineSegment := func(v bendigo.Vec) bool {
		return math.Abs(v[0])-math.Abs(v[0]) < delta
	}
	for i := 0; i < 100; i++ {
		AssertRandSplinePointProperty(t, cardBuilder.Spline(), isOnLineSegment, "cardinal point must be on line segment between Vase points")
	}
}
