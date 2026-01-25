package hello

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/samuel-poirier/go-ref/events"
	"github.com/samuel-poirier/go-ref/shared/publisher"
	"github.com/samuel-poirier/go-ref/shared/response"
)

type Handler struct {
	publisher *publisher.Publisher
}

func NewHandler(publisher *publisher.Publisher) Handler {
	return Handler{
		publisher: publisher,
	}
}

// @Summary		Hello world endpoint
// @Description	Returns a message that gets published to rabbitmq
// @Produce		json
// @Success		200	{object}	events.Message
// @Router			/api/v1/hello [get]
func (handler *Handler) HelloWorld(w http.ResponseWriter, r *http.Request) {
	// Use request context to correlate traces
	ctx := r.Context()

	pub := *handler.publisher
	message := events.DataGeneratedEvent{
		Id:   uuid.NewString(),
		Data: "PUBLISHED FROM HELLO WORLD ENDPOINT",
	}
	if pub != nil {
		m, err := publisher.NewMessageEnvelope(message)
		if err != nil {
			response.WriteInternalServerError(w, err.Error())
			return
		} else {
			// Pass request context to link publish operation to HTTP request trace
			pub.Publish(ctx, m)
		}
	}

	response.WriteJsonOk(w, message)
}
