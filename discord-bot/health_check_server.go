package main

import (
	"log/slog"
	"net/http"
	"os"
)

func HealthCheckServer() {
	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	slog.Error("Cannot start health check server on :8080")
	os.Exit(1)
}
