package cubic

import "github.com/walpod/bendigo"

// CardinalVertBuilder is an hermite vertex-based builder for cardinal splines
type CardinalVertBuilder struct {
	HermiteVertBuilder
	tension float64
}

func NewCardinalVertBuilder(tknots []float64, tension float64, vertices ...*HermiteVertex) *CardinalVertBuilder {
	sb := &CardinalVertBuilder{
		HermiteVertBuilder: *NewHermiteVertBuilder(tknots, vertices...),
		tension:            tension}
	sb.CalcTangents()
	return sb
}

func (sb *CardinalVertBuilder) Tension() float64 {
	return sb.tension
}

func (sb *CardinalVertBuilder) SetTension(tension float64) {
	sb.tension = tension
	sb.CalcTangents()
}

func (sb *CardinalVertBuilder) AddVertex(knotNo int, vertex bendigo.Vertex) (err error) {
	err = sb.HermiteVertBuilder.AddVertex(knotNo, vertex)
	if err == nil {
		sb.CalcTangents() // TODO recalculate only around new knot
	}
	return err
}

func (sb *CardinalVertBuilder) UpdateVertex(knotNo int, vertex bendigo.Vertex) (err error) {
	err = sb.HermiteVertBuilder.UpdateVertex(knotNo, vertex)
	if err == nil {
		sb.CalcTangents() // TODO recalculate only around updated knot
	}
	return err
}

func (sb *CardinalVertBuilder) DeleteVertex(knotNo int) (err error) {
	err = sb.HermiteVertBuilder.DeleteVertex(knotNo)
	if err == nil {
		sb.CalcTangents() // TODO recalculate only around deleted knot
	}
	return err
}

// CalcTangents calculates and sets the tangent controls of the hermite vertices
func (sb *CardinalVertBuilder) CalcTangents() {
	n := len(sb.vertices)
	if n < 2 {
		return
	}
	dim := sb.vertices[0].loc.Dim()

	// transform tension to 'scale' factor of distance vector
	scale := (1 - sb.tension) / 2

	// calculate tangents for uniform case: entry and exit tangents are equal
	setUniformCardinalTangent := func(vt, vtstart, vtend *HermiteVertex) {
		tan := bendigo.NewZeroVec(vt.loc.Dim())
		for d := 0; d < dim; d++ {
			tan[d] = scale * (vtend.loc[d] - vtstart.loc[d])
		}
		vt.entry, vt.exit = tan, tan // TODO or clone ?
	}

	setUniformCardinalTangent(sb.vertices[0], sb.vertices[0], sb.vertices[1])
	for i := 1; i < n-1; i++ {
		setUniformCardinalTangent(sb.vertices[i], sb.vertices[i-1], sb.vertices[i+1]) // use vertex before and after
	}
	setUniformCardinalTangent(sb.vertices[n-1], sb.vertices[n-2], sb.vertices[n-1])

	// handle non-uniform case: double tangent, same direction but different lengths
	if !sb.knots.IsUniform() {
		for i := 0; i < n-1; i++ {
			// modify length of uniform tangents according to segment-length
			segmentLen, _ := sb.knots.SegmentLen(i)
			if segmentLen != 0 {
				scf := 1 / segmentLen
				sb.vertices[i].exit = sb.vertices[i].exit.Scale(scf)
				sb.vertices[i+1].entry = sb.vertices[i+1].entry.Scale(scf)
			}
			// TODO segmentLen == 0
		}
	}
}

// NewCatmullRomVertBuilder creates a special cardinal builder with tension = 0
func NewCatmullRomVertBuilder(tknots []float64, vertices ...*HermiteVertex) *CardinalVertBuilder {
	return NewCardinalVertBuilder(tknots, 0, vertices...)
}
