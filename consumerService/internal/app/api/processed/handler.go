package processed

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service"
	"github.com/samuel-poirier/go-pubsub-demo/consumer/internal/app/service/queries"
	_ "github.com/samuel-poirier/go-pubsub-demo/consumer/internal/repository"
	"github.com/samuel-poirier/go-pubsub-demo/shared/response"
)

type Handler struct {
	service service.Service
}

func NewHandler(service *service.Service) Handler {
	return Handler{
		service: *service,
	}
}

// @Summary		Endpoint to get processed items
// @Description	Returns top 100 last processed items
// @Produce		json
// @Param limit query int false "Number of items to return" default(100)
// @Param offset query int false "Number of items to skip" default(0)
// @Success		200	{array} repository.ProcessedItem
// @Success		400	{object} response.ErrorModel
// @Success		500	{object} response.ErrorModel
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
		response.WriteJsonBadRequest(w, validateErrs.Error())
		return
	}

	if err != nil {
		response.WriteInternalServerError(w, err.Error())
		return
	}

	response.WriteJsonOk(w, items)
}

// @Summary		Endpoint to count processed items
// @Description	Returns count of all processed items
// @Produce		json
// @Success		200	{number} count
// @Success		500	{object} response.ErrorModel
// @Router			/api/v1/items/processed/count [get]
func (handler *Handler) CountProcessedItems(w http.ResponseWriter, r *http.Request) {
	count, err := handler.service.Queries.CountAllProcessedItems(r.Context(), queries.CountAllProcessedItemsQuery{})
	if err != nil {
		response.WriteInternalServerError(w, err.Error())
	} else {
		response.WriteJsonOk(w, count)
	}
}

// @Summary		Endpoint get one processed item by id
// @Description	Returns the processed item found by id
// @Produce		json
// @Success		200	{object} repository.ProcessedItem
// @Success		400	{object} response.ErrorModel
// @Success		404	{object} response.ErrorModel
// @Success		500	{object} response.ErrorModel
// @Router			/api/v1/items/processed/{id} [get]
// @Param id path string true "id of the item"
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
