package cubic

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

func (c *Control) Move(dx float64, dy float64) *Control {
	return NewControl(c.x+dx, c.y+dy)
}

// NewDependentControl creates a symmetric (reflective) control to given point x,y and base-control
func NewDependentControl(x, y float64, base *Control) *Control {
	if base == nil {
		return nil
	} else {
		return NewControl(2*x-base.x, 2*y-base.y)
	}
}
