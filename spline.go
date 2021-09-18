package bendit

import "fmt"

type Knots interface {
	IsUniform() bool
	Tstart() float64
	Tend() float64

	KnotCnt() int
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

func AdjacentSegments(knots Knots, knotNo int, inclBefore bool, inclAfter bool) (fromSegmentNo int, toSegmentNo int, err error) {
	if !knots.KnotExists(knotNo) {
		return 0, -1, fmt.Errorf("knot with number %v doesn't exist", knotNo)
	} else {
		if inclBefore && knotNo > 0 {
			fromSegmentNo = knotNo - 1
		} else {
			fromSegmentNo = knotNo
		}
		if inclAfter && knotNo < knots.KnotCnt()-1 {
			toSegmentNo = knotNo
		} else {
			toSegmentNo = knotNo - 1
		}
		if toSegmentNo < fromSegmentNo {
			err = fmt.Errorf("no matching segments found")
		}
		return
	}
}
