package bendit

type Spline2d interface {
	Knots() Knots

	At(t float64) Vec

	Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector LineCollector2d)
}

type Vertex interface {
	Loc() Vec
}

// VertSpline2d can be constructed by adding vertices
type VertSpline2d interface {
	Spline2d

	Vertex(knotNo int) Vertex
	AddVertex(knotNo int, vertex Vertex) (err error)
	UpdateVertex(knotNo int, vertex Vertex) (err error)
	DeleteVertex(knotNo int) (err error)
}

func ApproxAll(spline Spline2d, maxDist float64, collector LineCollector2d) {
	spline.Approx(0, spline.Knots().SegmentCnt()-1, maxDist, collector)
}

func Vertices(spline VertSpline2d) []Vertex {
	cnt := spline.Knots().KnotCnt()
	vertices := make([]Vertex, cnt)
	for i := 0; i < cnt; i++ {
		vertices[i] = spline.Vertex(i)
	}
	return vertices
}
