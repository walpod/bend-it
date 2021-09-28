package bendit

import "math"

type Vec []float64

func NewZeroVec(dim int) Vec {
	return make(Vec, dim)
}

func (v Vec) Dim() int {
	return len(v)
}

func (v Vec) Add(w Vec) Vec {
	dim := v.Dim()
	r := make([]float64, dim)
	for d := 0; d < dim; d++ {
		r[d] = v[d] + w[d]
	}
	return r
}

func (v Vec) Sub(w Vec) Vec {
	dim := v.Dim()
	r := make([]float64, dim)
	for d := 0; d < dim; d++ {
		r[d] = v[d] - w[d]
	}
	return r
}

func (v Vec) Len() float64 {
	vl := 0.
	for d := 0; d < len(v); d++ {
		vl += v[d] * v[d]
	}
	return math.Sqrt(vl)
}

// calculate distance of vector v to projected vector v on w
func (v Vec) ProjectedVecDist(w Vec) float64 {
	// distance = area of parallelogram(v, w) / length(w)
	var area float64
	if len(v) == 2 {
		area = math.Abs(w[0]*v[1] - w[1]*v[0])
	} else if len(v) == 3 {
		area = Vec{v[1]*w[2] - v[2]*w[1], v[2]*w[0] - v[0]*w[2], v[0]*w[1] - v[1]*w[0]}.Len()
	} else {
		panic("ProjectedVecDist not yet implemented for dim >= 4")
	}
	return area / w.Len()
}
