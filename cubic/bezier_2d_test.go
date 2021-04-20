package cubic

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

const delta = 0.0000000001

func TestProjectedVectorDist(t *testing.T) {
	assert.InDeltaf(t, 1., ProjectedVectorDist(0, 1, 1, 0), delta, "unit square")
	assert.InDeltaf(t, 1., ProjectedVectorDist(1./math.Sqrt2, 1./math.Sqrt2, 1./math.Sqrt2, -1./math.Sqrt2), delta, "unit square, rotated")
	assert.InDeltaf(t, 2., ProjectedVectorDist(0, 2, 2, 0), delta, "square - side length 2")

	assert.InDeltaf(t, 1., ProjectedVectorDist(0, 1, 2, 0), delta, "rectangle")
	assert.InDeltaf(t, 1., ProjectedVectorDist(1./math.Sqrt2, 1./math.Sqrt2, math.Sqrt2, -math.Sqrt2), delta, "rectangle, rotated")
	assert.InDeltaf(t, 1., ProjectedVectorDist(1./math.Sqrt2, 1./math.Sqrt2, 10, -10), delta, "rectangle, rotated, enlarged")

	assert.InDeltaf(t, 1., ProjectedVectorDist(1, 1, 1, 0), delta, "45 degree")
	assert.InDeltaf(t, 1., ProjectedVectorDist(1, 1, -10, 0), delta, "45 degree, other direction")
	assert.InDeltaf(t, 1., ProjectedVectorDist(math.Sqrt2, 0, -3, 3), delta, "45 degree")
}
