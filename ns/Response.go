package ns

type Response struct {
	Cpanelresult struct {
		Data struct {
			Result string `json:"result"`
			Reason string `json:"reason"`
		} `json:"data"`
	} `json:"cpanelresult"`
}
