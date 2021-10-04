package cubic

import bendit "github.com/walpod/bend-it"

type ControlVertex interface {
	bendit.Vertex
	Entry() bendit.Vec
	Exit() bendit.Vec
	Dependent() bool // are entry and exit dependent on each other? TODO extend to scalingFactor (some direction)
	New(loc bendit.Vec, entry bendit.Vec, exit bendit.Vec) ControlVertex
	Translate(dv bendit.Vec) ControlVertex
	ControlToLoc(control bendit.Vec, isEntry bool) bendit.Vec
	LocToControl(loc bendit.Vec, isEntry bool) bendit.Vec
}

type BezierVertex struct {
	loc       bendit.Vec
	entry     bendit.Vec
	exit      bendit.Vec
	dependent bool
}

// one of entry or exit control can be nil, is handled as dependent control (on other side of the vertex)
func NewBezierVertex(loc, entry, exit bendit.Vec) *BezierVertex {
	dependent := false

	// handle dependent controls
	if entry == nil && exit != nil {
		entry = loc.InvertInPoint(exit)
		dependent = true
	} else if entry != nil && exit == nil {
		exit = loc.InvertInPoint(entry)
		dependent = true
	}

	return &BezierVertex{loc: loc, entry: entry, exit: exit, dependent: dependent}
}

func (vt *BezierVertex) Loc() bendit.Vec {
	return vt.loc
}

func (vt *BezierVertex) Entry() bendit.Vec {
	return vt.entry
}

func (vt *BezierVertex) Exit() bendit.Vec {
	return vt.exit
}

func (vt *BezierVertex) Dependent() bool {
	return vt.dependent
}

func (vt *BezierVertex) ControlToLoc(control bendit.Vec, isEntry bool) bendit.Vec {
	return control
}

func (vt *BezierVertex) LocToControl(loc bendit.Vec, isEntry bool) bendit.Vec {
	return loc
}

func (vt *BezierVertex) Translate(dv bendit.Vec) ControlVertex {
	var exit bendit.Vec
	if !vt.dependent {
		exit = vt.exit.Add(dv)
	}
	return NewBezierVertex(vt.loc.Add(dv), vt.entry.Add(dv), exit)
}

func (vt *BezierVertex) New(loc bendit.Vec, entry bendit.Vec, exit bendit.Vec) ControlVertex {
	return NewBezierVertex(loc, entry, exit)
}

type HermiteVertex struct {
	loc       bendit.Vec
	entry     bendit.Vec
	exit      bendit.Vec
	dependent bool
}

func NewHermiteVertex(loc, entry, exit bendit.Vec) *HermiteVertex {
	dependent := false

	// handle dependent tangents
	if entry == nil && exit != nil {
		entry = exit // TODO clone
		dependent = true
	} else if entry != nil && exit == nil {
		exit = entry // TODO clone
		dependent = true
	}

	return &HermiteVertex{loc, entry, exit, dependent}
}

func NewRawHermiteVertex(loc bendit.Vec) *HermiteVertex {
	return &HermiteVertex{loc: loc, entry: nil, exit: nil, dependent: false}
}

func (vt *HermiteVertex) Loc() bendit.Vec {
	return vt.loc
}

func (vt *HermiteVertex) Entry() bendit.Vec {
	return vt.entry
}

func (vt *HermiteVertex) Exit() bendit.Vec {
	return vt.exit
}

func (vt *HermiteVertex) Dependent() bool {
	return vt.dependent
}

func (vt *HermiteVertex) ControlToLoc(control bendit.Vec, isEntry bool) bendit.Vec {
	if isEntry {
		return vt.loc.Sub(control)
	} else {
		return vt.loc.Add(control)
	}
}

func (vt *HermiteVertex) LocToControl(loc bendit.Vec, isEntry bool) bendit.Vec {
	if isEntry {
		return vt.loc.Sub(loc)
	} else {
		return loc.Sub(vt.loc)
	}
}

func (vt *HermiteVertex) Translate(dv bendit.Vec) ControlVertex {
	return NewHermiteVertex(vt.loc.Add(dv), vt.entry, vt.exit)
}

func (vt *HermiteVertex) New(loc bendit.Vec, entry bendit.Vec, exit bendit.Vec) ControlVertex {
	return NewHermiteVertex(loc, entry, exit)
}

func NewControlVertexWithControl(vt ControlVertex, control bendit.Vec, isEntry bool) ControlVertex {
	var entry, exit bendit.Vec
	if isEntry {
		entry = control
		exit = vt.Exit()
		if vt.Dependent() {
			exit = nil
		}
	} else {
		entry = vt.Entry()
		exit = control
		if vt.Dependent() {
			entry = nil
		}
	}
	return vt.New(vt.Loc(), entry, exit)
}

func NewControlVertexWithControlLoc(vt ControlVertex, loc bendit.Vec, isEntry bool) ControlVertex {
	return NewControlVertexWithControl(vt, vt.LocToControl(loc, isEntry), isEntry)
}

func Control(vt ControlVertex, isEntry bool) bendit.Vec {
	if isEntry {
		return vt.Entry()
	} else {
		return vt.Exit()
	}
}

func ControlLoc(vt ControlVertex, isEntry bool) bendit.Vec {
	if isEntry {
		return vt.ControlToLoc(vt.Entry(), true)
	} else {
		return vt.ControlToLoc(vt.Exit(), false)
	}
}
