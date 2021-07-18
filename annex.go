package bendit

import "fmt"

type Annex struct {
	knots      Knots
	toVertices []interface{}
	toSegments []interface{}
}

func NewAnnex(knots Knots) *Annex {
	return &Annex{knots: knots}
}

func (an *Annex) ToVertex(knotNo int) (data interface{}) {
	if !an.knots.KnotExists(knotNo) || knotNo >= len(an.toVertices) {
		return nil
	} else {
		return an.toVertices[knotNo]
	}
}

func (an *Annex) AttachToVertex(knotNo int, data interface{}) (err error) {
	if !an.knots.KnotExists(knotNo) {
		return fmt.Errorf("knot with no. %v doesn't exist", knotNo)
	} else {
		if knotNo >= len(an.toVertices) {
			an.toVertices = append(an.toVertices, make([]interface{}, an.knots.Count()-len(an.toVertices)))
		}
		an.toVertices[knotNo] = data
		return nil
	}
}

// TODO ToSegment, AttachToSegment
