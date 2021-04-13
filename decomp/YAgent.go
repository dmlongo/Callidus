package decomp

// Performer can perform an action
type Performer interface {
	Perform(a Action) bool
}

// Action is a function to perform
type Action func() bool
