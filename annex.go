package bendigo

import "fmt"

type Annex struct {
	knots      Knots
	toVertices []interface{}
	toSegments []interface{}
}

func NewAnnex(knots Knots) *Annex {
	return &Annex{knots: knots}
}

func (an *Annex) AttachToVertex(knotNo int, data interface{}) (err error) {
	if !an.knots.KnotExists(knotNo) {
		return fmt.Errorf("knot with no. %v doesn't exist", knotNo)
	} else {
		if knotNo >= len(an.toVertices) {
			an.toVertices = append(an.toVertices, make([]interface{}, knotNo-len(an.toVertices)+1))
		}
		an.toVertices[knotNo] = data
		return nil
	}
}

func (an *Annex) GetFromVertex(knotNo int) (data interface{}) {
	if !an.knots.KnotExists(knotNo) || knotNo >= len(an.toVertices) {
		return nil
	} else {
		return an.toVertices[knotNo]
	}
}

func (an *Annex) AttachToSegment(segmentNo int, data interface{}) (err error) {
	if !an.knots.SegmentExists(segmentNo) {
		return fmt.Errorf("segment with no. %v doesn't exist", segmentNo)
	} else {
		if segmentNo >= len(an.toSegments) {
			an.toSegments = append(an.toSegments, make([]interface{}, segmentNo-len(an.toSegments)+1))
		}
		an.toSegments[segmentNo] = data
		return nil
	}
}

func (an *Annex) GetFromSegment(segmentNo int) (data interface{}) {
	if !an.knots.SegmentExists(segmentNo) || segmentNo >= len(an.toSegments) {
		return nil
	} else {
		return an.toSegments[segmentNo]
	}
}
