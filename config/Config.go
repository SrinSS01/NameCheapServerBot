package config

type Config struct {
	Token           string `json:"token"`
	ApiUser         string `json:"apiUser"`
	ApiKey          string `json:"apiKey"`
	UserName        string `json:"userName"`
	ClientIP        string `json:"clientIP"`
	DefaultPassword string `json:"defaultPassword"`
	BasicAuth       struct {
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"basicAuth"`
	DefaultNameServers string `json:"defaultNameServers"`
}

const (
	ContactOrganizationName    = "Jobo Inc."
	ContactJobTitle            = "Webmaster"
	ContactAddress1            = "123 Main Street"
	ContactAddress2            = "Suite 456"
	ContactCity                = "New York"
	ContactStateProvince       = "New York"
	ContactStateProvinceChoice = "NY"
	ContactPostalCode          = "10001"
	ContactCountry             = "US"
	ContactPhone               = "+1.85589652478"
	ContactPhoneExt            = "855"
	ContactFax                 = "+1.85589652478"
	ContactEmailAddress        = "madypanel@gmail.com"
	WHMAPIKey                  = "08F7G6UQO2YVITIKF04PDBXZIVV0XGH7"
	WHMHost                    = "https://swapped1.lat:2087"
)
