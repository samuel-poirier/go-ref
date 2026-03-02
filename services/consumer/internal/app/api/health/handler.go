package health

import (
	"net/http"

	"github.com/samuel-poirier/go-ref/shared/response"
)

type Handler struct {
}

func NewHandler() Handler {
	return Handler{}
}

// @Summary		Health check endpoint
// @Description	Returns ok if healthy
// @Produce		json
// @Success		200	{boolean}	healthy
// @Router			/api/v1/hc [get]
func (handler *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.WriteJsonOk(w, true)
}
