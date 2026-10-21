package v1

import "github.com/sivdead/OmniBotGo/internal/entity"

// ProcessorResponse is a safe processor DTO without secret config keys (T07).
type ProcessorResponse struct {
	ID            string                 `json:"id"`
	ProcessorName string                 `json:"processor_name"`
	ProcessorType string                 `json:"processor_type"`
	Config        map[string]interface{} `json:"config,omitempty"`
	Priority      int                    `json:"priority"`
	Status        entity.Status          `json:"status"`
	CreatedAt     interface{}            `json:"created_at,omitempty"`
	UpdatedAt     interface{}            `json:"updated_at,omitempty"`
}

// ToProcessorResponse redacts secret keys from processor config.
func ToProcessorResponse(p *entity.MessageProcessor) *ProcessorResponse {
	if p == nil {
		return nil
	}

	return &ProcessorResponse{
		ID:            p.ID,
		ProcessorName: p.ProcessorName,
		ProcessorType: p.ProcessorType,
		Config:        RedactConfigSecrets(p.Config),
		Priority:      p.Priority,
		Status:        p.Status,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// ToProcessorListResponse maps a list of processors with secret redaction.
func ToProcessorListResponse(items []entity.MessageProcessor) []*ProcessorResponse {
	out := make([]*ProcessorResponse, 0, len(items))
	for i := range items {
		out = append(out, ToProcessorResponse(&items[i]))
	}

	return out
}
