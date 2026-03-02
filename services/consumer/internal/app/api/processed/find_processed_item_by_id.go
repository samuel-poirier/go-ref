package processed

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/samuel-poirier/go-ref/consumer/internal/app/service/queries"
	_ "github.com/samuel-poirier/go-ref/consumer/internal/repository"
	"github.com/samuel-poirier/go-ref/shared/response"
)

// @Summary		Endpoint get one processed item by id
// @Description	Returns the processed item found by id
// @Produce		json
// @Success		200	{object}	repository.ProcessedItem
// @Failure		400	{object}	response.ErrorModel
// @Failure		404	{object}	response.ErrorModel
// @Failure		500	{object}	response.ErrorModel
// @Router			/api/v1/items/processed/{id} [get]
// @Param			id	path	string	true	"id of the item"
func (handler *Handler) FindProcessedItemById(w http.ResponseWriter, r *http.Request) {

	idString := r.PathValue("id")
	id, err := uuid.Parse(idString)
	if err != nil {
		response.WriteJsonBadRequest(w, "invalid id format")
		return
	}

	item, err := handler.service.Queries.FindProcessedItemById(r.Context(), queries.FindProcessedItemByIdQuery{
		Id: id,
	})
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}

	if item == nil {
		response.WriteNotFound(w)
		return
	}
	response.WriteJsonOk(w, item)
}
