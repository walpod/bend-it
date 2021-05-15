package bendit

type Fn2d func(t float64) (x, y float64)

// range of parameter t for which the spline is defined
type SplineDomain struct {
	Start, End float64
}

type Spline2d interface {
	SegmentCnt() int
	Knots() *Knots
	At(t float64) (x, y float64)
	Fn() Fn2d
	Approx(maxDist float64, collector LineCollector2d)
}

type LineCollector2d interface {
	// CollectLine from start (sx,sy) to end point (ex,ey) for parameter range (ts..te)
	CollectLine(ts, te, sx, sy, ex, ey float64)
}

type DirectCollector2d struct {
	line func(ts, te, sx, sy, ex, ey float64)
}

func NewDirectCollector2d(line func(ts, te, sx, sy, ex, ey float64)) *DirectCollector2d {
	return &DirectCollector2d{line: line}
}

func (lc DirectCollector2d) CollectLine(ts, te, sx, sy, ex, ey float64) {
	lc.line(ts, te, sx, sy, ex, ey)
}
