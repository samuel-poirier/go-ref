package app

import (
	"fmt"
	"net/http"

	"github.com/samuel-poirier/go-ref/consumer/internal/app/api/health"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/api/processed"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service"
	"github.com/samuel-poirier/go-ref/shared/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func (a *App) loadRoutes(service *service.Service) (http.Handler, error) {
	// Create a new router
	router := http.NewServeMux()

	healthHandler := health.NewHandler()
	processedHandler := processed.NewHandler(service)

	v1 := http.NewServeMux()
	v1.HandleFunc("GET /api/v1/hc", healthHandler.HealthCheck)
	v1.HandleFunc("GET /api/v1/items/processed", processedHandler.ProcessedItems)
	v1.HandleFunc("GET /api/v1/items/processed/count", processedHandler.CountProcessedItems)
	v1.HandleFunc("GET /api/v1/items/processed/{id}", processedHandler.FindProcessedItemById)

	swaggerEndpoints := http.NewServeMux()

	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s%s/swagger/doc.json", a.config.Hostname, a.config.Addr)),
	)
	swaggerEndpoints.HandleFunc("GET /swagger/{path}", swaggerHandler)
	swaggerEndpoints.HandleFunc("GET /swagger/", swaggerHandler)
	swaggerEndpoints.HandleFunc("GET /swagger", swaggerHandler)

	router.Handle("/api/v1/", v1)
	router.Handle("/swagger/", swaggerEndpoints)

	// Create a middleware chain from the Chain function of the
	// middleware package
	chain := middleware.Chain(
		middleware.Logging(a.logger),
	)
	return chain(router), nil
}
