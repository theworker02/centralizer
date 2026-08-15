// Command centralizerd is an optional local daemon for long-lived bridges
// and a process-wide service registry. Library users do not need it.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/theworker02/centralizer/internal/telemetry"
	"github.com/theworker02/centralizer/internal/version"
	"github.com/theworker02/centralizer/pkg/centralizer"
	"github.com/theworker02/centralizer/pkg/diagnostics"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:4780", "loopback listen address")
	flag.Parse()

	hub := centralizer.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "ok",
			"version":  version.Version,
			"protocol": version.Protocol,
		})
	})
	mux.HandleFunc("/v1/adapters", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"adapters": hub.Adapters()})
	})
	mux.HandleFunc("/v1/services", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(hub.List())
	})
	mux.HandleFunc("/v1/doctor", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(diagnostics.Run(hub.Adapters()))
	})
	mux.HandleFunc("/v1/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(telemetry.DefaultMetrics.Snap())
	})
	mux.HandleFunc("/v1/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Name   string `json:"name"`
			Source string `json:"source"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		svc, err := hub.Connect(r.Context(), req.Source)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"name": svc.Name(), "health": svc.Health()})
	})

	srv := &http.Server{Addr: *addr, Handler: localhostOnly(mux), ReadHeaderTimeout: 5 * time.Second}
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		slog.Error("listen", "err", err)
		os.Exit(1)
	}
	slog.Info("centralizerd listening", "addr", ln.Addr().String(), "version", version.Version)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = hub.Close(shctx)
		_ = srv.Shutdown(shctx)
	}()
	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		slog.Error("serve", "err", err)
		os.Exit(1)
	}
}

func localhostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil || !localhost(host) {
			http.Error(w, "localhost only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func localhost(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
