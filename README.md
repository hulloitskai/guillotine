# guillotine

_Terminating systems with multiple components, in style._

[![Tag][tag-img]][tag]
[![Drone][drone-img]][drone]
[![Go Report Card][grp-img]][grp]
[![GoDoc][godoc-img]][godoc]

## Usage

`guillotine` is a package for shutting down complex, multi-component systems.

It defines a `Finalizer`, which is a function that should be run during program
termination.

```go
type Finalizer func() error
```

`Finalizers` should be added to the `Guillotine` in your `main` function, after
resources are initialized. After all `Finalizers` are added, you should start
long-running processes like HTTP servers on separate goroutines; call
`guillotine.Trigger` after these processes exit, in case they exit early due
to a startup failure.

```go
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/guillotine"
)

func main() {
	// Create a guillotine, configure it to trigger when receiving a termination
	// signal from the OS.
	guillo := guillotine.New()
	guillo.TriggerOnTerminate()

	// Execute the guillotine before main finishes.
	defer func() {
		if errs := guillo.Execute(); len(errs) > 0 {
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
	guillo.AddFinalizer(
		func() error { return srv.Shutdown(context.Background()) },
		guillotine.WithPrefix("shutting down server"),
	)

	// Blocks thread while server runs; stops either when the Guillotine
	// shuts down the server, or the server fails to start up.
	fmt.Printf("Listening on port %d...\n", port)
	if err = srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "Error while starting server: %v\n", err)
			guillo.Execute()
			os.Exit(2)
		}
	}
}
```

See [the full example](./example/main.go) for more details.

[tag]: https://github.com/stevenxie/gopkg/releases
[tag-img]: https://img.shields.io/github/tag/stevenxie/gopkg.svg
[drone]: https://ci.stevenxie.me/stevenxie/gopkg
[drone-img]: https://ci.stevenxie.me/api/badges/stevenxie/gopkg/status.svg
[grp]: https://goreportcard.com/report/go.stevenxie.me/gopkg
[grp-img]: https://goreportcard.com/badge/go.stevenxie.me/gopkg
[godoc]: https://godoc.org/go.stevenxie.me/gopkg
[godoc-img]: https://godoc.org/go.stevenxie.me/gopkg?status.svg
