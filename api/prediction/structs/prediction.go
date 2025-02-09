package structs

type PredictionRequest struct {
	ImageID   string `json:"image_id"`
	RequestID string `json:"request_id"`
}

type StyleProbability struct {
	Style       string  `json:"style"`
	Probability float64 `json:"probability"`
}

type PredictionResponse struct {
	MainPrediction struct {
		Style      string  `json:"style"`
		Confidence float64 `json:"confidence"`
	} `json:"main_prediction"`
	AllProbabilities []StyleProbability `json:"all_probabilities"`
	RequestID        string             `json:"request_id"`
	Error            string             `json:"error,omitempty"`
}
