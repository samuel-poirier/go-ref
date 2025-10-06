package events

type DataGeneratedEvent struct {
	Id   string `json:"id"`
	Data string `json:"data"`
}

type DataProcessedEvent struct {
	Id   string `json:"id"`
	Data string `json:"data"`
}
