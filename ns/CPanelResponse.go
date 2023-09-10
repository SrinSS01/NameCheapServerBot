package ns

type CPanelResponse struct {
	CPanelResult struct {
		Data  interface{} `json:"data"`
		Error string      `json:"error"`
	} `json:"cpanelresult"`
}
