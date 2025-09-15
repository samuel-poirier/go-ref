package hello

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/sam9291/go-pubsub-demo/events"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/domain"
	"github.com/sam9291/go-pubsub-demo/shared/response"
)

type Handler struct {
	publisher *domain.Publisher
}

func NewHandler(publisher *domain.Publisher) Handler {
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

	publisher := *handler.publisher
	message := events.Message{
		Id:   uuid.NewString(),
		Data: "PUBLISHED FROM HELLO WORLD ENDPOINT",
	}
	if publisher != nil {
		publisher.Publish(message)
	}

	response.WriteJsonOk(w, message)
}
