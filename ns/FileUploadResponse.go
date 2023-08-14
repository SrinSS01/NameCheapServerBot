package ns

type FileUploadResponse struct {
	Errors []string    `json:"errors"`
	Data   interface{} `json:"data"`
}
