package bendit

type Spline interface {
	Knots() Knots
	At(t float64) Vec
}

type SplineApproxer interface {
	Knots() Knots
	Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector LineCollector2d)
}

type Vertex interface {
	Loc() Vec
}

type SplineBuilder interface {
	Knots() Knots
	Build() Spline
	BuildApproxer() SplineApproxer
}

// VertSplineBuilder can be constructed by adding vertices
type VertSplineBuilder interface {
	SplineBuilder

	Vertex(knotNo int) Vertex
	AddVertex(knotNo int, vertex Vertex) (err error)
	UpdateVertex(knotNo int, vertex Vertex) (err error)
	DeleteVertex(knotNo int) (err error)
}

func ApproxAll(splineApproxer SplineApproxer, maxDist float64, collector LineCollector2d) {
	splineApproxer.Approx(0, splineApproxer.Knots().SegmentCnt()-1, maxDist, collector)
}

func Vertices(builder VertSplineBuilder) []Vertex {
	cnt := builder.Knots().KnotCnt()
	vertices := make([]Vertex, cnt)
	for i := 0; i < cnt; i++ {
		vertices[i] = builder.Vertex(i)
	}
	return vertices
}
