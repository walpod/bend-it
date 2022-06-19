package cubic

import "github.com/walpod/bendigo"

type EnexVertex struct {
	loc        bendigo.Vec
	entry      bendigo.Vec
	exit       bendigo.Vec
	relative   bool // are entry and exit controls relative to loc or absolute
	leading    bool // is one of 'entry' or 'exit' the leader and the other the follower
	entryLeads bool // is 'entry' leading and 'exit' leading or vice versa, irrelevant if !leading
	// TODO leadingFactor
}

func follower(leader bendigo.Vec, loc bendigo.Vec, relative bool) bendigo.Vec {
	if relative {
		return leader // TODO clone ? at least in the future when vec can be modified (not only replaced)
	} else {
		return loc.InvertInPoint(leader)
	}
}

// NewEnexVertex creates entry-exit vertex. if one of entry or exit control is nil then they are handled as leading controls
func NewEnexVertex(loc, entry, exit bendigo.Vec, relative bool) *EnexVertex {
	leading := false
	entryLeads := false

	// handle leading controls
	if entry != nil && exit == nil {
		leading = true
		entryLeads = true
		exit = follower(entry, loc, relative)
	} else if entry == nil && exit != nil {
		leading = true
		entryLeads = false
		entry = follower(exit, loc, relative)
	}

	return &EnexVertex{loc: loc, entry: entry, exit: exit, relative: relative, leading: leading, entryLeads: entryLeads}
}

func NewEnexVertexDep(loc, entry, exit bendigo.Vec, relative bool, leading bool, entryLeads bool) *EnexVertex {
	return &EnexVertex{loc: loc, entry: entry, exit: exit, relative: relative, leading: leading, entryLeads: entryLeads}
}

func (ev *EnexVertex) Loc() bendigo.Vec {
	return ev.loc
}

func (ev *EnexVertex) Entry() bendigo.Vec {
	return ev.entry
}

func (ev *EnexVertex) SetEntry(entry bendigo.Vec) {
	ev.entry = entry
	if ev.leading {
		ev.entryLeads = true
		ev.exit = follower(entry, ev.loc, ev.relative)
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
	if ev.leading {
		ev.entryLeads = false
		ev.entry = follower(exit, ev.loc, ev.relative)
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

func (ev *EnexVertex) Leading() bool {
	return ev.leading
}

// RecreateFollower recalculates the follower if vertex is set to leading
func (ev *EnexVertex) RecreateFollower() {
	if ev.leading {
		leader := ev.Control(ev.entryLeads)
		ev.SetControl(follower(leader, ev.loc, ev.relative), !ev.entryLeads)
	}
}

func (ev *EnexVertex) SetLeading(leading bool, entryLeads bool) {
	ev.leading = leading
	ev.entryLeads = entryLeads
	ev.RecreateFollower()
}

func (ev *EnexVertex) ToggleLeading(entryLeads bool) {
	ev.leading = !ev.leading
	ev.entryLeads = entryLeads
	ev.RecreateFollower()
}

func (ev *EnexVertex) EntryLeads() bool {
	return ev.entryLeads
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
	return NewEnexVertexDep(ev.loc, ev.entry, ev.exit, ev.relative, ev.leading, ev.entryLeads)
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
