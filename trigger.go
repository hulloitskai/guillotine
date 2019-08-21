package guillotine

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.stevenxie.me/gopkg/zero"
)

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
	sigs := make(chan os.Signal)
	go signal.Notify(sigs, sig...)

	// Wait for signal (or cancellation), then trigger execution.
	go func(ctx context.Context, sigs <-chan os.Signal) {
		select {
		case <-g.ctx.Done(): // abort
			return
		case sig := <-sigs: // continue
			g.log.
				WithField("signal", sig).
				Info("Received execution signal.")
		}

		// Execute
		g.Trigger()
	}(g.ctx, sigs)
}

func (g *Guillotine) execute() {
	// Cancel listener context; stop all goroutines that are waiting to trigger
	// the Guillotine.
	g.cancel()

	// Run finalizers in reverse order.
	g.log.Info("Running finalizers...")
	for i := len(g.finalizers) - 1; i >= 0; i-- {
		if finErr := g.finalizers[i](); finErr != nil {
			g.log.WithError(finErr).Error("A finalizer failed.")
			g.errors = append(g.errors, finErr)
		}
	}

	// Signal execution completion.
	g.done <- zero.Empty()
}
