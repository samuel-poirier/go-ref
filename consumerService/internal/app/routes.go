package app

import (
	"net/http"

	"github.com/sam9291/go-pubsub-demo/consumer/internal/api/health"
	"github.com/sam9291/go-pubsub-demo/shared/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (a *App) loadRoutes() (http.Handler, error) {
	// Create a new router
	router := http.NewServeMux()

	healthHandler := health.NewHandler()

	v1 := http.NewServeMux()
	v1.HandleFunc("GET /api/v1/hc", healthHandler.HealthCheck)

	swaggerEndpoints := http.NewServeMux()
	swaggerEndpoints.HandleFunc("GET /{path}", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8081/swagger/doc.json"), //The url pointing to API definition
	))

	router.Handle("/api/v1/", v1)
	router.Handle("/swagger/", http.StripPrefix("/swagger", swaggerEndpoints))

	// Create a middleware chain from the Chain function of the
	// middleware package
	chain := middleware.Chain(
		middleware.Logging(a.logger),
	)
	return chain(router), nil
}
