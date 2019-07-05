package gchook

import (
	"fmt"
	"runtime"
	"sync"
)

type gchookManager struct {
	sync.Mutex
	hooks []func()
}

var hooks []func()

var gc = make(chan struct{}, 1)
var canceled = make(chan struct{}, 1)

func init() {
	// var defaultManager = &gchookManager{
	// 	// gc:       make(chan struct{}, 1),
	// 	// canceled: make(chan struct{}, 1),
	// }
	runtime.SetFinalizer(&gchookManager{}, gcHappened)
	startWorker()
}

func gcHappened(m *gchookManager) {
	fmt.Printf("m = %+v\n", m)
	gc <- struct{}{}
	runtime.SetFinalizer(m, gcHappened)
}

// Register ...
func Register(f ...func()) {
	// defaultManager.register(f)
	hooks = append(hooks, f...)
}

// // Cancel ...
// func Cancel() {
// 	defaultManager.cancel()
// }

func (m *gchookManager) register(hooks []func()) {
	m.Lock()
	m.hooks = append(m.hooks, hooks...)
	m.Unlock()
}

func (m *gchookManager) cancel() {
	canceled <- struct{}{}
}

func startWorker() {
	fmt.Println("startWorker")
	fmt.Printf("hooks = %+v\n", hooks)
	go func() {
		for {
			fmt.Printf("hooks = %+v\n", hooks)
			select {
			case <-canceled:
				fmt.Println("canceled")
				return

			case <-gc:
				fmt.Println("gc!!!!!!!!")
				// m.Lock()
				wg := &sync.WaitGroup{}
				wg.Add(len(hooks))
				for _, hook := range hooks {
					go func(hook func()) {
						hook()
						wg.Done()
					}(hook)
				}
				// m.Unlock()
			}
		}
	}()
}
