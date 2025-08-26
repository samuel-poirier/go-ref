package hello

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/sam9291/go-pubsub-demo/events"
	"github.com/sam9291/go-pubsub-demo/publisher/internal/domain"
)

type Handler struct {
	publisher *domain.Publisher
}

func NewHandler(publisher *domain.Publisher) Handler {
	return Handler{
		publisher: publisher,
	}
}

func (handler *Handler) HelloWorld(w http.ResponseWriter, r *http.Request) {

	publisher := *handler.publisher

	if publisher != nil {
		publisher.Publish(events.Message{
			Id:   uuid.NewString(),
			Data: "PUBLISHED FROM HELLO WORLD ENDPOINT",
		})
	}

	w.Write([]byte("hello world"))
}
