package presenters

import "github.com/tungnt127/scb-core-v2/internal/domain/entities"

type KafkaStatusPayload struct {
	RequestUUID string             `json:"requestUUID"`
	Scanner     string             `json:"scanner,omitempty"`
	Stage       string             `json:"stage"`
	Status      string             `json:"status"`
	Message     string             `json:"message"`
	Error       *KafkaErrorPayload `json:"error,omitempty"`
}

type KafkaErrorPayload struct {
	Code       string            `json:"code"`
	Stage      string            `json:"stage,omitempty"`
	Message    string            `json:"message"`
	Cause      string            `json:"cause,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
	LogExcerpt string            `json:"logExcerpt,omitempty"`
}

type StatusPresenter struct{}

func NewStatusPresenter() *StatusPresenter {
	return &StatusPresenter{}
}

func (p *StatusPresenter) Present(event entities.ScanStatusEvent) KafkaStatusPayload {
	payload := KafkaStatusPayload{
		RequestUUID: event.RequestUUID,
		Scanner:     string(event.Scanner),
		Stage:       string(event.Stage),
		Status:      event.Status,
		Message:     event.Message,
	}
	if event.Error != nil {
		payload.Error = &KafkaErrorPayload{
			Code:       string(event.Error.Code),
			Stage:      string(event.Error.Stage),
			Message:    event.Error.Message,
			Cause:      event.Error.Cause,
			Context:    event.Error.Context,
			LogExcerpt: event.Error.LogExcerpt,
		}
	}
	return payload
}
