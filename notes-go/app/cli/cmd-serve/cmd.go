package cmdserve

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cyril-fl/notes/notes-go/app/api"
	"github.com/cyril-fl/notes/notes-go/app/api/services/guard"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/postgrey"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/purgefiles"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/purgeitems"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/process/purgesessions"
	"github.com/cyril-fl/notes/notes-go/app/shared/services/silo"
	"github.com/spf13/cobra"
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
	web, err := cobra.Flags().GetString("web")
	if err != nil {
		return err
	}
	published, err := cobra.Flags().GetString("published")
	if err != nil {
		return err
	}
	if published == "" {
		published = addr
	}

	ctx, stop := signal.NotifyContext(cobra.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db := postgrey.GetDb()
	store, err := silo.New(ctx)
	if err != nil {
		return err
	}
	// The files after the items: a note purged now frees its images in the same round.
	go every(ctx, purgeInterval, func() {
		purge := purgeitems.Process(ctx, db)
		purge.Run()
		if purge.HasErr() && ctx.Err() == nil {
			log.Printf("purge: %v", purge.GetErr())
		}
		files := purgefiles.Process(ctx, db, store)
		files.Run()
		if files.HasErr() && ctx.Err() == nil {
			log.Printf("purge files: %v", files.GetErr())
		}
		sessions := purgesessions.Process(ctx, db)
		sessions.Run()
		if sessions.HasErr() && ctx.Err() == nil {
			log.Printf("purge sessions: %v", sessions.GetErr())
		}
	})

	handler := api.New(db, store, origin)
	if web != "" {
		handler = withWeb(web, handler)
	}
	if onLoopback(published) {
		handler = withProxy(handler)
	}
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
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

// onLoopback tells if only this machine reaches addr: then the X-Forwarded-For of a request can only come from a proxy
// running here (tailscale serve), never from a client.
func onLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// withProxy gives the requests the client's address: behind a proxy, RemoteAddr is the proxy's for every device, and
// the login delay would block them all. The last X-Forwarded-For entry is the one the proxy wrote, the client cannot
// forge it.
func withProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		forwarded := r.Header.Values("X-Forwarded-For")
		if len(forwarded) > 0 {
			entries := strings.Split(forwarded[len(forwarded)-1], ",")
			if address := strings.TrimSpace(entries[len(entries)-1]); address != "" {
				r.RemoteAddr = net.JoinHostPort(address, "0")
			}
		}
		next.ServeHTTP(w, r)
	})
}

// withWeb serves the web app of dir next to the API, so both share one origin. A path that is no file is a page of
// the app: 200.html, the fallback `nuxt generate` writes, lets its router show it.
func withWeb(dir string, api http.Handler) http.Handler {
	files := http.FileServer(http.Dir(dir))
	mux := http.NewServeMux()
	mux.Handle("/api/", api)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if f, err := http.Dir(dir).Open(r.URL.Path); err == nil {
			f.Close()
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, "200.html"))
	})
	return mux
}
