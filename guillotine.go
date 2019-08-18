package guillotine

import (
	"context"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/zero"
)

// New creates a new Guillotine.
func New(opts ...Option) *Guillotine {
	cfg := Config{
		Logger: zero.Logger(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Guillotine{
		log: cfg.Logger,

		ctx:    ctx,
		cancel: cancel,
		done:   make(chan zero.Struct, 1),
	}
}

type (
	// A Guillotine is used to shut down a system consisting of multiple
	// components.
	Guillotine struct {
		finalizers []Finalizer
		errors     []error
		log        logrus.FieldLogger

		ctx    context.Context
		cancel context.CancelFunc

		execOnce sync.Once
		waitOnce sync.Once
		done     chan zero.Struct
	}

	// A Finalizer is a function that returns an error.
	Finalizer func() error

	// A Callback is called when a Guillotine runs a Finalizer.
	//
	// It is used as a way to perform other actions after a Finalizer is run,
	// and to modify the error returned by the Finalizer.
	Callback func(error) error
)

// AddFinalizer adds a Finalizer to the Guillotine.
func (g *Guillotine) AddFinalizer(f Finalizer, cb ...Callback) {
	g.finalizers = append(g.finalizers, func() error {
		return runCallbacks(f(), cb)
	})
}

// AddFunc adds a regular function to the Guillotine.
func (g *Guillotine) AddFunc(f func(), cb ...Callback) {
	g.finalizers = append(g.finalizers, func() error {
		f()
		return runCallbacks(nil, cb)
	})
}

// AddCloser adds an io.Closer to the Guillotine.
func (g *Guillotine) AddCloser(closer io.Closer, cb ...Callback) {
	g.finalizers = append(g.finalizers, func() error {
		return runCallbacks(closer.Close(), cb)
	})
}

// Wait blocks until the Guillotine completes its asynchronous execution
// process, and returns the resulting errors from its finalizers.
func (g *Guillotine) Wait() []error {
	g.waitOnce.Do(func() { <-g.done })
	return g.errors
}

func runCallbacks(err error, cbs []Callback) error {
	for _, cb := range cbs {
		err = cb(err)
	}
	return err
}
