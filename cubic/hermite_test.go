package cubic

import (
	"github.com/stretchr/testify/assert"
	"github.com/walpod/bendigo"
	"math"
	"testing"
)

func createHermDiag00to11() *HermiteVertBuilder {
	return NewHermiteVertBuilder(nil,
		NewHermiteVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 0), bendigo.NewVec(1, 1)),
		NewHermiteVertex(bendigo.NewVec(1, 1), bendigo.NewVec(1, 1), bendigo.NewVec(0, 0)),
	)
}

func createNonUniHermDiag00to11() *HermiteVertBuilder {
	return NewHermiteVertBuilder([]float64{0, math.Sqrt2},
		NewHermiteVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 0), bendigo.NewVec(1, 1)),
		NewHermiteVertex(bendigo.NewVec(1, 1), bendigo.NewVec(1, 1), bendigo.NewVec(0, 0)),
	)
}

func isOnDiag(v bendigo.Vec) bool {
	return math.Abs(v[0]-v[1]) < delta
}

func createHermParabola00to11(uniform bool) *HermiteVertBuilder {
	var tknots []float64
	if uniform {
		tknots = nil
	} else {
		tknots = []float64{0, 1} // is in fact uniform but specified as non-uniform
	}
	return NewHermiteVertBuilder(tknots,
		NewHermiteVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 0), bendigo.NewVec(1, 0)),
		NewHermiteVertex(bendigo.NewVec(1, 1), bendigo.NewVec(1, 2), bendigo.NewVec(0, 0)),
	)
}

func createDoubleHermParabola00to11to22(uniform bool) *HermiteVertBuilder {
	var tknots []float64
	if uniform {
		tknots = nil
	} else {
		tknots = []float64{0, 1, 2} // is in fact uniform but specified as non-uniform
	}
	return NewHermiteVertBuilder(tknots,
		NewHermiteVertex(bendigo.NewVec(0, 0), bendigo.NewVec(0, 0), bendigo.NewVec(1, 0)),
		NewHermiteVertex(bendigo.NewVec(1, 1), bendigo.NewVec(1, 2), bendigo.NewVec(1, 0)),
		NewHermiteVertex(bendigo.NewVec(2, 2), bendigo.NewVec(1, 2), bendigo.NewVec(0, 0)),
	)
}

func TestHermiteSpline_At(t *testing.T) {
	herm := createHermDiag00to11().Build()
	AssertSplineAt(t, herm, 0, bendigo.NewVec(0, 0))
	AssertSplineAt(t, herm, 0.25, bendigo.NewVec(0.25, 0.25))
	AssertSplineAt(t, herm, .5, bendigo.NewVec(.5, .5))
	AssertSplineAt(t, herm, 0.75, bendigo.NewVec(0.75, 0.75))
	AssertSplineAt(t, herm, 1, bendigo.NewVec(1, 1))

	herm = createNonUniHermDiag00to11().Build()
	ts, te := herm.Knots().Tstart(), herm.Knots().Tend()
	AssertSplineAt(t, herm, ts, bendigo.NewVec(0, 0))
	AssertSplineAt(t, herm, te/2, bendigo.NewVec(.5, .5))
	AssertSplineAt(t, herm, te, bendigo.NewVec(1, 1))
	for i := 0; i < 100; i++ {
		AssertRandSplinePointProperty(t, herm, isOnDiag, "hermite point must be on diagonal")
	}

	herm = createHermParabola00to11(true).Build()
	AssertSplineAt(t, herm, 0, bendigo.NewVec(0, 0))
	AssertSplineAt(t, herm, 0.25, bendigo.NewVec(0.25, 0.25*0.25))
	AssertSplineAt(t, herm, 0.5, bendigo.NewVec(0.5, 0.25))
	AssertSplineAt(t, herm, 0.75, bendigo.NewVec(0.75, 0.75*0.75))
	AssertSplineAt(t, herm, 1, bendigo.NewVec(1, 1))

	herm = createDoubleHermParabola00to11to22(true).Build()
	AssertSplineAt(t, herm, 0, bendigo.NewVec(0, 0))
	AssertSplineAt(t, herm, 0.25, bendigo.NewVec(0.25, 0.25*0.25))
	AssertSplineAt(t, herm, 0.5, bendigo.NewVec(0.5, 0.25))
	AssertSplineAt(t, herm, 0.75, bendigo.NewVec(0.75, 0.75*0.75))
	AssertSplineAt(t, herm, 1, bendigo.NewVec(1, 1))
	AssertSplineAt(t, herm, 1.25, bendigo.NewVec(1.25, 1+0.25*0.25))
	AssertSplineAt(t, herm, 1.5, bendigo.NewVec(1.5, 1.25))
	AssertSplineAt(t, herm, 1.75, bendigo.NewVec(1.75, 1+0.75*0.75))
	AssertSplineAt(t, herm, 2, bendigo.NewVec(2, 2))

	// domain with ony one value: 0
	herm = NewHermiteVertBuilder(nil,
		NewHermiteVertex(bendigo.NewVec(1, 2), bendigo.NewVec(0, 0), bendigo.NewVec(0, 0))).
		Build()
	AssertSplineAt(t, herm, 0, bendigo.NewVec(1, 2))

	// empty domain
	herm = NewHermiteVertBuilder(nil).Build()
	ts, te = herm.Knots().Tstart(), herm.Knots().Tend()
	assert.Greaterf(t, ts, te, "empty knots: tstart %v must be greater than tend %v", ts, te)

	// uniform and regular non-uniform must match
	herm = createHermParabola00to11(true).Build()
	nuherm := createHermParabola00to11(false).Build()
	AssertSplinesEqual(t, herm, nuherm, 100)

	herm = createDoubleHermParabola00to11to22(true).Build()
	nuherm = createDoubleHermParabola00to11to22(false).Build()
	AssertSplinesEqual(t, herm, nuherm, 100)
}

// TODO implement with LinaxSpline
func TestHermiteApproxer_Approx(t *testing.T) {
	hermBuilder := createDoubleHermParabola00to11to22(true)
	lc := bendigo.NewLineToSliceCollector()
	bendigo.LinaxAll(hermBuilder, lc, bendigo.NewLinaxParams(0.02))
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	AssertVecInDelta(t, bendigo.NewVec(0, 0), lc.Lines[0].Pstart, "start point = [0,0]")
	AssertVecInDelta(t, bendigo.NewVec(2, 2), lc.Lines[len(lc.Lines)-1].Pend, "end point = [2,2]")

	// start points of approximated lines must be on bezier curve and match bezier.At
	hermBuilder = createHermParabola00to11(true)
	lc = bendigo.NewLineToSliceCollector()
	bendigo.LinaxAll(hermBuilder, lc, bendigo.NewLinaxParams(0.02))
	assert.Greater(t, len(lc.Lines), 1, "approximated with more than one line")
	AssertApproxStartPointsMatchSpline(t, lc.Lines, hermBuilder.Build())
}

func TestHermiteVertBuilder_AddVertex(t *testing.T) {
	hermite := createHermDiag00to11()
	err := hermite.AddVertex(3, nil)
	assert.NotNil(t, err, "knot-no. too large")
	err = hermite.AddVertex(2, NewHermiteVertex(bendigo.NewVec(2, 2), bendigo.NewVec(1.5, 1.5), nil))
	assert.Equal(t, hermite.knots.KnotCnt(), 3, "knot-cnt %v wrong", hermite.knots.KnotCnt())
	err = hermite.AddVertex(0, NewHermiteVertex(bendigo.NewVec(-1, -1), bendigo.NewVec(-2, -2), nil))
	assert.Equal(t, hermite.knots.KnotCnt(), 4, "knot-cnt %v wrong", hermite.knots.KnotCnt())
	assert.Equal(t, hermite.Vertex(1), createHermDiag00to11().Vertex(0), "vertices don't match")
	assert.Equal(t, hermite.Vertex(2), createHermDiag00to11().Vertex(1), "vertices don't match")
}

func TestHermiteVertBuilder_DeleteVertex(t *testing.T) {
	hermite := createHermDiag00to11()
	err := hermite.DeleteVertex(2)
	assert.NotNil(t, err, "knot-no. doesn't exist")
	err = hermite.DeleteVertex(1)
	assert.Equal(t, hermite.knots.KnotCnt(), 1, "knot-cnt %v wrong", hermite.knots.KnotCnt())
	assert.Equal(t, hermite.Vertex(0), createHermDiag00to11().Vertex(0), "vertices don't match")
	err = hermite.DeleteVertex(0)
	assert.Equal(t, hermite.knots.KnotCnt(), 0, "knot-cnt %v wrong", hermite.knots.KnotCnt())
}
