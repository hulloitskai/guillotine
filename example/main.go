package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.stevenxie.me/guillotine"
)

func main() {
	// Create guillotine.
	guillo := guillotine.New()

	// Execute the guillotine before main finishes.
	defer func() {
		fmt.Println("Executing guillotine...")
		if errs := guillo.Execute(); len(errs) > 0 {
			for _, err := range errs {
				fmt.Fprintf(os.Stderr, "A finalizer failed: %v", err)
			}
			os.Exit(1)
		}
	}()

	// Initialize resources, like files or databases.
	file, err := os.Open("resource.txt")
	if err != nil {
		panic(err)
	}
	guillo.AddCloser(file, guillotine.WithPrefix("closing file"))

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
	go func() {
		fmt.Printf("Listening on port %d...\n", port)
		if err := srv.ListenAndServe(); err != nil {
			guillo.Trigger()
		}
	}()
	guillo.AddFinalizer(
		func() error { return srv.Shutdown(context.Background()) },
		guillotine.WithPrefix("shutting down server"),
	)

	// Listen for an termination signal.
	guillo.TriggerOnTerminate()
}
