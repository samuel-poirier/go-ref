package processed

import (
	"net/http"
	"strconv"

	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
	"github.com/samuel-poirier/go-pubsub-demo/shared/response"
)

type Handler struct {
	repo *repository.Queries
}

func NewHandler(repo *repository.Queries) Handler {
	return Handler{
		repo: repo,
	}
}

// @Summary		Endpoint to get processed items
// @Description	Returns top 100 last processed items
// @Produce		json
// @Param limit query int false "Number of items to return" default(100)
// @Param offset query int false "Number of items to skip" default(0)
// @Success		200	{array} repository.ProcessedItem
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

	items, err := handler.repo.FindProcessedItemsWithPaging(r.Context(), repository.FindProcessedItemsWithPagingParams{
		Offset: int32(offset),
		Limit:  int32(limit),
	})
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
	} else {
		response.WriteJsonOk(w, items)
	}
}

// @Summary		Endpoint to count processed items
// @Description	Returns count of all processed items
// @Produce		json
// @Success		200	{number} count
// @Router			/api/v1/items/processed/count [get]
func (handler *Handler) CountProcessedItems(w http.ResponseWriter, r *http.Request) {
	count, err := handler.repo.CountAllProcessedItems(r.Context())
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
	} else {
		response.WriteJsonOk(w, count)
	}
}
