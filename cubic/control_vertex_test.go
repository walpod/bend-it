package cubic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/walpod/bendigo"
	"testing"
)

func AssertControlVerticesAreEqual(t *testing.T, expected ControlVertex, expectedDependent bool, actual ControlVertex) {
	AssertVecInDelta(t, expected.Loc(), actual.Loc(), fmt.Sprintf("expected bezier = %v != actual bezier = %v", expected.Loc(), actual.Loc()))
	AssertVecInDelta(t, expected.Entry(), actual.Entry(), fmt.Sprintf("expected entry-control = %v != actual = %v", expected.Entry(), actual.Entry()))
	AssertVecInDelta(t, expected.Exit(), actual.Exit(), fmt.Sprintf("expected exit-control = %v != actual = %v", expected.Entry(), actual.Entry()))
	assert.Equal(t, expectedDependent, actual.Dependent(), "expected dependent = %v != actual dependent = %v", expectedDependent, actual.Dependent())
}

func TestBezierVertex_Dependent(t *testing.T) {
	bvx := NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(1, 2), nil)
	AssertVecInDelta(t, bvx.entry.Negate(), bvx.exit, "dependent control must be reflected by origin [0,0]")
	bvx = NewBezierVertex(bendigo.NewVec(0, 0), nil, bendigo.NewVec(3, -5))
	AssertVecInDelta(t, bvx.entry, bvx.exit.Negate(), "dependent control must be reflected by origin [0,0]")
}

func TestBezierVertex_Translate(t *testing.T) {
	bvx := NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), nil).Translate(bendigo.NewVec(2, 0))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendigo.NewVec(2, 0), bendigo.NewVec(2, 1), bendigo.NewVec(2, -1)), true, bvx)
	bvx = NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 2), bendigo.NewVec(3, 0)).Translate(bendigo.NewVec(1, 1))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendigo.NewVec(1, 1), bendigo.NewVec(1, 3), bendigo.NewVec(4, 1)), false, bvx)
}

// TODO ControlToLoc, LocToControl
// TODO HermiteVertex like BezierVertex
// TODO NewControlVertexWithControl, NewControlVertexWithControlLoc, Control, ControlLoc for BezierVertext and/or HermiteVertex

/*func TestBezierVertex_WithEntry(t *testing.T) {
	bvx := NewBezierVertex(bendigo.NewVec(0, 0), nil, bendigo.NewVec(0, 1)).
		WithEntry(bendigo.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(2, 2), bendigo.NewVec(-2, -2)), true, bvx)
	bvx = NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), bendigo.NewVec(0, 1)).
		WithEntry(bendigo.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(2, 2), bendigo.NewVec(0, 1)), false, bvx)
}

func TestBezierVertex_WithExit(t *testing.T) {
	bvx := NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), nil).
		WithExit(bendigo.NewVec(-2, -2))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(2, 2), bendigo.NewVec(-2, -2)), true, bvx)
	bvx = NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), bendigo.NewVec(0, 1)).
		WithExit(bendigo.NewVec(2, 2))
	AssertControlVerticesAreEqual(t, NewBezierVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), bendigo.NewVec(2, 2)), false, bvx)
}
*/
