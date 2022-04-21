package bendigo

type Spline interface {
	Knots() Knots

	// At calculates the Point (Vector) on the spline at parameter t
	At(t float64) Vec
}

type SplineBuilder interface {
	Knots() Knots

	// Spline builds it
	Spline() Spline

	// LinApproximate linearly approximates spline and passes lines consecutively to consumer
	LinApproximate(fromSegmentNo, toSegmentNo int, consumer LineConsumer, linaxParams *LinaxParams)

	// LinaxSpline builds linearly approximated spline
	LinaxSpline(linaxParams *LinaxParams) *LinaxSpline
}

type Vertex interface {
	Loc() Vec
}

// SplineVertBuilder is constructed by adding vertices
type SplineVertBuilder interface {
	SplineBuilder

	Vertex(knotNo int) Vertex
	AddVertex(knotNo int, vertex Vertex) (err error)
	UpdateVertex(knotNo int, vertex Vertex) (err error)
	DeleteVertex(knotNo int) (err error)
}

func Vertices(builder SplineVertBuilder) []Vertex {
	cnt := builder.Knots().KnotCnt()
	vertices := make([]Vertex, cnt)
	for i := 0; i < cnt; i++ {
		vertices[i] = builder.Vertex(i)
	}
	return vertices
}
