package ymca

type Actor interface {
}

// Message to be processed by Actors
type Message interface {
	Receiver() Actor
	Content() interface{}
}
