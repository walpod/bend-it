package bendit

import (
	"errors"
	"fmt"
	"math"
)

type Knots interface {
	IsUniform() bool
	Tstart() float64
	Tend() float64

	KnotCnt() int
	KnotExists(knotNo int) bool
	Knot(knotNo int) (t float64, err error)
	AddKnot(knotNo int) (err error)
	DeleteKnot(knotNo int) (err error)

	SegmentCnt() int
	SegmentExists(segmentNo int) bool
	SegmentLen(segmentNo int) (l float64, err error)
	SetSegmentLen(segmentNo int, l float64) (err error)
	MapToSegment(t float64) (segmentNo int, u float64, err error)

	External() []float64 // external representation: uniform = nil, non-uniform = slice (non nil)
}

type UniformKnots struct {
	cnt int // number of knots
}

func NewUniformKnots(knotsCnt int) *UniformKnots {
	return &UniformKnots{cnt: knotsCnt}
}

func (k *UniformKnots) IsUniform() bool {
	return true
}

func (k *UniformKnots) Tstart() float64 {
	return 0
}

func (k *UniformKnots) Tend() float64 {
	return float64(k.cnt - 1)
}

func (k *UniformKnots) KnotCnt() int {
	return k.cnt
}

func (k *UniformKnots) KnotExists(knotNo int) bool {
	return knotNo >= 0 && knotNo < k.cnt
}

func (k *UniformKnots) Knot(knotNo int) (t float64, err error) {
	if !k.KnotExists(knotNo) {
		return 0, fmt.Errorf("knot with no. %v doesn't exist", knotNo)
	} else {
		return float64(knotNo), nil
	}
}

func (k *UniformKnots) AddKnot(knotNo int) (err error) {
	if !(knotNo >= 0 && knotNo <= k.cnt) {
		return fmt.Errorf("knot must be added at existing knot or at end, incorrect knot no. %v", knotNo)
	} else {
		k.cnt += 1
		return nil
	}
}

func (k *UniformKnots) DeleteKnot(knotNo int) (err error) {
	if !k.KnotExists(knotNo) {
		return fmt.Errorf("knot with no. %v doesn't exist", knotNo)
	} else {
		k.cnt -= 1
		return nil
	}
}

func (k *UniformKnots) SegmentCnt() int {
	sc := k.cnt - 1
	if sc < 0 {
		return 0
	} else {
		return sc
	}
}

func (k *UniformKnots) SegmentExists(segmentNo int) bool {
	return segmentNo >= 0 && segmentNo < k.cnt-1
}

func (k *UniformKnots) SegmentLen(segmentNo int) (l float64, err error) {
	if !k.SegmentExists(segmentNo) {
		return 0, fmt.Errorf("segment with no. %v doesn't exist", segmentNo)
	} else {
		return 1, nil
	}
}

// SetSegmentLen sets length of segment, which must be 1 in uniform case
func (k *UniformKnots) SetSegmentLen(segmentNo int, l float64) (err error) {
	if !k.SegmentExists(segmentNo) {
		return fmt.Errorf("segment with no. %v doesn't exist", segmentNo)
	} else if l != 1 {
		return fmt.Errorf("segment length %v not correct for uniform knots, must be 1", l)
	} else {
		return nil
	}
}

func (k *UniformKnots) MapToSegment(t float64) (segmentNo int, u float64, err error) {
	tend := k.Tend()
	if t < 0 {
		err = fmt.Errorf("%v smaller than 0", t)
		return
	}
	if t > tend {
		err = fmt.Errorf("%v greater than last knot %v", t, tend)
		return
	}

	var ifl float64
	ifl, u = math.Modf(t)
	segmentNo = int(ifl)

	// special case t == tend
	if ifl == tend {
		segmentNo -= 1
		u = 1
	}
	return
}

func (k *UniformKnots) External() []float64 {
	return nil
}

type NonUniformKnots struct {
	tknots []float64
}

func NewNonUniformKnots(tknots []float64) *NonUniformKnots {
	// TODO validate knots: monotonically increasing
	return &NonUniformKnots{tknots}
}

func (k *NonUniformKnots) IsUniform() bool {
	return false
}

func (k *NonUniformKnots) Tstart() float64 {
	return 0
}

func (k *NonUniformKnots) Tend() float64 {
	if len(k.tknots) == 0 {
		return 0
	} else {
		return k.tknots[len(k.tknots)-1]
	}
}

func (k *NonUniformKnots) KnotCnt() int {
	return len(k.tknots)
}

func (k *NonUniformKnots) KnotExists(knotNo int) bool {
	return knotNo >= 0 && knotNo < len(k.tknots)
}

func (k *NonUniformKnots) Knot(knotNo int) (t float64, err error) {
	if !k.KnotExists(knotNo) {
		return 0, fmt.Errorf("knot with no. %v doesn't exist", knotNo)
	} else {
		return k.tknots[knotNo], nil
	}
}

func (k *NonUniformKnots) AddKnot(knotNo int) (err error) {
	if !(knotNo >= 0 && knotNo <= len(k.tknots)) {
		return fmt.Errorf("knot must be added at existing knot or at end, incorrect knot no. %v", knotNo)
	}
	if len(k.tknots) == 0 {
		k.tknots = append(k.tknots, 0) // add first knot with value 0
	} else if knotNo < len(k.tknots) {
		k.tknots = append(k.tknots[:knotNo+1], k.tknots[knotNo:]...)
	} else {
		k.tknots = append(k.tknots, k.tknots[len(k.tknots)-1])
	}
	return nil
}

func (k *NonUniformKnots) DeleteKnot(knotNo int) (err error) {
	if !k.KnotExists(knotNo) {
		return fmt.Errorf("knot with no. %v doesn't exist", knotNo)
	} else {
		k.tknots = append(k.tknots[:knotNo], k.tknots[knotNo+1:]...)
		return nil
	}
}

func (k *NonUniformKnots) SegmentCnt() int {
	sc := len(k.tknots) - 1
	if sc < 0 {
		return 0
	} else {
		return sc
	}
}

func (k *NonUniformKnots) SegmentExists(segmentNo int) bool {
	return segmentNo >= 0 && segmentNo < len(k.tknots)-1
}

func (k *NonUniformKnots) SegmentLen(segmentNo int) (l float64, err error) {
	if !k.SegmentExists(segmentNo) {
		return 0, fmt.Errorf("segment with no. %v doesn't exist", segmentNo)
	} else {
		return k.tknots[segmentNo+1] - k.tknots[segmentNo], nil
	}
}

func (k *NonUniformKnots) SetSegmentLen(segmentNo int, l float64) (err error) {
	if !k.SegmentExists(segmentNo) {
		return fmt.Errorf("segment with no. %v doesn't exist", segmentNo)
	} else {
		diff := l - (k.tknots[segmentNo+1] - k.tknots[segmentNo])
		for i := segmentNo + 1; i < len(k.tknots); i++ {
			k.tknots[i] += diff
		}
		return nil
	}
}

func (k *NonUniformKnots) MapToSegment(t float64) (segmentNo int, u float64, err error) {
	segmCnt := len(k.tknots) - 1
	if segmCnt < 1 {
		err = errors.New("at least one segment having 2 knots required")
		return
	}
	if t < k.tknots[0] {
		err = fmt.Errorf("%v smaller than first knot %v", t, k.tknots[0])
		return
	}

	// TODO speed up mapping
	for i := 0; i < segmCnt; i++ {
		if t < k.tknots[i+1] {
			if k.tknots[i+1] == k.tknots[i] {
				u = 0
			} else {
				u = (t - k.tknots[i]) / (k.tknots[i+1] - k.tknots[i])
			}
			return i, u, nil
		}
		if t == k.Tend() { // TODO within-delta
			return segmCnt - 1, 1, nil
		}
	}
	err = fmt.Errorf("%v greater than upper limit %v", t, k.tknots[segmCnt])
	return
}

func (k *NonUniformKnots) External() []float64 {
	xtknots := make([]float64, len(k.tknots))
	copy(xtknots, k.tknots)
	return xtknots
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
