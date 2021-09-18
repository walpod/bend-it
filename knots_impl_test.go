package bendit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const delta = 0.0000000001

func TestUniformKnots(t *testing.T) {
	knotsCnt := 4
	knots := NewUniformKnots(knotsCnt)
	assert.True(t, knots.IsUniform(), "knots must be uniform")
	assert.Empty(t, knots.External(), "external representation must be nil")

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
	tstart, tend := 0., 3.
	ks := []float64{tstart, 0.8, 2.5, tend}
	knots := NewNonUniformKnots(ks)

	assert.False(t, knots.IsUniform(), "knots may not be uniform")
	assert.Equal(t, knots.External(), ks, "external representation must be %v", ks)

	assert.Equal(t, len(ks), knots.KnotCnt(), "must have %v knots", len(ks))
	assert.Equal(t, tstart, knots.Tstart(), "T must start at 0")
	assert.Equal(t, tend, knots.Tend(), "T must end at %v", tend)
	t1, _ := knots.Knot(1)
	assert.Equal(t, 0.8, t1, "knot must be 0.8")
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

// TODO move to other test file
func TestAdjacentSegments(t *testing.T) {
	knots := NewUniformKnots(4)
	fromSegmentNo, toSegmentNo, err := AdjacentSegments(knots, 2, true, true)
	assert.Equal(t, 1, fromSegmentNo, "fromSegmentNo is 1")
	assert.Equal(t, 2, toSegmentNo, "toSegmentNo is 2")
	fromSegmentNo, toSegmentNo, err = AdjacentSegments(knots, 2, true, false)
	assert.Equal(t, 1, fromSegmentNo, "fromSegmentNo is 1")
	assert.Equal(t, 1, toSegmentNo, "toSegmentNo is 1 (segment after is ignored)")
	_, _, err = AdjacentSegments(knots, 0, true, false)
	assert.NotEqual(t, nil, err, "AdjacentSegments don't exist, error must be not-nil")
	_, _, err = AdjacentSegments(knots, knots.KnotCnt()-1, false, true)
	assert.NotEqual(t, nil, err, "AdjacentSegments don't exist, error must be not-nil")
	_, _, err = AdjacentSegments(knots, 2, false, false)
	assert.NotEqual(t, nil, err, "AdjacentSegments don't exist, error must be not-nil")
	emptyKnots := NewUniformKnots(0)
	_, _, err = AdjacentSegments(emptyKnots, 0, true, true)
	assert.NotEqual(t, nil, err, "KnotNo doesn't exist, error must be not-nil")
	singleKnots := NewUniformKnots(1)
	_, _, err = AdjacentSegments(singleKnots, 0, true, true)
	assert.NotEqual(t, nil, err, "AdjacentSegments don't exist, error must be not-nil")

}
