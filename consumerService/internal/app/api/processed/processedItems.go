package processed

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service/queries"
	_ "github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
	"github.com/samuel-poirier/go-pubsub-demo/shared/response"
)

// @Summary		Endpoint to get processed items
// @Description	Returns top 100 last processed items
// @Produce		json
// @Param			limit	query		int	false	"Number of items to return"	default(100)
// @Param			offset	query		int	false	"Number of items to skip"	default(0)
// @Success		200		{array}		repository.ProcessedItem
// @Failure		400		{object}	response.ErrorModel
// @Failure		500		{object}	response.ErrorModel
// @Router			/api/v1/items/processed [get]
func (handler *Handler) ProcessedItems(w http.ResponseWriter, r *http.Request) {

	offsetParam := r.URL.Query().Get("offset")
	limitParam := r.URL.Query().Get("limit")

	if offsetParam == "" {
		offsetParam = "0"
	}

	if limitParam == "" {
		limitParam = "100"
	}

	offset, err := strconv.Atoi(offsetParam)

	if err != nil {
		response.WriteJsonBadRequest(w, err.Error())
		return
	}
	limit, err := strconv.Atoi(limitParam)

	if err != nil {
		response.WriteJsonBadRequest(w, err.Error())
		return
	}

	items, err := handler.service.Queries.FindProcessedItemsWithPaging(r.Context(), queries.FindProcessedItemsWithPagingQuery{
		Offset: int32(offset),
		Limit:  int32(limit),
	})

	var validateErrs validator.ValidationErrors

	if errors.As(err, &validateErrs) {
		response.WriteJsonBadRequestFromValidationErrors(w, validateErrs)
		return
	}

	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}

	response.WriteJsonOk(w, items)
}
