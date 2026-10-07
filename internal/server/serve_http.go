package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

func serveHTTP(ctx context.Context, srv *http.Server, serve func() error, local *slog.Logger) error {
	serveDone := make(chan struct{})
	drainDone := make(chan struct{})
	go func() {
		defer close(drainDone)
		select {
		case <-ctx.Done():
			slog.Info("Shutting down server…")
			shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutCancel()
			if err := srv.Shutdown(shutCtx); err != nil {
				local.Warn("HTTP shutdown did not complete")
				_ = srv.Close()
			}
		case <-serveDone:
		}
	}()

	err := serve()
	close(serveDone)
	<-drainDone
	return err
}
