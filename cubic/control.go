package cubic

// TODO rename to or use Vec
type Control struct {
	x, y float64
}

func NewControl(x float64, y float64) *Control {
	return &Control{x: x, y: y}
}

func (c *Control) X() float64 {
	return c.x
}

func (c *Control) Y() float64 {
	return c.y
}

func (c *Control) Translate(dx float64, dy float64) *Control {
	return NewControl(c.x+dx, c.y+dy)
}

func (c *Control) Scale(s float64) *Control {
	return NewControl(c.x*s, c.y*s)
}

// NewMirroredControl creates a symmetric, mirrored control of base-control to given point x,y
func NewMirroredControl(x, y float64, base *Control) *Control {
	if base == nil {
		return nil
	} else {
		return NewControl(2*x-base.x, 2*y-base.y)
	}
}
