package ctr

// Constraint is an interface for constraints
type Constraint interface {
	Name() string
	Variables() []string
	ToXCSP() []string
}
