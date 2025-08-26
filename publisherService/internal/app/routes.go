package app

import (
	"net/http"

	"github.com/sam9291/go-pubsub-demo/publisher/internal/api/hello"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/middleware"
)

func (a *App) loadRoutes() (http.Handler, error) {
	// Create a new router
	router := http.NewServeMux()

	handler := hello.NewHandler(a.publisher)

	router.HandleFunc("GET /", handler.HelloWorld)

	// Create a middleware chain from the Chain function of the
	// middleware package
	chain := middleware.Chain(
		middleware.Logging(a.logger),
	)

	return chain(router), nil
}
