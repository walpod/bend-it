package bendit

type LineCollector interface {
	// CollectLine is called from start (pstartx,pstarty) to end point (pendx,pendy) in consecutive order
	// for parameter range (tstart..tend)
	CollectLine(segmentNo int, tstart, tend float64, pstart, pend Vec)
}

// DirectCollector supports the simple case of using a single collect func
type DirectCollector struct {
	line func(segmentNo int, tstart, tend float64, pstart, pend Vec)
}

func NewDirectCollector(line func(segmentNo int, tstart, tend float64, pstart, pend Vec)) *DirectCollector {
	return &DirectCollector{line: line}
}

func (lc DirectCollector) CollectLine(segmentNo int, tstart, tend float64, pstart, pend Vec) {
	lc.line(segmentNo, tstart, tend, pstart, pend)
}

// LineToSliceCollector collects lines in slice
type LineToSliceCollector struct {
	Lines []LineParams
}

type LineParams struct {
	SegmentNo    int
	Tstart, Tend float64
	Pstart, Pend Vec
}

func NewLineToSliceCollector() *LineToSliceCollector {
	return &LineToSliceCollector{Lines: make([]LineParams, 0)}
}

func (lc *LineToSliceCollector) CollectLine(segmentNo int, tstart, tend float64, pstart, pend Vec) {
	lc.Lines = append(lc.Lines, LineParams{segmentNo, tstart, tend, pstart, pend})
}
