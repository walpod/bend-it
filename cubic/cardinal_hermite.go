package cubic

import "github.com/walpod/bend-it"

// CardinalTanFinder is an Hermite tangent finder for cardinal splines
type CardinalTanFinder struct {
	tension float64
}

func (ct CardinalTanFinder) Find(knots bendit.Knots, vertices []*HermiteVertex) {
	n := len(vertices)
	if n < 2 {
		return
	}

	dim := vertices[0].loc.Dim() // precondition: len(vertices) >= 1

	// transform tension to 'scale' factor of distance vector
	b := (1 - ct.tension) / 2

	// calculate tangents for uniform case: entry and exit tangents are equal
	setUniformCardinalTan := func(vt, vtstart, vtend *HermiteVertex) {
		tan := bendit.NewZeroVec(vt.loc.Dim())
		for d := 0; d < dim; d++ {
			tan[d] = b * (vtend.loc[d] - vtstart.loc[d])
		}
		vt.entry, vt.exit = tan, tan // TODO or clone ?
	}

	setUniformCardinalTan(vertices[0], vertices[0], vertices[1])
	for i := 1; i < n-1; i++ {
		setUniformCardinalTan(vertices[i], vertices[i-1], vertices[i+1]) // use vertex before and after
	}
	setUniformCardinalTan(vertices[n-1], vertices[n-2], vertices[n-1])

	// handle non-uniform case: double tangent, same direction but different lengths
	if !knots.IsUniform() {
		for i := 0; i < n-1; i++ {
			// modify length of uniform tangents according to segment-length
			segmentLen, _ := knots.SegmentLen(i)
			/*vertices[i].exitTan.x /= segmentLen
			vertices[i].exitTan.y /= segmentLen
			vertices[i+1].entryTan.x /= segmentLen
			vertices[i+1].entryTan.y /= segmentLen*/
			if segmentLen != 0 {
				scf := 1 / segmentLen
				vertices[i].exit = vertices[i].exit.Scale(scf)
				vertices[i+1].entry = vertices[i+1].entry.Scale(scf)
			}
			// TODO segmentLen == 0
		}
	}
}

// CardinalVertBuilder is an hermite vertex-based builder for cardinal splines
type CardinalVertBuilder struct {
	HermiteVertBuilder
	tension float64
}

func NewCardinalVertBuilder(tknots []float64, tension float64, vertices ...*HermiteVertex) *CardinalVertBuilder {
	sp := &CardinalVertBuilder{
		HermiteVertBuilder: *NewHermiteVertBuilderTanFinder(tknots, CardinalTanFinder{tension: tension}, vertices...),
		tension:            tension}
	return sp
}

// NewCatmullRomVertBuilder creates a special cardinal builder with tension = 0
func NewCatmullRomVertBuilder(tknots []float64, vertices ...*HermiteVertex) *CardinalVertBuilder {
	return NewCardinalVertBuilder(tknots, 0, vertices...)
}

func (sp *CardinalVertBuilder) Tension() float64 {
	return sp.tension
}

func (sp *CardinalVertBuilder) SetTension(tension float64) {
	sp.HermiteVertBuilder.tanFinder = CardinalTanFinder{tension: tension}
	sp.tension = tension
}
