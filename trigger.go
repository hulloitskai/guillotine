package guillotine

import (
	"os"
	"os/signal"
	"syscall"

	"go.stevenxie.me/gopkg/zero"
)

// Execute runs each Finalizer in reverse order, and returns the resulting
// errors.
func (g *Guillotine) Execute() []error {
	g.Trigger()
	return g.Wait()
}

// Trigger triggers an asynchronous execution process which will run each
// Finalizer in reverse order.
//
// Use g.Wait to wait for the execution process to complete.
func (g *Guillotine) Trigger() { go g.execOnce.Do(g.execute) }

// TriggerOnTerminate waits for a termination signal (syscall.SIGTERM,
// syscall.SIGINT) before triggering an exeuction (see g.Terminate).
func (g *Guillotine) TriggerOnTerminate() {
	g.TriggerOnSignal(os.Interrupt, syscall.SIGTERM)
}

// TriggerOnSignal waits for a signal before triggering an execution
// (see g.Terminate).
func (g *Guillotine) TriggerOnSignal(sig ...os.Signal) {
	ch := make(chan os.Signal)
	go signal.Notify(ch, sig...)

	// Wait for signal (or cancellation).
	select {
	case <-g.ctx.Done(): // abort
		return
	case sig := <-ch: // continue
		g.log.
			WithField("signal", sig).
			Info("Received execution signal.")
	}

	// Execute
	g.Trigger()
}

func (g *Guillotine) execute() {
	// Cancel listener context.
	g.cancel()

	// Run finalizers.
	g.log.Info("Running finalizers...")
	for i := len(g.finalizers) - 1; i >= 0; i-- {
		if finErr := g.finalizers[i](); finErr != nil {
			g.log.WithError(finErr).Error("A finalizer failed.")
			g.errors = append(g.errors, finErr)
		}
	}

	// Signal execution completion.
	g.done <- zero.Empty() // signal completion
}
