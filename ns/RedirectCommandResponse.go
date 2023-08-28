package ns

type RedirectCommandResponse struct {
	Status   int      `json:"status"`
	Messages []string `json:"messages"`
	Errors   []string `json:"errors"`
}
