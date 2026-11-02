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
	"main.go/app/api/services/guard"
	"main.go/app/shared/services/postgrey"
	"main.go/app/shared/services/process/purgeitems"
	"main.go/config"
)

const (
	purgeInterval   = time.Hour
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
	if err := guard.CheckOrigin(origin); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(cobra.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := postgrey.GetDb()
	go every(ctx, purgeInterval, func() {
		purge := purgeitems.Process(ctx, db, config.Current.Trash.RetentionDays)
		purge.Run()
		if purge.HasErr() && ctx.Err() == nil {
			log.Printf("purge: %v", purge.GetErr())
		}
	})

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

// every runs task at each interval until ctx ends: the root command only purges at startup, a server runs for weeks.
func every(ctx context.Context, interval time.Duration, task func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			task()
		}
	}
}
