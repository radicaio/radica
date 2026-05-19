// Command radicad is the standalone Radica binary.
//
// MVP: starts the replica loop and serves HTTP on the configured address
// with /healthz and a /kv/{key} endpoint so curl can drive Get/Set/Delete.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/radicaio/radica"
	radicahttp "github.com/radicaio/radica/internal/http"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	r := radica.Open()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Replica loop. Per the style guide, library code never spawns
	// goroutines; the binary owns this one.
	loopDone := make(chan error, 1)
	go func() { loopDone <- r.Run(ctx) }()

	srv := &http.Server{
		Addr:              *addr,
		Handler:           radicahttp.NewServer(r.Transport()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	httpDone := make(chan error, 1)
	go func() {
		log.Printf("radicad: listening on %s", *addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			httpDone <- err
			return
		}
		httpDone <- nil
	}()

	select {
	case <-ctx.Done():
		log.Print("radicad: shutting down")
	case err := <-httpDone:
		if err != nil {
			log.Printf("radicad: http server error: %v", err)
		}
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)

	<-loopDone
	log.Print("radicad: stopped")
}
