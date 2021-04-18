package cubic

import (
	"errors"
	"fmt"
	bendit "github.com/walpod/bend-it"
	"math"
)

// cubic polynomial
type CubicPoly struct {
	a, b, c, d float64
}

func NewCubicPoly(a float64, b float64, c float64, d float64) CubicPoly {
	return CubicPoly{a: a, b: b, c: c, d: d}
}

func (cb *CubicPoly) At(u float64) float64 {
	return cb.a + u*(cb.b+u*(cb.c+cb.d*u))
}

func (cb *CubicPoly) Fn() func(float64) float64 {
	return func(u float64) float64 {
		return cb.At(u)
	}
}

type Cubic2d struct {
	// TODO maybe use instead 2x4 matrix and matrix multiplication
	cubx, cuby CubicPoly
}

func NewCubic2d(cubx CubicPoly, cuby CubicPoly) Cubic2d {
	return Cubic2d{cubx: cubx, cuby: cuby}
}

func (cb *Cubic2d) At(u float64) (x, y float64) {
	return cb.cubx.At(u), cb.cuby.At(u)
}

func (cb *Cubic2d) Fn() bendit.Fn2d {
	return func(u float64) (x, y float64) {
		return cb.At(u)
	}
}

type CanonicalSpline2d struct {
	cubics []Cubic2d
	knots  []float64
}

func NewCanonicalSpline2d(cubics []Cubic2d, knots []float64) *CanonicalSpline2d {
	if len(knots) > 0 && len(knots) != len(cubics)+1 {
		panic("knots must be empty or having length of cubics + 1")
	}
	return &CanonicalSpline2d{cubics: cubics, knots: knots}
}

func (cs *CanonicalSpline2d) SegmentCnt() int {
	return len(cs.cubics)
}

func (cs *CanonicalSpline2d) Domain() (fr, to float64) {
	if cs.knots == nil {
		to = float64(cs.SegmentCnt())
	} else {
		to = cs.knots[len(cs.knots)-1]
	}
	return 0, to
}

func (cs *CanonicalSpline2d) At(t float64) (x, y float64) {
	if len(cs.cubics) == 0 {
		return 0, 0 // TODO or panic? or error?
	}

	var (
		segmNo int
		u      float64
		err    error
	)
	if len(cs.knots) == 0 {
		segmNo, u, err = cs.mapUniToSegm(t)
	} else {
		segmNo, u, err = cs.mapNonUniToSegm(t)
	}
	if err != nil {
		return 0, 0 // TODO or panic? or error?
	} else {
		return cs.cubics[segmNo].At(u)
	}
}

func (cs *CanonicalSpline2d) mapUniToSegm(t float64) (segmNo int, u float64, err error) {
	segmCnt := cs.SegmentCnt()
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

func (cs *CanonicalSpline2d) mapNonUniToSegm(t float64) (segmNo int, u float64, err error) {
	segmCnt := len(cs.knots) - 1
	if segmCnt < 1 {
		err = errors.New("at least one segment having 2 knots required")
		return
	}
	if t < cs.knots[0] {
		err = fmt.Errorf("%v smaller than first knot %v", t, cs.knots[0])
		return
	}

	// TODO speed up mapping
	for i := 0; i < segmCnt; i++ {
		if t <= cs.knots[i+1] {
			return i, (t - cs.knots[i]) / (cs.knots[i+1] - cs.knots[i]), nil
		}
	}
	err = fmt.Errorf("%v greater than upper limit %v", t, cs.knots[segmCnt+1])
	return
}

func (cs *CanonicalSpline2d) Fn() bendit.Fn2d {
	return func(t float64) (x, y float64) {
		return cs.At(t)
	}
}
