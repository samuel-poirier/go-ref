package processed

import (
	"net/http"

	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service/queries"
	"github.com/samuel-poirier/go-pubsub-demo/shared/response"
)

// @Summary		Endpoint to count processed items
// @Description	Returns count of all processed items
// @Produce		json
// @Success		200	{number}	count
// @Failure		500	{object}	response.ErrorModel
// @Router			/api/v1/items/processed/count [get]
func (handler *Handler) CountProcessedItems(w http.ResponseWriter, r *http.Request) {
	count, err := handler.service.Queries.CountAllProcessedItems(r.Context(), queries.CountAllProcessedItemsQuery{})
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
	} else {
		response.WriteJsonOk(w, count)
	}
}
