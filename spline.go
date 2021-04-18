package bendit

type Fn2d func(t float64) (x, y float64)

// range of parameter t the spline is defined for
type SplineDomain struct {
	From, To float64
}

type Spline2d interface {
	Domain() SplineDomain
	At(t float64) (x, y float64)
	Fn() Fn2d
}
