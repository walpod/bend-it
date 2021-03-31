package cubspl

type Vertex2d interface {
	Point() (px, py float64)
}

type HermiteVertex2d interface {
	Vertex2d
	// entry and exit tangents
	EntryTan() (lx, ly float64)
	ExitTan() (mx, my float64)
}

type SingleTanH2d struct {
	px, py float64
	mx, my float64
}

func NewSingleTanH2d(px float64, py float64, mx float64, my float64) *SingleTanH2d {
	return &SingleTanH2d{px: px, py: py, mx: mx, my: my}
}

func (st *SingleTanH2d) Point() (px, py float64) {
	return st.px, st.py
}

func (st *SingleTanH2d) EntryTan() (lx, ly float64) {
	// entry = exit tangent
	return st.mx, st.my
}

func (st *SingleTanH2d) ExitTan() (mx, my float64) {
	return st.mx, st.my
}
