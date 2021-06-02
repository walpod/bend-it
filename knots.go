package bendit

import (
	"errors"
	"fmt"
	"math"
)

type Knots interface {
	IsUniform() bool
	Domain() SplineDomain
	Count() int
	Knot(knotNo int) (t float64, err error)
	SegmentLen(segmentNo int) (t float64, err error)
	MapToSegment(t float64) (segmentNo int, u float64, err error)
}

// range of parameter t for which the spline is defined
type SplineDomain struct {
	Start, End float64
}

type UniformKnots struct {
	Spline Spline2d
}

func NewUniformKnots() *UniformKnots {
	// TODO validate ks: monotonically increasing
	return &UniformKnots{}
}

func (k *UniformKnots) IsUniform() bool {
	return true
}

func (k *UniformKnots) Domain() SplineDomain {
	return SplineDomain{Start: 0, End: float64(k.Spline.SegmentCnt())}
}

func (k *UniformKnots) Count() int {
	return k.Spline.SegmentCnt() + 1
}

func (k *UniformKnots) Knot(knotNo int) (t float64, err error) {
	// TODO assert knotNo <= segmCnt
	if knotNo < 0 || knotNo >= k.Count() {
		return 0, errors.New("knot doesn't exist")
	} else {
		return float64(knotNo), nil
	}
}

func (k *UniformKnots) SegmentLen(segmentNo int) (t float64, err error) {
	if segmentNo < 0 || segmentNo >= k.Count()-1 {
		return 0, errors.New("segment doesn't exist")
	} else {
		return 1, nil
	}
}

func (k *UniformKnots) MapToSegment(t float64) (segmentNo int, u float64, err error) {
	segmCnt := k.Spline.SegmentCnt()
	upper := float64(segmCnt)
	if t < 0 {
		err = fmt.Errorf("%v smaller than 0", t)
		return
	}
	if t > upper {
		err = fmt.Errorf("%v greater than last knot %v", t, upper)
		return
	}

	var ifl float64
	ifl, u = math.Modf(t)
	if ifl == upper {
		// special case t == upper
		segmentNo = segmCnt - 1
		u = 1
	} else {
		segmentNo = int(ifl)
	}
	return
}

func (k *UniformKnots) SetSplineIfEmpty(spline Spline2d) {
	if k.Spline == nil {
		k.Spline = spline
	}
}

type NonUniformKnots struct {
	ks []float64
}

func NewNonUniformKnots(ks []float64) *NonUniformKnots {
	// TODO validate ks: monotonically increasing
	return &NonUniformKnots{ks}
}

func (k *NonUniformKnots) IsUniform() bool {
	return false
}

func (k *NonUniformKnots) Domain() SplineDomain {
	return SplineDomain{Start: 0, End: k.ks[len(k.ks)-1]}
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
