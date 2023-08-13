package ns

type DomainCreateResult struct {
	Registered        bool          `xml:"Registered,attr"`
	ChargedAmount     string        `xml:"ChargedAmount,attr"`
	DomainID          int           `xml:"DomainID,attr"`
	OrderID           int           `xml:"OrderID,attr"`
	TransactionID     int           `xml:"TransactionID,attr"`
	WhoisguardEnable  string        `xml:"WhoisguardEnable,attr"`
	NonRealTimeDomain string        `xml:"NonRealTimeDomain,attr"`
	Errors            []DomainError `xml:"Errors>Error"`
}
