package ns

type DomainError struct {
	Code        int    `xml:"Code,attr"`
	Description string `xml:"Description,attr"`
}
