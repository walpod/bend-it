package cubic

import "github.com/walpod/bend-it"

// Hermite tangent finder for cardinal spline
type CardinalTanf2d struct {
	tension float64
}

func (ct CardinalTanf2d) Find(knots *bendit.Knots, verts []*HermiteVertex2d) {
	n := len(verts)
	if n < 2 {
		return
	}

	// transform tension to 'scale' factor of distance vector
	b := (1 - ct.tension) / 2

	// calculate tangents for uniform case: entry and exit tangents are equal
	setUniformCardinalTan := func(v, vstart, vend *HermiteVertex2d) {
		tanx := b * (vend.x - vstart.x)
		tany := b * (vend.y - vstart.y)
		v.entryTanx, v.exitTanx = tanx, tanx
		v.entryTany, v.exitTany = tany, tany
	}

	setUniformCardinalTan(verts[0], verts[0], verts[1])
	for i := 1; i < n-1; i++ {
		setUniformCardinalTan(verts[i], verts[i-1], verts[i+1]) // use vertex before and after
	}
	setUniformCardinalTan(verts[n-1], verts[n-2], verts[n-1])

	// handle non-uniform case: double tangent, same direction but different lengths
	if !knots.IsUniform() {
		for i := 0; i < n-1; i++ {
			// modify length of uniform tangents according to segment-length
			segmLen := knots.SegmentLength(i)
			verts[i].exitTanx /= segmLen // TODO segmLen == 0
			verts[i].exitTany /= segmLen
			verts[i+1].entryTanx /= segmLen
			verts[i+1].entryTany /= segmLen
		}
	}
}

type CardinalHermiteSpline2d struct {
	HermiteSpline2d
	tension float64
}

func NewCardinalHermiteSpline2d(knots *bendit.Knots, tension float64, verts ...*HermiteVertex2d) *CardinalHermiteSpline2d {
	cs := &CardinalHermiteSpline2d{
		HermiteSpline2d: *NewHermiteSplineTanFinder2d(knots, CardinalTanf2d{tension: tension}, verts...),
		tension:         tension}
	cs.Build() // TODO don't build automatically
	return cs
}

func NewCatmullRomHermiteSpline2d(knots *bendit.Knots, verts ...*HermiteVertex2d) *CardinalHermiteSpline2d {
	return NewCardinalHermiteSpline2d(knots, 0, verts...)
}

func (cs *CardinalHermiteSpline2d) Tension() float64 {
	return cs.tension
}

func (cs *CardinalHermiteSpline2d) SetTension(tension float64) {
	cs.HermiteSpline2d.tanFinder = CardinalTanf2d{tension: tension}
	cs.tension = tension
	cs.Build() // TODO don't build automatically
}
