package ns

type FileUploadResponse struct {
	Errors []string `json:"Errors"`
	Data   struct {
		Uploads []struct {
			Status int    `json:"Status"`
			Reason string `json:"Reason"`
		} `json:"Uploads"`
	} `json:"Data"`
}
