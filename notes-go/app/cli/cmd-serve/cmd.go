package cmdserve

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"main.go/app/api"
	"main.go/app/shared/services/postgrey"
)

const (
	shutdownTimeout = 10 * time.Second
)

// Serve runs the API until SIGINT or SIGTERM; requests in progress finish first.
func Serve(cobra *cobra.Command, args []string) error {
	addr, err := cobra.Flags().GetString("addr")
	if err != nil {
		return err
	}
	origin, err := cobra.Flags().GetString("origin")
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(cobra.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := postgrey.GetDb()
	server := &http.Server{
		Addr:              addr,
		Handler:           api.New(db, origin),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Minute,
	}

	failed := make(chan error, 1)
	go func() { failed <- server.ListenAndServe() }()
	log.Printf("listening on %s", addr)

	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
	}

	// Not ctx: it is already canceled, and requests in progress need time to finish.
	shutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdown); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	log.Print("stopped")
	return nil
}
