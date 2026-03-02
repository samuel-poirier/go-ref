package processed

import (
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service"
)

type Handler struct {
	service service.Service
}

func NewHandler(service *service.Service) Handler {
	return Handler{
		service: *service,
	}
}
