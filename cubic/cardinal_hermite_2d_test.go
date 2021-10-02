package cubic

import (
	bendit "github.com/walpod/bend-it"
	"math"
	"math/rand"
	"testing"
)

func createCardinalDiag00to11() *CardinalHermiteSpline2d {
	return NewCardinalHermiteSpline2d(nil, 0,
		NewHermiteVx2Raw(bendit.NewVec(0, 0)),
		NewHermiteVx2Raw(bendit.NewVec(1, 1)))
}

func createCardinalVase() *CardinalHermiteSpline2d {
	return NewCardinalHermiteSpline2d(
		nil, 0,
		NewHermiteVx2Raw(bendit.NewVec(-1, 1)),
		NewHermiteVx2Raw(bendit.NewVec(0, 0)),
		NewHermiteVx2Raw(bendit.NewVec(1, 1)))
}

func TestCardinalHermiteSpline_At(t *testing.T) {
	card := createCardinalDiag00to11()
	for i := 0; i < 100; i++ {
		card.SetTension(rand.Float64()*4 - 2)
		card.Prepare()
		AssertSplineAt(t, card, 0, bendit.NewVec(0, 0))
		AssertSplineAt(t, card, 0.5, bendit.NewVec(0.5, 0.5))
		AssertSplineAt(t, card, 1, bendit.NewVec(1, 1))
	}
	card.SetTension(-1)
	card.Prepare()
	AssertSplineAt(t, card, 0.25, bendit.NewVec(0.25, 0.25))
	AssertSplineAt(t, card, 0.75, bendit.NewVec(0.75, 0.75))

	card = createCardinalVase()
	for i := 0; i < 100; i++ {
		card.SetTension(rand.Float64()*4 - 2)
		card.Prepare()
		AssertSplineAt(t, card, 0, bendit.NewVec(-1, 1))
		AssertSplineAt(t, card, 1, bendit.NewVec(0, 0))
		AssertSplineAt(t, card, 2, bendit.NewVec(1, 1))
	}
	card.SetTension(1) // high tension: stretched to line segments
	card.Prepare()
	isOnLineSegment := func(v bendit.Vec) bool {
		return math.Abs(v[0])-math.Abs(v[0]) < delta
	}
	for i := 0; i < 100; i++ {
		AssertRandSplinePointProperty(t, card, isOnLineSegment, "cardinal point must be on line segment between Vase points")
	}
}
