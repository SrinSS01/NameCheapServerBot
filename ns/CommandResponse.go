package ns

type CommandResponse struct {
	Type               string              `xml:"Type,attr"`
	DomainCheckData    []DomainCheckResult `xml:"DomainCheckResult"`
	DomainCreateResult DomainCreateResult  `xml:"DomainCreateResult"`
}
