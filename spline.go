package bendit

type Fn2d func(t float64) (x, y float64)

type Spline2d interface {
	Knots() Knots

	At(t float64) (x, y float64)
	Fn() Fn2d

	Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector LineCollector2d)
}

type Vertex2d interface {
	Coord() (x, y float64)
	Translate(dx, dy float64) Vertex2d
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
