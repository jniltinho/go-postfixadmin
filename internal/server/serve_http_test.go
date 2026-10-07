package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
)

func TestServeHTTPWaitsForDrain(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-release; w.WriteHeader(204) })}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- serveHTTP(ctx, srv, func() error { return srv.Serve(listener) }, slog.New(slog.NewTextHandler(io.Discard, nil)))
	}()
	requestDone := make(chan error, 1)
	go func() {
		res, err := http.Get("http://" + listener.Addr().String())
		if err == nil {
			res.Body.Close()
		}
		requestDone <- err
	}()
	<-entered
	cancel()
	select {
	case err := <-done:
		t.Fatalf("returned before request drain: %v", err)
	default:
	}
	close(release)
	if err := <-requestDone; err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, http.ErrServerClosed) {
		t.Fatal(err)
	}
}

func TestServeHTTPStartupError(t *testing.T) {
	want := errors.New("bind failed")
	err := serveHTTP(t.Context(), &http.Server{}, func() error { return want }, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if !errors.Is(err, want) {
		t.Fatal(err)
	}
}
