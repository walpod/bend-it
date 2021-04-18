package bendit

type Fn2d func(t float64) (x, y float64)

type Spline2d interface {
	Domain() (fr, to float64)
	At(t float64) (x, y float64)
	Fn() Fn2d
}
