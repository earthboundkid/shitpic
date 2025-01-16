package main

import (
	"log/slog"
	"net/http"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	port := `:8080`
	slog.Default().Info("init", "port", port)
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(`static`)))
	http.ListenAndServe(port, mux)
}
