package bendit

import (
	"errors"
	"fmt"
	"math"
)

type Knots interface {
	IsUniform() bool
	Count() int
	Tstart() float64
	Tend() float64
	Knot(knotNo int) (t float64, err error)
	SegmentLen(segmentNo int) (t float64, err error)
	MapToSegment(t float64) (segmentNo int, u float64, err error)
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

func (k *UniformKnots) Count() int {
	return k.cnt
}

func (k *UniformKnots) Knot(knotNo int) (t float64, err error) {
	// TODO assert knotNo <= segmCnt
	if knotNo < 0 || knotNo >= k.cnt {
		return 0, errors.New("knot doesn't exist")
	} else {
		return float64(knotNo), nil
	}
}

func (k *UniformKnots) SegmentLen(segmentNo int) (t float64, err error) {
	if segmentNo < 0 || segmentNo >= k.cnt-1 {
		return 0, errors.New("segment doesn't exist")
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

func (k *UniformKnots) Add(segmentLen float64) error {
	if segmentLen != 1 {
		return errors.New("cannot add length != 1 to uniform knots")
	}
	k.cnt += 1
	return nil
}

type NonUniformKnots struct {
	ks []float64
}

func NewNonUniformKnots(knots []float64) *NonUniformKnots {
	// TODO validate knots: monotonically increasing
	return &NonUniformKnots{knots}
}

func (k *NonUniformKnots) IsUniform() bool {
	return false
}

func (k *NonUniformKnots) Tstart() float64 {
	return 0
}

func (k *NonUniformKnots) Tend() float64 {
	if len(k.ks) == 0 {
		return 0
	} else {
		return k.ks[len(k.ks)-1]
	}
}

func (k *NonUniformKnots) Count() int {
	return len(k.ks)
}

func (k *NonUniformKnots) Knot(knotNo int) (t float64, err error) {
	// TODO assert knotNo <= segmCnt
	if knotNo < 0 || knotNo >= len(k.ks) {
		return 0, errors.New("knot doesn't exist")
	} else {
		return k.ks[knotNo], nil
	}
}

func (k *NonUniformKnots) SegmentLen(segmentNo int) (t float64, err error) {
	if segmentNo < 0 || segmentNo >= len(k.ks)-1 {
		return 0, errors.New("segment doesn't exist")
	} else {
		return k.ks[segmentNo+1] - k.ks[segmentNo], nil
	}
}

func (k *NonUniformKnots) MapToSegment(t float64) (segmentNo int, u float64, err error) {
	segmCnt := len(k.ks) - 1
	if segmCnt < 1 {
		err = errors.New("at least one segment having 2 knots required")
		return
	}
	if t < k.ks[0] {
		err = fmt.Errorf("%v smaller than first knot %v", t, k.ks[0])
		return
	}

	// TODO speed up mapping
	for i := 0; i < segmCnt; i++ {
		if t <= k.ks[i+1] {
			if k.ks[i+1] == k.ks[i] {
				u = 0
			} else {
				u = (t - k.ks[i]) / (k.ks[i+1] - k.ks[i])
			}
			return i, u, nil
		}
	}
	err = fmt.Errorf("%v greater than upper limit %v", t, k.ks[segmCnt])
	return
}

func (k *NonUniformKnots) Add(segmentLen float64) {
	k.ks = append(k.ks, k.Tend()+segmentLen)
}
