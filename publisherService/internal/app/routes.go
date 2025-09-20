package app

import (
	"fmt"
	"net/http"

	"github.com/sam9291/go-pubsub-demo/publisher/internal/api/health"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/api/hello"
	"github.com/sam9291/go-pubsub-demo/shared/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (a *App) loadRoutes() (http.Handler, error) {
	// Create a new router
	router := http.NewServeMux()

	helloHandler := hello.NewHandler(a.publisher)
	healthHandler := health.NewHandler()

	v1 := http.NewServeMux()
	v1.HandleFunc("GET /api/v1/hello", helloHandler.HelloWorld)
	v1.HandleFunc("GET /api/v1/hc", healthHandler.HealthCheck)

	swaggerEndpoints := http.NewServeMux()
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s%s/swagger/doc.json", a.config.Hostname, a.config.Addr)),
	)
	swaggerEndpoints.HandleFunc("GET /swagger/{path}", swaggerHandler)
	swaggerEndpoints.HandleFunc("GET /swagger/", swaggerHandler)

	router.Handle("/api/v1/", v1)
	router.Handle("/swagger/", swaggerEndpoints)

	// Create a middleware chain from the Chain function of the
	// middleware package
	chain := middleware.Chain(
		middleware.Logging(a.logger),
	)
	return chain(router), nil
}
