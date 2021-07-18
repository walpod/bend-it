package cubic

type Controller interface {
	ControlX() float64
	ControlY() float64
	IsForExchange() bool
	IsCalculated() bool
}

type Control struct {
	x, y float64
}

func NewControl(x float64, y float64) *Control {
	return &Control{x: x, y: y}
}

func (c *Control) ControlX() float64 {
	return c.x
}

func (c *Control) ControlY() float64 {
	return c.y
}

func (c *Control) IsForExchange() bool {
	return false
}

func (c *Control) IsCalculated() bool {
	return false
}

type Reflective struct{}

func NewReflective() *Reflective {
	return &Reflective{}
}

func (r *Reflective) ControlX() float64 {
	panic("cannot be used directly. Reflective can only be used as parameter for spline constructors and construction API")
}

func (r *Reflective) ControlY() float64 {
	panic("cannot be used directly. Reflective can only be used as parameter for spline constructors and construction API")
}

func (r *Reflective) IsForExchange() bool {
	return true
}

func (r *Reflective) IsCalculated() bool {
	return true
}

/*type OriginReflection struct {
	baseControlX, baseControlY float64
}

func (o OriginReflection) ControlX() float64 {
	return -o.baseControlX
}

func (o OriginReflection) ControlY() float64 {
	return -o.baseControlY
}*/

type PointReflection struct {
	pointX, pointY             float64
	baseControlX, baseControlY float64
}

func NewPointReflection(pointX float64, pointY float64, baseControlX float64, baseControlY float64) *PointReflection {
	return &PointReflection{pointX: pointX, pointY: pointY, baseControlX: baseControlX, baseControlY: baseControlY}
}

func (p *PointReflection) ControlX() float64 {
	return p.pointX + (p.pointX - p.baseControlX)
}

func (p *PointReflection) ControlY() float64 {
	return p.pointY + (p.pointY - p.baseControlY)
}

func (p *PointReflection) IsForExchange() bool {
	return false
}

func (p *PointReflection) IsCalculated() bool {
	return true
}
