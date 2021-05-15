package bendit

import (
	"errors"
	"fmt"
	"math"
)

type Knots struct {
	ks []float64
	// TODO SegmentCnt for uniform
}

func NewUniformKnots() *Knots {
	return &Knots{ks: nil}
}

func NewKnots(ks []float64) *Knots {
	return &Knots{ks: ks}
}

func (k *Knots) Count() int {
	return len(k.ks)
}

func (k *Knots) IsUniform() bool {
	return len(k.ks) == 0
}

func (k *Knots) Domain(segmCnt int) SplineDomain {
	if k.IsUniform() {
		return SplineDomain{Start: 0, End: float64(segmCnt)}
	} else {
		return SplineDomain{Start: 0, End: k.ks[len(k.ks)-1]}
	}
}

func (k *Knots) SegmentLength(segmNo int) float64 {
	if k.IsUniform() {
		return 1
	} else {
		return k.ks[segmNo+1] - k.ks[segmNo]
	}
}

func (k *Knots) SegmentRange(segmNo int) (start, end float64) {
	if k.IsUniform() {
		return float64(segmNo), float64(segmNo + 1)
	} else {
		return k.ks[segmNo], k.ks[segmNo+1]
	}
}

func (k *Knots) MapToSegment(t float64, segmCnt int) (segmNo int, u float64, err error) {
	if k.IsUniform() {
		return k.mapUniToSegment(t, segmCnt)
	} else {
		return k.mapNonUniToSegment(t)
	}
}

func (k *Knots) mapUniToSegment(t float64, segmCnt int) (segmNo int, u float64, err error) {
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
		segmNo = segmCnt - 1
		u = 1
	} else {
		segmNo = int(ifl)
	}
	return
}

func (k *Knots) mapNonUniToSegment(t float64) (segmNo int, u float64, err error) {
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
