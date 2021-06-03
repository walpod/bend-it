package bendit

type Fn2d func(t float64) (x, y float64)

type Vertex2d interface {
	Coord() (x, y float64)
}

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

// DirectCollector2d supports the simple case of using a single collect func
// TODO move to other file
type DirectCollector2d struct {
	line func(tstart, tend, pstartx, pstarty, pendx, pendy float64)
}

func NewDirectCollector2d(line func(tstart, tend, pstartx, pstarty, pendx, pendy float64)) *DirectCollector2d {
	return &DirectCollector2d{line: line}
}

func (lc DirectCollector2d) CollectLine(tstart, tend, pstartx, pstarty, pendx, pendy float64) {
	lc.line(tstart, tend, pstartx, pstarty, pendx, pendy)
}
