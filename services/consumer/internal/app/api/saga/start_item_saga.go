package saga

import (
	"encoding/json"
	"net/http"

	"github.com/samuel-poirier/go-ref/consumer/internal/app/service"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/commands"
	"github.com/samuel-poirier/go-ref/shared/response"
)

type startItemSagaRequest struct {
	Data string `json:"data"`
}

// @Summary		Start a new item saga
// @Description	Triggers an orchestrated saga that processes an item through pending, processing, and terminal states
// @Accept		json
// @Produce		json
// @Param		body	body		startItemSagaRequest	true	"Item data"
// @Success		202		{object}	map[string]string
// @Failure		400		{object}	response.ErrorModel
// @Failure		500		{object}	response.ErrorModel
// @Router		/api/v1/saga/items [post]
func (h Handler) StartItemSaga(w http.ResponseWriter, r *http.Request) {
	var req startItemSagaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteJsonBadRequest(w, "invalid request body")
		return
	}
	if req.Data == "" {
		response.WriteJsonBadRequest(w, "data is required")
		return
	}

	var sagaIDStr string
	err := h.service.RunWithUnitOfWork(r.Context(), func(s service.Service) error {
		id, err := s.Commands.SagaCreate(r.Context(), commands.SagaCreateCommand{Data: req.Data})
		sagaIDStr = id.String()
		return err
	})
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"saga_id": sagaIDStr})
}
