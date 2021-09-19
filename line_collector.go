package bendit

type LineCollector2d interface {
	// CollectLine is called from start (pstartx,pstarty) to end point (pendx,pendy) in consecutive order
	// for parameter range (tstart..tend)
	CollectLine(segmentNo int, tstart, tend, pstartx, pstarty, pendx, pendy float64)
}

// DirectCollector2d supports the simple case of using a single collect func
type DirectCollector2d struct {
	line func(segmentNo int, tstart, tend, pstartx, pstarty, pendx, pendy float64)
}

func NewDirectCollector2d(line func(segmentNo int, tstart, tend, pstartx, pstarty, pendx, pendy float64)) *DirectCollector2d {
	return &DirectCollector2d{line: line}
}

func (lc DirectCollector2d) CollectLine(segmentNo int, tstart, tend, pstartx, pstarty, pendx, pendy float64) {
	lc.line(segmentNo, tstart, tend, pstartx, pstarty, pendx, pendy)
}

// LineToSliceCollector2d collects lines in slice
type LineToSliceCollector2d struct {
	Lines []LineParams
}

type LineParams struct {
	SegmentNo                                    int
	Tstart, Tend, Pstartx, Pstarty, Pendx, Pendy float64
}

func NewLineToSliceCollector2d() *LineToSliceCollector2d {
	return &LineToSliceCollector2d{Lines: make([]LineParams, 0)}
}

func (lc *LineToSliceCollector2d) CollectLine(segmentNo int, tstart, tend, pstartx, pstarty, pendx, pendy float64) {
	lc.Lines = append(lc.Lines, LineParams{segmentNo, tstart, tend,
		pstartx, pstarty, pendx, pendy})
}
