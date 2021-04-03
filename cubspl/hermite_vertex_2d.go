package cubspl

// entry and exit tangents for given vertex
type VertexTan2d interface {
	EntryTan() (lx, ly float64)
	ExitTan() (mx, my float64)
}

type SingleTan2d struct {
	Mx, My float64
}

func NewSingleTan2d(mx float64, my float64) *SingleTan2d {
	return &SingleTan2d{Mx: mx, My: my}
}

func (st *SingleTan2d) EntryTan() (lx, ly float64) {
	// entry = exit tangent
	return st.Mx, st.My
}

func (st *SingleTan2d) ExitTan() (mx, my float64) {
	return st.Mx, st.My
}
