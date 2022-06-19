package cubic

import "github.com/walpod/bendigo"

type EnexVertex struct {
	loc       bendigo.Vec
	entry     bendigo.Vec
	exit      bendigo.Vec
	relative  bool // are entry and exit controls relative to loc or absolut
	dependent bool // TODO dependencyFactor ? entry-or-exit dependent?
}

func dependantControl(loc bendigo.Vec, control bendigo.Vec, relative bool) bendigo.Vec {
	if relative {
		return control // TODO clone ? at least in the future when vec can be modified (not only replaced)
	} else {
		return loc.InvertInPoint(control)
	}
}

// NewEnexVertex creates entry-exit vertex. if one of entry or exit control is nil, they are handled as dependent controls
func NewEnexVertex(loc, entry, exit bendigo.Vec, relative bool) *EnexVertex {
	dependent := false

	// handle dependent controls
	if entry == nil && exit != nil {
		entry = dependantControl(loc, exit, relative)
		dependent = true
	} else if entry != nil && exit == nil {
		exit = dependantControl(loc, entry, relative)
		dependent = true
	}

	return &EnexVertex{loc: loc, entry: entry, exit: exit, relative: relative, dependent: dependent}
}

func NewEnexVertexDep(loc, entry, exit bendigo.Vec, relative bool, dependent bool) *EnexVertex {
	return &EnexVertex{loc: loc, entry: entry, exit: exit, relative: relative, dependent: dependent}
}

func (ev *EnexVertex) Loc() bendigo.Vec {
	return ev.loc
}

func (ev *EnexVertex) Entry() bendigo.Vec {
	return ev.entry
}

func (ev *EnexVertex) SetEntry(entry bendigo.Vec) {
	ev.entry = entry
	if ev.dependent {
		ev.exit = dependantControl(ev.loc, entry, ev.relative)
	}
}

func (ev *EnexVertex) EntryAsAbsolute() bendigo.Vec {
	if ev.relative {
		return ev.loc.Sub(ev.entry)
	} else {
		return ev.entry
	}
}

func (ev *EnexVertex) Exit() bendigo.Vec {
	return ev.exit
}

func (ev *EnexVertex) SetExit(exit bendigo.Vec) {
	ev.exit = exit
	if ev.dependent {
		ev.entry = dependantControl(ev.loc, exit, ev.relative)
	}
}

func (ev *EnexVertex) ExitAsAbsolute() bendigo.Vec {
	if ev.relative {
		return ev.loc.Add(ev.exit)
	} else {
		return ev.exit
	}
}

func (ev *EnexVertex) Relative() bool {
	return ev.relative
}

func (ev *EnexVertex) Absolute() bool {
	return !ev.relative
}

func (ev *EnexVertex) Dependent() bool {
	return ev.dependent
}

func (ev *EnexVertex) SetDependent(dependent bool) {
	ev.dependent = dependent
}

func (ev *EnexVertex) ToggleDependent(isEntry bool) {
	ev.dependent = !ev.dependent

	// if changed to dependent then recreate the other control
	if ev.dependent {
		ev.SetControl(dependantControl(ev.loc, ev.Control(isEntry), ev.relative), !isEntry)
	}
}

// Control returns requested entry or exit control
func (ev *EnexVertex) Control(isEntry bool) bendigo.Vec {
	if isEntry {
		return ev.entry
	} else {
		return ev.exit
	}
}

func (ev *EnexVertex) SetControl(control bendigo.Vec, isEntry bool) {
	if isEntry {
		ev.SetEntry(control)
	} else {
		ev.SetExit(control)
	}
}

// ControlAsAbsolute returns requested entry or exit control, converted to absolute if given vertex has relative controls
func (ev *EnexVertex) ControlAsAbsolute(isEntry bool) bendigo.Vec {
	if isEntry {
		return ev.EntryAsAbsolute()
	} else {
		return ev.ExitAsAbsolute()
	}
}

// Shift shifts (translates) the vertex in direction given by vector dv
func (ev *EnexVertex) Shift(dv bendigo.Vec) {
	ev.loc = ev.loc.Add(dv)
	if ev.Absolute() {
		ev.entry = ev.entry.Add(dv)
		ev.exit = ev.exit.Add(dv)
	}
}

// Clone returns a shallow copy of the vertex
func (ev *EnexVertex) Clone() *EnexVertex {
	return NewEnexVertexDep(ev.loc, ev.entry, ev.exit, ev.relative, ev.dependent)
}

// WithShift creates a new EnexVertex, shifted (translated) in direction given by vector dv
func (ev *EnexVertex) WithShift(dv bendigo.Vec) *EnexVertex {
	nev := ev.Clone()
	nev.Shift(dv)
	return nev
}

// WithEntry creates a new EnexVertex with a modified entry control
func (ev *EnexVertex) WithEntry(control bendigo.Vec) *EnexVertex {
	nev := ev.Clone()
	nev.SetEntry(control)
	return nev
}

// WithExit creates a new EnexVertex with a modified entry control
func (ev *EnexVertex) WithExit(control bendigo.Vec) *EnexVertex {
	nev := ev.Clone()
	nev.SetExit(control)
	return nev
}

// WithControl creates a new EnexVertex with a modified control
func (ev *EnexVertex) WithControl(control bendigo.Vec, isEntry bool) *EnexVertex {
	nev := ev.Clone()
	nev.SetControl(control, isEntry)
	return nev
}
