package bendit

type Knots interface {
	IsUniform() bool
	Count() int
	Tstart() float64
	Tend() float64
	Knot(knotNo int) (t float64, err error)
	SegmentCnt() int
	SegmentLen(segmentNo int) (t float64, err error)
	MapToSegment(t float64) (segmentNo int, u float64, err error)
	External() []float64 // external representation: uniform = nil, non-uniform = slice (non nil)
}

type Vertex2d interface {
	Coord() (x, y float64)
}

type Fn2d func(t float64) (x, y float64)

type Spline2d interface {
	Knots() Knots
	Vertex(knotNo int) (vertex Vertex2d, err error)
	At(t float64) (x, y float64)
	Fn() Fn2d
	Approx(maxDist float64, collector LineCollector2d) // TODO from-to knot
}

type LineCollector2d interface {
	// CollectLine from start (pstartx,pstarty) to end point (pendx,pendy) for parameter range (tstart..tend)
	CollectLine(tstart, tend, pstartx, pstarty, pendx, pendy float64)
}
