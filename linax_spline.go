package bendigo

// LinaxSpline is a linearly approximated spline consisting of consecutive line segments
type LinaxSpline struct {
	knots Knots
	lines []Line
}

type Line struct {
	SegmentNo    int
	Tstart, Tend float64
	Pstart, Pend Vec
}

func NewLinaxSpline(knots Knots, lines []Line) *LinaxSpline {
	return &LinaxSpline{knots: knots, lines: lines}
}

func (sp LinaxSpline) Knots() Knots {
	return sp.knots
}

func (sp LinaxSpline) Lines() []Line {
	return sp.lines
}

func (sp LinaxSpline) At(t float64) Vec {
	// TODO improve using binary sort
	for _, line := range sp.lines {
		if t >= line.Tstart && t <= line.Tend {
			fac := (t - line.Tstart) / (line.Tend - line.Tstart)
			return line.Pend.Sub(line.Pstart).Scale(fac)
		}
	}
	return nil
}

// LinaxParams contains parameters to control linear approximation TODO enhance
type LinaxParams struct {
	MaxDist float64
}

func NewLinaxParams(maxDist float64) *LinaxParams {
	return &LinaxParams{MaxDist: maxDist}
}

func BuildLinaxSpline(splineBuilder SplineBuilder, linaxParams *LinaxParams) *LinaxSpline {
	lineCollector := NewLineToSliceCollector()
	splineBuilder.LinApproximate(0, splineBuilder.Knots().SegmentCnt()-1, lineCollector, linaxParams)
	return NewLinaxSpline(splineBuilder.Knots(), lineCollector.Lines)
}

// LineConsumer interface is used during linear approximation
type LineConsumer interface {
	// ConsumeLine consumes next line segment of linear approximation
	// assert: is called from start to end point in consecutive order
	ConsumeLine(segmentNo int, tstart, tend float64, pstart, pend Vec)
}

// LineToSliceCollector collects lines in slice
type LineToSliceCollector struct {
	Lines []Line
}

func NewLineToSliceCollector() *LineToSliceCollector {
	return &LineToSliceCollector{Lines: make([]Line, 0)}
}

func (lc *LineToSliceCollector) ConsumeLine(segmentNo int, tstart, tend float64, pstart, pend Vec) {
	lc.Lines = append(lc.Lines, Line{segmentNo, tstart, tend, pstart, pend})
}

// FuncLineConsumer consumes next line by calling a prepared function
type FuncLineConsumer struct {
	line func(segmentNo int, tstart, tend float64, pstart, pend Vec)
}

func NewFuncLineConsumer(line func(segmentNo int, tstart, tend float64, pstart, pend Vec)) *FuncLineConsumer {
	return &FuncLineConsumer{line: line}
}

func (lc FuncLineConsumer) ConsumeLine(segmentNo int, tstart, tend float64, pstart, pend Vec) {
	lc.line(segmentNo, tstart, tend, pstart, pend)
}
