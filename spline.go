package bendit

type Vertex2d interface {
	Coord() (x, y float64)
}

type Fn2d func(t float64) (x, y float64)

type Spline2d interface {
	Knots() Knots

	Vertex(knotNo int) Vertex2d
	// TODO
	//AddVertex(knotNo int, vertex Vertex2d) (err error)
	//UpdateVertex(knotNo int, vertex Vertex2d) (err error)
	//DeleteVertex(knotNo int, vertex Vertex2d) (err error)

	At(t float64) (x, y float64)
	Fn() Fn2d

	Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector LineCollector2d)
}

func ApproxAll(spline Spline2d, maxDist float64, collector LineCollector2d) {
	spline.Approx(0, spline.Knots().SegmentCnt()-1, maxDist, collector)
}
