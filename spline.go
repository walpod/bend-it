package bendit

type Knots interface {
	IsUniform() bool
	Tstart() float64
	Tend() float64

	Cnt() int
	KnotExists(knotNo int) bool
	Knot(knotNo int) (t float64, err error)

	SegmentCnt() int
	SegmentExists(segmentNo int) bool
	SegmentLen(segmentNo int) (l float64, err error)
	MapToSegment(t float64) (segmentNo int, u float64, err error)

	External() []float64 // external representation: uniform = nil, non-uniform = slice (non nil)
}

type Vertex2d interface {
	Coord() (x, y float64)
}

type Fn2d func(t float64) (x, y float64)

type Spline2d interface {
	Knots() Knots
	Vertex(knotNo int) Vertex2d
	At(t float64) (x, y float64)
	Fn() Fn2d
	Approx(fromSegmentNo, toSegmentNo int, maxDist float64, collector LineCollector2d)
}

type LineCollector2d interface {
	// CollectLine is called from start (pstartx,pstarty) to end point (pendx,pendy) in consecutive order
	// for parameter range (tstart..tend)
	CollectLine(segmentNo int, tstart, tend, pstartx, pstarty, pendx, pendy float64)
}

func ApproxAll(spline Spline2d, maxDist float64, collector LineCollector2d) {
	spline.Approx(0, spline.Knots().SegmentCnt()-1, maxDist, collector)
}
