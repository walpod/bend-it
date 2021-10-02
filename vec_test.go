package bendit

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

// TODO

func TestProjectedVectorDist(t *testing.T) {
	assert.InDeltaf(t, 1., NewVec(0, 1).ProjectedVecDist(NewVec(1, 0)), delta, "unit square")
	assert.InDeltaf(t, 1., NewVec(1./math.Sqrt2, 1./math.Sqrt2).ProjectedVecDist(NewVec(1./math.Sqrt2, -1./math.Sqrt2)), delta, "unit square, rotated")
	assert.InDeltaf(t, 2., NewVec(0, 2).ProjectedVecDist(NewVec(2, 0)), delta, "square - side length 2")

	assert.InDeltaf(t, 1., NewVec(0, 1).ProjectedVecDist(NewVec(2, 0)), delta, "rectangle")
	assert.InDeltaf(t, 1., NewVec(1./math.Sqrt2, 1./math.Sqrt2).ProjectedVecDist(NewVec(math.Sqrt2, -math.Sqrt2)), delta, "rectangle, rotated")
	assert.InDeltaf(t, 1., NewVec(1./math.Sqrt2, 1./math.Sqrt2).ProjectedVecDist(NewVec(10, -10)), delta, "rectangle, rotated, enlarged")

	assert.InDeltaf(t, 1., NewVec(1, 1).ProjectedVecDist(NewVec(1, 0)), delta, "45 degree")
	assert.InDeltaf(t, 1., NewVec(1, 1).ProjectedVecDist(NewVec(-10, 0)), delta, "45 degree, other direction")
	assert.InDeltaf(t, 1., NewVec(math.Sqrt2, 0).ProjectedVecDist(NewVec(3, 3)), delta, "45 degree")
}
