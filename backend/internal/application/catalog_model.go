package application

type ProductMatchResult struct {
	Matched           bool    `json:"matched"`
	InternalProductID string  `json:"internalProductId"`
	Confidence        float64 `json:"confidence"`
	Reason            string  `json:"reason"`
}
