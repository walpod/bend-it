package cubic

import "github.com/walpod/bendigo"

type ControlVertex interface {
	bendigo.Vertex
	Entry() bendigo.Vec
	Exit() bendigo.Vec
	Dependent() bool // are entry and exit dependent on each other? TODO extend to scalingFactor (same direction)
	New(loc bendigo.Vec, entry bendigo.Vec, exit bendigo.Vec) ControlVertex
	Translate(dv bendigo.Vec) ControlVertex // TODO => Shift
	ControlToLoc(control bendigo.Vec, isEntry bool) bendigo.Vec
	LocToControl(loc bendigo.Vec, isEntry bool) bendigo.Vec
}

type BezierVertex struct {
	loc       bendigo.Vec
	entry     bendigo.Vec
	exit      bendigo.Vec
	dependent bool
}

// one of entry or exit control can be nil, is handled as dependent control (on other side of the vertex)
func NewBezierVertex(loc, entry, exit bendigo.Vec) *BezierVertex {
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

func (vt *BezierVertex) Loc() bendigo.Vec {
	return vt.loc
}

func (vt *BezierVertex) Entry() bendigo.Vec {
	return vt.entry
}

func (vt *BezierVertex) Exit() bendigo.Vec {
	return vt.exit
}

func (vt *BezierVertex) Dependent() bool {
	return vt.dependent
}

func (vt *BezierVertex) ControlToLoc(control bendigo.Vec, isEntry bool) bendigo.Vec {
	return control
}

func (vt *BezierVertex) LocToControl(loc bendigo.Vec, isEntry bool) bendigo.Vec {
	return loc
}

func (vt *BezierVertex) Translate(dv bendigo.Vec) ControlVertex {
	var exit bendigo.Vec
	if !vt.dependent {
		exit = vt.exit.Add(dv)
	}
	return NewBezierVertex(vt.loc.Add(dv), vt.entry.Add(dv), exit)
}

func (vt *BezierVertex) New(loc bendigo.Vec, entry bendigo.Vec, exit bendigo.Vec) ControlVertex {
	return NewBezierVertex(loc, entry, exit)
}

type HermiteVertex struct {
	loc       bendigo.Vec
	entry     bendigo.Vec
	exit      bendigo.Vec
	dependent bool
}

func NewHermiteVertex(loc, entry, exit bendigo.Vec) *HermiteVertex {
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

func NewRawHermiteVertex(loc bendigo.Vec) *HermiteVertex {
	return &HermiteVertex{loc: loc, entry: nil, exit: nil, dependent: false}
}

func (vt *HermiteVertex) Loc() bendigo.Vec {
	return vt.loc
}

func (vt *HermiteVertex) Entry() bendigo.Vec {
	return vt.entry
}

func (vt *HermiteVertex) Exit() bendigo.Vec {
	return vt.exit
}

func (vt *HermiteVertex) Dependent() bool {
	return vt.dependent
}

func (vt *HermiteVertex) ControlToLoc(control bendigo.Vec, isEntry bool) bendigo.Vec {
	if isEntry {
		return vt.loc.Sub(control)
	} else {
		return vt.loc.Add(control)
	}
}

func (vt *HermiteVertex) LocToControl(loc bendigo.Vec, isEntry bool) bendigo.Vec {
	if isEntry {
		return vt.loc.Sub(loc)
	} else {
		return loc.Sub(vt.loc)
	}
}

func (vt *HermiteVertex) Translate(dv bendigo.Vec) ControlVertex {
	return NewHermiteVertex(vt.loc.Add(dv), vt.entry, vt.exit)
}

func (vt *HermiteVertex) New(loc bendigo.Vec, entry bendigo.Vec, exit bendigo.Vec) ControlVertex {
	return NewHermiteVertex(loc, entry, exit)
}

func NewControlVertexWithControl(vt ControlVertex, control bendigo.Vec, isEntry bool) ControlVertex {
	var entry, exit bendigo.Vec
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

func NewControlVertexWithControlLoc(vt ControlVertex, loc bendigo.Vec, isEntry bool) ControlVertex {
	return NewControlVertexWithControl(vt, vt.LocToControl(loc, isEntry), isEntry)
}

func Control(vt ControlVertex, isEntry bool) bendigo.Vec {
	if isEntry {
		return vt.Entry()
	} else {
		return vt.Exit()
	}
}

func ControlLoc(vt ControlVertex, isEntry bool) bendigo.Vec {
	if isEntry {
		return vt.ControlToLoc(vt.Entry(), true)
	} else {
		return vt.ControlToLoc(vt.Exit(), false)
	}
}
