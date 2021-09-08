package bendit

import (
	"errors"
	"fmt"
	"math"
)

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

func (k *UniformKnots) Cnt() int {
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

func (k *UniformKnots) Add(segmentLen float64) error {
	if segmentLen != 1 {
		return errors.New("cannot add length != 1 to uniform knots")
	}
	k.cnt += 1
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

func (k *NonUniformKnots) Cnt() int {
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

func (k *NonUniformKnots) Add(segmentLen float64) {
	k.tknots = append(k.tknots, k.Tend()+segmentLen)
}
