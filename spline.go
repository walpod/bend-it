package bendit

type Fn2d func(t float64) (x, y float64)

// range of parameter t for which the spline is defined
type SplineDomain struct {
	From, To float64
}

type Spline2d interface {
	SegmentCnt() int
	Knots() *Knots
	At(t float64) (x, y float64)
	Fn() Fn2d
	//Approximate(maxDist float64, collector LineCollector2d)
}

type LineCollector2d interface {
	CollectLine(x0, y0, x3, y3 float64)
}
