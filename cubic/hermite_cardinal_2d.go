package cubic

import "github.com/walpod/bend-it"

// hermite tangent finder for cardinal spline
type CardinalTanf2d struct {
	tension float64
}

func NewCardinalTanf2d(tension float64) CardinalTanf2d {
	return CardinalTanf2d{tension: tension}
}

func NewCatmullRomTanf2d() CardinalTanf2d {
	return NewCardinalTanf2d(0)
}

func (ct CardinalTanf2d) Find(knots *bendit.Knots, verts []*HermiteVertex2d) {
	n := len(verts)
	if n < 2 {
		return
	}

	// calculate tangents for uniform case: entry and exit tangents are equal
	b := (1 - ct.tension) / 2
	setUniformCardinalTan := func(vert *HermiteVertex2d, xstart, xend, ystart, yend float64) {
		tanx := b * (xend - xstart)
		tany := b * (yend - ystart)
		vert.entryTanx, vert.exitTanx = tanx, tanx
		vert.entryTany, vert.exitTany = tany, tany
	}

	setUniformCardinalTan(verts[0], verts[0].x, verts[1].x, verts[0].y, verts[1].y)
	for i := 1; i < n-1; i++ {
		setUniformCardinalTan(verts[i], verts[i-1].x, verts[i+1].x, verts[i-1].y, verts[i+1].y) // use vertex before and after
	}
	setUniformCardinalTan(verts[n-1], verts[n-2].x, verts[n-1].x, verts[n-2].y, verts[n-1].y)
	/*exitTansx[0] = b * (vertsx[1] - vertsx[0])
	exitTansy[0] = b * (vertsy[1] - vertsy[0])
	for i := 1; i < n-1; i++ {
		exitTansx[i] = b * (vertsx[i+1] - vertsx[i-1])
		exitTansy[i] = b * (vertsy[i+1] - vertsy[i-1])
	}
	exitTansx[n-1] = b * (vertsx[n-1] - vertsx[n-2])
	exitTansy[n-1] = b * (vertsy[n-1] - vertsy[n-2])*/

	// handle non-uniform case: double tangent, same direction but different lengths
	if !knots.IsUniform() {
		for i := 0; i < n-1; i++ {
			// modify length of uniform tangents according to segment-length
			segmLen := knots.SegmentLength(i)
			verts[i].exitTanx /= segmLen // TODO segmLen = 0
			verts[i].exitTany /= segmLen
			verts[i+1].entryTanx /= segmLen
			verts[i+1].entryTany /= segmLen
		}
	}
}
