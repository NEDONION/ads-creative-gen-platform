package handler

type CreateExperimentRequest struct {
	Name        string                     `json:"name" binding:"required"`
	ProductName string                     `json:"product_name"`
	Variants    []ExperimentVariantPayload `json:"variants" binding:"required"`
}

type ExperimentVariantPayload struct {
	CreativeID string  `json:"creative_id" binding:"required"`
	Weight     float64 `json:"weight" binding:"required"`
}

type GenerateResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
type UpdateExperimentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type TrackRequest struct {
	CreativeID uint `json:"creative_id" binding:"required"`
}
