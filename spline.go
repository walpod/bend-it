package bendit

type Fn2d func(t float64) Vec

type Spline2d interface {
	Knots() Knots

	At(t float64) Vec
	Fn() Fn2d

	Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector LineCollector2d)
}

type Vertex2d interface {
	Loc() Vec
	Translate(d Vec) Vertex2d
}

// VertSpline2d can be constructed by adding vertices
type VertSpline2d interface {
	Spline2d

	Vertex(knotNo int) Vertex2d
	AddVertex(knotNo int, vertex Vertex2d) (err error)
	UpdateVertex(knotNo int, vertex Vertex2d) (err error)
	DeleteVertex(knotNo int) (err error)
}

func ApproxAll(spline Spline2d, maxDist float64, collector LineCollector2d) {
	spline.Approx(0, spline.Knots().SegmentCnt()-1, maxDist, collector)
}

func Vertices(spline VertSpline2d) []Vertex2d {
	cnt := spline.Knots().KnotCnt()
	vertices := make([]Vertex2d, cnt)
	for i := 0; i < cnt; i++ {
		vertices[i] = spline.Vertex(i)
	}
	return vertices
}
