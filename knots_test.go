package bendit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const delta = 0.0000000001

func TestUniformKnots(t *testing.T) {
	knots := NewUniformKnots(3)
	assert.True(t, knots.IsUniform(), "knots must be uniform")
	assert.Empty(t, knots.External(), "external representation must be nil")

	err := knots.AddKnot(10)
	assert.NotNil(t, err, "not existent knot no., error must be not-nil")
	err = knots.AddKnot(3)
	assert.Equal(t, nil, err, "must be success")
	err = knots.DeleteKnot(1)
	assert.Equal(t, nil, err, "must be success")
	err = knots.AddKnot(0)
	assert.Equal(t, nil, err, "must be success")
	knotsCnt := 4

	assert.Equal(t, knotsCnt, knots.KnotCnt(), "must have %v knots", knotsCnt)
	assert.Equal(t, 0., knots.Tstart(), "T must start at 0")
	assert.Equal(t, float64(knotsCnt-1), knots.Tend(), "T must end at %v", float64(knotsCnt-1))
	t1, _ := knots.Knot(1)
	assert.Equal(t, 1., t1, "knot must be 1")
	segmentLen, _ := knots.SegmentLen(2)
	assert.Equal(t, 1., segmentLen, "segment must have length 1")

	segmentNo, u, _ := knots.MapToSegment(2)
	assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	assert.Equal(t, 0., u, "segment-local u must be %v", 0)
	segmentNo, u, _ = knots.MapToSegment(2.5)
	assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	assert.Equal(t, 0.5, u, "segment-local u must be %v", 0.5)
	segmentNo, u, _ = knots.MapToSegment(3)
	assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	assert.Equal(t, 1., u, "segment-local u must be %v", 1)
}

func TestNonUniformKnots(t *testing.T) {
	t0, t2, t3 := 0., 2.5, 3.
	ks := []float64{t0, 0.8, t2, t3}
	knots := NewNonUniformKnots(ks)
	assert.False(t, knots.IsUniform(), "knots may not be uniform")
	assert.Equal(t, knots.External(), ks, "external representation must be %v", ks)

	err := knots.AddKnot(10)
	assert.NotNil(t, err, "not existent knot no., error must be not-nil")
	err = knots.AddKnot(2)
	assert.Equal(t, nil, err, "must be success")
	err = knots.DeleteKnot(2)
	assert.Equal(t, nil, err, "must be success")
	err = knots.DeleteKnot(3)
	err = knots.AddKnot(3)
	err = knots.SetSegmentLen(2, t3-t2)
	assert.Equal(t, nil, err, "must be success")
	assert.Equal(t, ks, knots.External(), "must be equal")

	// move t1 from 0.8 to 0.5
	err = knots.SetSegmentLen(0, 0.5)
	err = knots.SetSegmentLen(1, 2)
	//fmt.Println(knots.External())

	assert.Equal(t, len(ks), knots.KnotCnt(), "must have %v knots", len(ks))
	assert.Equal(t, t0, knots.Tstart(), "T must start at 0")
	assert.Equal(t, t3, knots.Tend(), "T must end at %v", t3)
	t1, _ := knots.Knot(1)
	assert.Equal(t, 0.5, t1, "knot must be 0.5")
	segmentLen, _ := knots.SegmentLen(2)
	assert.Equal(t, 0.5, segmentLen, "segment must have length 0.5")

	segmentNo, u, _ := knots.MapToSegment(2.5)
	assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	assert.InDeltaf(t, 0., u, delta, "segment-local u must be %v", 0.)
	segmentNo, u, _ = knots.MapToSegment(2.8)
	assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	assert.InDeltaf(t, 3./5, u, delta, "segment-local u must be %v", 3./5)
	segmentNo, u, _ = knots.MapToSegment(3)
	assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	assert.InDeltaf(t, 1., u, delta, "segment-local u must be %v", 1.)
}

func TestAdjacentSegments(t *testing.T) {
	knots := NewUniformKnots(4)
	fromSegmentNo, toSegmentNo, err := AdjacentSegments(knots, 2, true, true)
	assert.Equal(t, 1, fromSegmentNo, "fromSegmentNo is 1")
	assert.Equal(t, 2, toSegmentNo, "toSegmentNo is 2")
	fromSegmentNo, toSegmentNo, err = AdjacentSegments(knots, 2, true, false)
	assert.Equal(t, 1, fromSegmentNo, "fromSegmentNo is 1")
	assert.Equal(t, 1, toSegmentNo, "toSegmentNo is 1 (segment after is ignored)")
	_, _, err = AdjacentSegments(knots, 0, true, false)
	assert.NotNil(t, err, "AdjacentSegments don't exist, error must be not-nil")
	_, _, err = AdjacentSegments(knots, knots.KnotCnt()-1, false, true)
	assert.NotNil(t, err, "AdjacentSegments don't exist, error must be not-nil")
	_, _, err = AdjacentSegments(knots, 2, false, false)
	assert.NotNil(t, err, "AdjacentSegments don't exist, error must be not-nil")

	emptyKnots := NewUniformKnots(0)
	_, _, err = AdjacentSegments(emptyKnots, 0, true, true)
	assert.NotNil(t, err, "KnotNo doesn't exist, error must be not-nil")

	singleKnots := NewUniformKnots(1)
	_, _, err = AdjacentSegments(singleKnots, 0, true, true)
	assert.NotNil(t, err, "AdjacentSegments don't exist, error must be not-nil")
}
