package ns

type DomainCheckResult struct {
	Domain                   string `xml:"Domain,attr"`
	Available                bool   `xml:"Available,attr"`
	IsPremiumName            bool   `xml:"IsPremiumName,attr"`
	PremiumRegistrationPrice string `xml:"PremiumRegistrationPrice,attr"`
	PremiumRenewalPrice      string `xml:"PremiumRenewalPrice,attr"`
	PremiumRestorePrice      string `xml:"PremiumRestorePrice,attr"`
	PremiumTransferPrice     string `xml:"PremiumTransferPrice,attr"`
	IcannFee                 string `xml:"IcannFee,attr"`
	EapFee                   string `xml:"EapFee,attr"`
}
