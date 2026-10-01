package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/relentlessworks/mathkit/internal/api"
	"github.com/relentlessworks/mathkit/internal/config"
)

func main() {
	cfg := config.Load()

	auth := api.NewAuth(cfg.Secret)
	handler := api.NewHandler(auth, cfg.Degree, cfg.NoAuth)

	mux := handler.Routes()

	// Wrap with auth middleware unless --no-auth
	var finalHandler http.Handler = mux
	if !cfg.NoAuth {
		finalHandler = auth.Middleware(mux)
	}

	fmt.Fprintf(os.Stderr, "mathkit starting on %s (degree=%v, no-auth=%v)\n",
		cfg.Addr, cfg.Degree, cfg.NoAuth)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: finalHandler,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %s", err)
	}
}
