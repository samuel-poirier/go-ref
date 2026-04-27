package events

const (
	StepProcessItem           = "process_item"
	StepCompensateProcessItem = "compensate_process_item"
)

// StartItemSagaCommand is published by the HTTP handler to kick off the orchestrator.
type StartItemSagaCommand struct {
	SagaID string `json:"saga_id"`
	Data   string `json:"data"`
}

// ProcessItemCommand is dispatched by the orchestrator to the process_item activity.
type ProcessItemCommand struct {
	SagaID string `json:"saga_id"`
	Data   string `json:"data"`
}

// CompensateProcessItemCommand is dispatched by the orchestrator when process_item fails.
type CompensateProcessItemCommand struct {
	SagaID string `json:"saga_id"`
}

// ItemProcessedSucceededEvent is published by the process_item activity on success.
type ItemProcessedSucceededEvent struct {
	SagaID        string `json:"saga_id"`
	ProcessedData string `json:"processed_data"`
}

// ItemProcessedFailedEvent is published by the process_item activity on failure.
type ItemProcessedFailedEvent struct {
	SagaID string `json:"saga_id"`
	Reason string `json:"reason"`
}

// ProcessItemCompensatedEvent is published by the compensate_process_item activity.
type ProcessItemCompensatedEvent struct {
	SagaID string `json:"saga_id"`
}
