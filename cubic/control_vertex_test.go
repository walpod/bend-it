package cubic

import (
	bendit "github.com/walpod/bend-it"
	"testing"
)

func TestBezierVertexDependent(t *testing.T) {
	bvx := NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(1, 2), nil)
	AssertVecInDelta(t, bvx.entry.Negate(), bvx.exit, "dependent control must be reflected by origin [0,0]")
	bvx = NewBezierVertex(bendit.NewVec(0, 0), nil, bendit.NewVec(3, -5))
	AssertVecInDelta(t, bvx.entry, bvx.exit.Negate(), "dependent control must be reflected by origin [0,0]")
}

func TestBezierVertex_Translate(t *testing.T) {
	bvx := NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), nil).Translate(bendit.NewVec(2, 0))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendit.NewVec(2, 0), bendit.NewVec(2, 1), bendit.NewVec(2, -1)), true, bvx)
	bvx = NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 2), bendit.NewVec(3, 0)).Translate(bendit.NewVec(1, 1))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendit.NewVec(1, 1), bendit.NewVec(1, 3), bendit.NewVec(4, 1)), false, bvx)
}

/*func TestBezierVertex_WithEntry(t *testing.T) {
	bvx := NewBezierVertex(bendit.NewVec(0, 0), nil, bendit.NewVec(0, 1)).
		WithEntry(bendit.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(2, 2), bendit.NewVec(-2, -2)), true, bvx)
	bvx = NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), bendit.NewVec(0, 1)).
		WithEntry(bendit.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(2, 2), bendit.NewVec(0, 1)), false, bvx)
}

func TestBezierVertex_WithExit(t *testing.T) {
	bvx := NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), nil).
		WithExit(bendit.NewVec(-2, -2))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(2, 2), bendit.NewVec(-2, -2)), true, bvx)
	bvx = NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), bendit.NewVec(0, 1)).
		WithExit(bendit.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendit.NewVec(0, 0), bendit.NewVec(0, 1), bendit.NewVec(2, 2)), false, bvx)
}
*/
