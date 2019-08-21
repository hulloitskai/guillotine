package guillotine

// Execute runs each Finalizer in reverse order.
//
// It returns a boolean (ok) that is true if all Finalizers ran without errors,
// and the slice of all the Finalizer errors that occurred during execution.
func (g *Guillotine) Execute() (ok bool, errs []error) {
	g.Trigger()
	errs = g.Wait()
	return len(errs) == 0, errs
}
