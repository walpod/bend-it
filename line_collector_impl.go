package bendit

// DirectCollector2d supports the simple case of using a single collect func
type DirectCollector2d struct {
	line func(tstart, tend, pstartx, pstarty, pendx, pendy float64)
}

func NewDirectCollector2d(line func(tstart, tend, pstartx, pstarty, pendx, pendy float64)) *DirectCollector2d {
	return &DirectCollector2d{line: line}
}

func (lc DirectCollector2d) CollectLine(tstart, tend, pstartx, pstarty, pendx, pendy float64) {
	lc.line(tstart, tend, pstartx, pstarty, pendx, pendy)
}

// LineToSliceCollector2d collects line in slice
type LineToSliceCollector2d struct {
	Lines []LineParams
}

type LineParams struct {
	Tstart, Tend, Pstartx, Pstarty, Pendx, Pendy float64
}

func NewLineToSliceCollector2d() *LineToSliceCollector2d {
	return &LineToSliceCollector2d{Lines: make([]LineParams, 0)}
}

func (lc *LineToSliceCollector2d) CollectLine(tstart, tend, pstartx, pstarty, pendx, pendy float64) {
	lc.Lines = append(lc.Lines, LineParams{tstart, tend, pstartx, pstarty, pendx, pendy})
}
