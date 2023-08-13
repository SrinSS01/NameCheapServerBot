package ns

type ApiResponse struct {
	Status          string          `xml:"Status,attr"`
	CommandResponse CommandResponse `xml:"CommandResponse"`
}
