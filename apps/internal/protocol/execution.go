package protocol

type ExecuteRequest struct {
	Source string `json:"source"`
}

type Limits struct {
	MaxInstructions uint64 `json:"maxInstructions"`
	MaxHeapBytes    uint64 `json:"maxHeapBytes"`
	MaxArrayLength  uint64 `json:"maxArrayLength"`
	MaxOutputBytes  uint64 `json:"maxOutputBytes"`
	TimeoutMS       int64  `json:"timeoutMs"`
}

type ExecuteResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Output     string `json:"output"`
	DurationMS int64  `json:"durationMs"`
	Limits     Limits `json:"limits"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
