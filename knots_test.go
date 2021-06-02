package bendit

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestSpline2d struct {
	segmentCnt int
}

func (sp TestSpline2d) SegmentCnt() int {
	return sp.segmentCnt
}

func (sp TestSpline2d) Knots() Knots {
	panic("implement me")
}

func (sp TestSpline2d) At(t float64) (x, y float64) {
	panic("implement me")
}

func (sp TestSpline2d) Fn() Fn2d {
	panic("implement me")
}

func (sp TestSpline2d) Approx(maxDist float64, collector LineCollector2d) {
	panic("implement me")
}

const delta = 0.0000000001

func TestNewUniformKnots(t *testing.T) {
	segmentCnt := 3
	uniknots := NewUniformKnots()
	uniknots.SetSplineIfEmpty(TestSpline2d{segmentCnt})
	var knots Knots = uniknots

	assert.True(t, knots.IsUniform(), "knots must be uniform")
	assert.Equal(t, 0., knots.Domain().Start, "domain must start at 0")
	assert.Equal(t, float64(segmentCnt), knots.Domain().End, "domain must end at %v", segmentCnt)
	assert.Equal(t, segmentCnt+1, knots.Count(), "must have %v knots", segmentCnt+1)
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

func TestNewNonUniformKnots(t *testing.T) {
	segmentCnt := 3
	var knots Knots = NewNonUniformKnots([]float64{0, 0.8, 2.5, 3})

	assert.False(t, knots.IsUniform(), "knots may not be uniform")
	assert.Equal(t, 0., knots.Domain().Start, "domain must start at 0")
	assert.Equal(t, float64(segmentCnt), knots.Domain().End, "domain must end at %v", float64(segmentCnt))
	assert.Equal(t, segmentCnt+1, knots.Count(), "must have %v knots", segmentCnt+1)
	t1, _ := knots.Knot(1)
	assert.Equal(t, 0.8, t1, "knot must be 0.8")
	segmentLen, _ := knots.SegmentLen(2)
	assert.Equal(t, 0.5, segmentLen, "segment must have length 0.5")
	segmentNo, u, _ := knots.MapToSegment(2.5)
	//TODO assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	//TODO assert.InDeltaf(t, 0., u, delta, "segment-local u must be %v", 0.)
	segmentNo, u, _ = knots.MapToSegment(2.8)
	assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	assert.InDeltaf(t, 3./5, u, delta, "segment-local u must be %v", 3./5)
	segmentNo, u, _ = knots.MapToSegment(3)
	assert.Equal(t, 2, segmentNo, "must be mapped to segment-no 2")
	assert.InDeltaf(t, 1., u, delta, "segment-local u must be %v", 1.)
}
