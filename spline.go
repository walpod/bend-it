package bendigo

type Spline interface {
	Knots() Knots
	At(t float64) Vec
}

type SplineBuilder interface {
	Knots() Knots
	Build() Spline
	// TODO LinaxSpline(linaxParams LinaxParams) *LinaxSpline
	// linear approximate spline with consecutive line segments
	Linax(fromSegmentNo, toSegmentNo int, collector LineCollector, linaxParams *LinaxParams)
}

// linear approximation parameters
type LinaxParams struct {
	MaxDist float64
}

func NewLinaxParams(maxDist float64) *LinaxParams {
	return &LinaxParams{MaxDist: maxDist}
}

type Vertex interface {
	Loc() Vec
}

// SplineVertBuilder is constructed by adding vertices
type SplineVertBuilder interface {
	SplineBuilder

	Vertex(knotNo int) Vertex
	AddVertex(knotNo int, vertex Vertex) (err error)
	UpdateVertex(knotNo int, vertex Vertex) (err error)
	DeleteVertex(knotNo int) (err error)
}

// TODO replace with LinaxSpline
func LinaxAll(splineBuilder SplineBuilder, collector LineCollector, linaxParams *LinaxParams) {
	splineBuilder.Linax(0, splineBuilder.Knots().SegmentCnt()-1, collector, linaxParams)
}

func Vertices(builder SplineVertBuilder) []Vertex {
	cnt := builder.Knots().KnotCnt()
	vertices := make([]Vertex, cnt)
	for i := 0; i < cnt; i++ {
		vertices[i] = builder.Vertex(i)
	}
	return vertices
}
