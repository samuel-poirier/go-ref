package processed

import (
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service"
)

type Handler struct {
	service service.Service
}

func NewHandler(service *service.Service) Handler {
	return Handler{
		service: *service,
	}
}
