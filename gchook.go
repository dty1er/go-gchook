package gchook

import (
	"runtime"
	"sync"
)

// gchookManager manages some stuffs about gchook core feature
// It contains mutex and registered hook functios, and some channels
// Mutex is actually used when adding hook functions to hooks
// This should not be exported because clients doesn't need to handle it, and
// they don't need to create its instance
// I expect clients to access the manager only through Register() and Cancel() functions
type gchookManager struct {
	sync.Mutex
	hooks    []func()
	gc       chan struct{}
	canceled chan struct{}
}

// garbage is just a meaningless struct
// This is expected to be collected by garbage collecto
// If garbage struct is empty, it won't be collected by gc so
// need to contain useless byte (or so on)
type garbage struct{ _ byte }

// manager is the singleton manager, it cannot be accessed by clients directly
// gc is the channel to receive notification when gc occurred
// canceled is the channel to receive notification when client canceled
var manager = &gchookManager{
	gc:       make(chan struct{}, 1),
	canceled: make(chan struct{}, 1),
}

// Register registers given functions to go-gchook internal manager
// When registered, functions are executed whenever garbage collection occurred.
// Once registered, you cannot unregister it.
func Register(f ...func()) {
	manager.register(f)
}

// Cancel sends cancel to internal manager
// After calling Cancel(), all registered functions will not be executed anymore
func Cancel() {
	manager.cancel()
}

// When initialize, try to set garbage and finalizer using `runtime.SetFinalizer`
// runtime.SetFinalizer tries to pass 1st arg to 2nd arg (it means that 2nd arg must be
// function, and type must be the same).
// `&garbage{}` is not named, so it will be a target of garbage collector.
func init() {
	runtime.SetFinalizer(&garbage{}, gcHook)
	manager.startWorker()
}

// gcHook is the function to be passed to runtime.SetFinalizer.
// In this function, send empty struct to manager.gc channel, and worker receives it.
// And, also given garbage also be set to finalizer again.
func gcHook(g *garbage) {
	defer runtime.SetFinalizer(g, gcHook)
	manager.gc <- struct{}{}
}

func (m *gchookManager) register(hooks []func()) {
	m.Lock()
	m.hooks = append(m.hooks, hooks...)
	m.Unlock()
}

func (m *gchookManager) cancel() {
	m.canceled <- struct{}{}
}

// This worker actually acts very simply.
// It loops with select statement, and when receives data in gc channel,
// tries to run every registered hook functions concurrently
// If it receives cancel, this worker will be terminated.
func (m *gchookManager) startWorker() {
	go func() {
		for {
			select {
			case <-m.canceled:
				return

			case <-m.gc:
				wg := &sync.WaitGroup{}
				wg.Add(len(m.hooks))
				for _, hook := range m.hooks {
					go func(hook func()) {
						hook()
						wg.Done()
					}(hook)
				}
			}
		}
	}()
}
