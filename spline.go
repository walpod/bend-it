package bendit

type Knots interface {
	IsUniform() bool
	Count() int
	Tstart() float64
	Tend() float64
	Knot(knotNo int) (t float64, err error)
	SegmentLen(segmentNo int) (t float64, err error)
	MapToSegment(t float64) (segmentNo int, u float64, err error)
}

type Vertex2d interface {
	Coord() (x, y float64)
}

type Fn2d func(t float64) (x, y float64)

type Spline2d interface {
	SegmentCnt() int
	Knots() Knots
	At(t float64) (x, y float64)
	Fn() Fn2d
	Approx(maxDist float64, collector LineCollector2d)
}

type LineCollector2d interface {
	// CollectLine from start (pstartx,pstarty) to end point (pendx,pendy) for parameter range (tstart..tend)
	CollectLine(tstart, tend, pstartx, pstarty, pendx, pendy float64)
}
