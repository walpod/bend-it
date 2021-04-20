package cubic

import (
	bendit "github.com/walpod/bend-it"
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
	knots  bendit.Knots
}

func NewCanonicalSpline2d(cubics []Cubic2d, knots bendit.Knots) *CanonicalSpline2d {
	if knots.Count() > 0 && knots.Count() != len(cubics)+1 {
		panic("knots must be empty or having length of cubics + 1")
	}
	return &CanonicalSpline2d{cubics: cubics, knots: knots}
}

func (cs *CanonicalSpline2d) SegmentCnt() int {
	return len(cs.cubics)
}

func (cs *CanonicalSpline2d) Domain() bendit.SplineDomain {
	return cs.knots.Domain(cs.SegmentCnt())
}

func (cs *CanonicalSpline2d) At(t float64) (x, y float64) {
	if len(cs.cubics) == 0 {
		return 0, 0 // TODO or panic? or error?
	}

	segmNo, u, err := cs.knots.MapToSegment(t, cs.SegmentCnt())
	if err != nil {
		return 0, 0 // TODO or panic? or error?
	} else {
		return cs.cubics[segmNo].At(u)
	}
}

func (cs *CanonicalSpline2d) Fn() bendit.Fn2d {
	return func(t float64) (x, y float64) {
		return cs.At(t)
	}
}
