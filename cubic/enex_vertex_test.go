package cubic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/walpod/bendigo"
	"testing"
)

func AssertEnexVerticesAreEqual(t *testing.T, expected *EnexVertex, expectedDependent bool, actual *EnexVertex) {
	AssertVecInDelta(t, expected.Loc(), actual.Loc(), fmt.Sprintf("expected location = %v != actual location = %v", expected.Loc(), actual.Loc()))
	AssertVecInDelta(t, expected.Entry(), actual.Entry(), fmt.Sprintf("expected entry-control = %v != actual = %v", expected.Entry(), actual.Entry()))
	AssertVecInDelta(t, expected.Exit(), actual.Exit(), fmt.Sprintf("expected exit-control = %v != actual = %v", expected.Entry(), actual.Entry()))
	assert.Equal(t, expected.Relative(), actual.Relative(), "expected relative = %v != actual relative = %v", expected.Relative(), actual.Relative())
	assert.Equal(t, expectedDependent, actual.Dependent(), "expected dependent = %v != actual dependent = %v", expectedDependent, actual.Dependent())
}

func TestEnexVertex_NewEnexVertex_DependentAbsolute(t *testing.T) {
	ev := NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(1, 2), nil, false)
	AssertVecInDelta(t, ev.entry.Negate(), ev.exit, "automatically created exit control must be on the other side of (= reflected by) origin [0,0] in absolute mode")

	ev = NewEnexVertex(bendigo.NewVec(0, 0), nil, bendigo.NewVec(3, -5), false)
	AssertVecInDelta(t, ev.entry, ev.exit.Negate(), "automatically created entry control must be on the other side of origin [0,0] in absolute mode")
}

func TestEnexVertex_NewEnexVertex_DependentRelative(t *testing.T) {
	ev := NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(1, 2), nil, true)
	AssertVecInDelta(t, ev.entry, ev.exit, "automatically created exit control must be equal to entry in relative mode")

	ev = NewEnexVertex(bendigo.NewVec(0, 0), nil, bendigo.NewVec(3, -5), true)
	AssertVecInDelta(t, ev.entry, ev.exit, "automatically created entry control must be equal to exit in relative mode")
}

func TestEnexVertex_ShiftRelative(t *testing.T) {
	ev := NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), nil, false)
	expected := NewEnexVertex(bendigo.NewVec(2, 0), bendigo.NewVec(2, 1), bendigo.NewVec(2, -1), false)
	AssertEnexVerticesAreEqual(t, expected, true, ev.WithShift(bendigo.NewVec(2, 0)))
	ev.Shift(bendigo.NewVec(2, 0))
	AssertEnexVerticesAreEqual(t, expected, true, ev)

	ev = NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 2), bendigo.NewVec(3, 0), false)
	expected = NewEnexVertex(bendigo.NewVec(1, 1), bendigo.NewVec(1, 3), bendigo.NewVec(4, 1), false)
	AssertEnexVerticesAreEqual(t, expected, false, ev.WithShift(bendigo.NewVec(1, 1)))
	ev.Shift(bendigo.NewVec(1, 1))
	AssertEnexVerticesAreEqual(t, expected, false, ev)
}

func TestEnexVertex_WithEntry(t *testing.T) {
	bvx := NewEnexVertex(bendigo.NewVec(0, 0), nil, bendigo.NewVec(0, 1), false).
		WithEntry(bendigo.NewVec(2, 2))
	AssertEnexVerticesAreEqual(t, NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(2, 2), bendigo.NewVec(-2, -2), false), true, bvx)

	bvx = NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), bendigo.NewVec(0, 1), false).
		WithEntry(bendigo.NewVec(2, 2))
	AssertEnexVerticesAreEqual(t, NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(2, 2), bendigo.NewVec(0, 1), false), false, bvx)
}

func TestEnexVertex_WithExit(t *testing.T) {
	bvx := NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), nil, false).
		WithExit(bendigo.NewVec(-2, -2))
	AssertEnexVerticesAreEqual(t, NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(2, 2), bendigo.NewVec(-2, -2), false), true, bvx)

	bvx = NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), bendigo.NewVec(0, 1), false).
		WithExit(bendigo.NewVec(2, 2))
	AssertEnexVerticesAreEqual(t, NewEnexVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 1), bendigo.NewVec(2, 2), false), false, bvx)
}

// TODO ControlAsAbsolute, Shift, ...
