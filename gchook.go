package gchook

type gchookManager struct {
	done chan struct{}
}

var defaultManager = &gchookManager{
	done: make(chan struct{}, 1),
}

// Register ...
func Register(func()) {

}
