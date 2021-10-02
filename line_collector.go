package bendit

type LineCollector2d interface {
	// CollectLine is called from start (pstartx,pstarty) to end point (pendx,pendy) in consecutive order
	// for parameter range (tstart..tend)
	CollectLine(segmentNo int, tstart, tend float64, pstart, pend Vec)
}

// DirectCollector2d supports the simple case of using a single collect func
type DirectCollector2d struct {
	line func(segmentNo int, tstart, tend float64, pstart, pend Vec)
}

func NewDirectCollector2d(line func(segmentNo int, tstart, tend float64, pstart, pend Vec)) *DirectCollector2d {
	return &DirectCollector2d{line: line}
}

func (lc DirectCollector2d) CollectLine(segmentNo int, tstart, tend float64, pstart, pend Vec) {
	lc.line(segmentNo, tstart, tend, pstart, pend)
}

// LineToSliceCollector2d collects lines in slice
type LineToSliceCollector2d struct {
	Lines []LineParams
}

type LineParams struct {
	SegmentNo    int
	Tstart, Tend float64
	Pstart, Pend Vec
}

func NewLineToSliceCollector2d() *LineToSliceCollector2d {
	return &LineToSliceCollector2d{Lines: make([]LineParams, 0)}
}

func (lc *LineToSliceCollector2d) CollectLine(segmentNo int, tstart, tend float64, pstart, pend Vec) {
	lc.Lines = append(lc.Lines, LineParams{segmentNo, tstart, tend, pstart, pend})
}
