package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/gopkg/cmdutil"
	"go.stevenxie.me/guillotine"
)

func main() {
	// Create a guillotine, configure it to trigger when receiving a termination
	// signal from the OS.
	guillo := guillotine.New()
	guillo.TriggerOnTerminate()

	// Execute the guillotine before main finishes.
	defer func() {
		if ok, errs := guillo.Execute(); !ok {
			for _, err := range errs {
				fmt.Fprintf(os.Stderr, "A finalizer failed: %v\n", err)
			}
			os.Exit(1)
		}
	}()

	// Initialize resources, like files or databases.
	file, err := os.Open("resource.txt")
	if err != nil {
		panic(err)
	}
	guillo.AddCloser(
		file,
		guillotine.WithPrefix("closing file"),
		guillotine.WithFunc(func() { fmt.Println("Closing file...") }),
	)

	// Start long-running processes, like servers.
	//
	// We add the server itself as a finalizer so it can be shut down in response
	// to some other shutdown signal, like an interrupt / termination signal.
	const port = 8080
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if _, err := io.Copy(w, file); err != nil {
				panic(err)
			}
		}),
	}
	guillo.AddFinalizer(
		func() error { return srv.Shutdown(context.Background()) },
		guillotine.WithPrefix("shutting down server"),
		guillotine.WithFunc(func() { fmt.Println("Shutting down server...") }),
	)

	// Block thread while server runs; stops either when the Guillotine
	// shuts down the server, or the server fails to start up.
	fmt.Printf("Listening on port %d...\n", port)
	if err = srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			cmdutil.Errf("Error while starting server: %v\n", err)
			guillo.Execute()
			os.Exit(2)
		}
	}
}
