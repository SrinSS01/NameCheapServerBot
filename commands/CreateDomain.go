package commands

import (
	"NS/config"
	"NS/ns"
	"encoding/xml"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-resty/resty/v2"
	"regexp"
	"strings"
)

type CreateDomainCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var CreateDomain = CreateDomainCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "create-domain",
		Description: "Create a new domain",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "The domain to create",
				Required:    true,
			},
		},
	},
}

func RequestCreateDomain(apiUser, apiKey, userName, clientIP, domain string) (*ns.ApiResponse, error) {
	apiUrl := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.check&ClientIp=%s&DomainList=%s", apiUser, apiKey, userName, clientIP, domain)
	resp, err := resty.New().R().Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("error making the request: %s", err.Error())
	}
	var apiResponse ns.ApiResponse
	body := resp.Body()
	err = xml.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing the response: %s", err.Error())
	}
	if apiResponse.Status != "OK" {
		return nil, fmt.Errorf("error in API response: \n```xml\n%s\n```", string(body))
	}
	return &apiResponse, nil
}

func (c *CreateDomainCommand) ExecuteDash(s *discordgo.Session, m *discordgo.MessageCreate, domain string) {
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	matched, err := regexp.MatchString("^\\w+(?:\\.\\w+)+$", domain)
	if err != nil || !matched {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, "Wrong domain format", m.Reference())
		return
	}
	apiResponse, err := RequestCreateDomain(apiUser, apiKey, userName, clientIP, domain)
	if err != nil {
		_, _ = s.ChannelMessageSendReply(m.ChannelID, err.Error(), m.Reference())
		return
	}

	for _, result := range apiResponse.CommandResponse.DomainCheckData {
		availability := "available"
		if !result.Available {
			availability = "unavailable"
		}
		msg, _ := s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("Domain: %s, Availability: %s", result.Domain, availability), m.Reference())

		if result.Available {
			msg, err := s.ChannelMessageSendReply(msg.ChannelID, "Registering domain...", msg.Reference())
			if err != nil {
				return
			}
			RegisterDomain(c.Config, result.Domain, msg, s)
		}
	}
}

func (c *CreateDomainCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	options := interaction.ApplicationCommandData().Options
	domain := options[0].StringValue()
	if len(strings.TrimSpace(domain)) == 0 {
		err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid domain name",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			return
		}
		return
	}
	//_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.check&ClientIp=%s&DomainList=%s", apiUser, apiKey, userName, clientIP, domain)
	// discord defer reply
	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return
	}
	//resp, err := resty.New().R().Get(_url)
	//if err != nil {
	//	content := fmt.Sprintf("Error making the request: %s", err.Error())
	//	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
	//		Content: &content,
	//	})
	//	return
	//}
	//
	//var apiResponse ns.ApiResponse
	//body := resp.Body()
	//err = xml.Unmarshal(body, &apiResponse)
	//if err != nil {
	//	content := fmt.Sprintf("Error parsing the response: %s", err.Error())
	//	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
	//		Content: &content,
	//	})
	//	return
	//}
	//
	//if apiResponse.Status != "OK" {
	//	content := fmt.Sprintf("Error in API response: \n```xml\n%s\n```", string(body))
	//	_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
	//		Content: &content,
	//	})
	//	return
	//}
	apiResponse, err := RequestCreateDomain(apiUser, apiKey, userName, clientIP, domain)
	if err != nil {
		content := err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	for _, result := range apiResponse.CommandResponse.DomainCheckData {
		availability := "available"
		if !result.Available {
			availability = "unavailable"
		}
		content := fmt.Sprintf("Domain: %s, Availability: %s", result.Domain, availability)

		msg, _ := session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})

		if result.Available {
			msg, err := session.ChannelMessageSendReply(msg.ChannelID, "Registering domain...", msg.Reference())
			if err != nil {
				return
			}
			RegisterDomain(c.Config, result.Domain, msg, session)
		}
	}
}

func RegisterDomain(c *config.Config, domain string, msg *discordgo.Message, session *discordgo.Session) bool {
	// Implement the domain registration API call here using the HTTP POST method.
	// Make an HTTP request to the NameCheap API for domain registration.

	_url := "https://api.namecheap.com/xml.response"
	params := map[string]string{
		"ApiUser":    c.ApiUser,
		"ApiKey":     c.ApiKey,
		"UserName":   c.UserName,
		"ClientIp":   c.ClientIP,
		"DomainName": domain,
		"Command":    "namecheap.domains.create",

		"Years":         "1", // Default value: 2, you can modify this as needed.
		"PromotionCode": "",  // You can add a promotional code here if available.
		// Registrant Contact Information
		"RegistrantOrganizationName":    config.ContactOrganizationName,
		"RegistrantJobTitle":            config.ContactJobTitle,
		"RegistrantFirstName":           "John",
		"RegistrantLastName":            "Doe",
		"RegistrantAddress1":            config.ContactAddress1,
		"RegistrantAddress2":            config.ContactAddress2,
		"RegistrantCity":                config.ContactCity,
		"RegistrantStateProvince":       config.ContactStateProvince,
		"RegistrantStateProvinceChoice": config.ContactStateProvinceChoice,
		"RegistrantPostalCode":          config.ContactPostalCode,
		"RegistrantCountry":             config.ContactCountry,
		"RegistrantPhone":               config.ContactPhone,
		"RegistrantPhoneExt":            config.ContactPhoneExt,
		"RegistrantFax":                 config.ContactFax,
		"RegistrantEmailAddress":        config.ContactEmailAddress,

		// Tech Contact Information
		"TechOrganizationName":    config.ContactOrganizationName,
		"TechJobTitle":            config.ContactJobTitle,
		"TechFirstName":           "Fred",
		"TechLastName":            "Johnson",
		"TechAddress1":            config.ContactAddress1,
		"TechAddress2":            config.ContactAddress2,
		"TechCity":                config.ContactCity,
		"TechStateProvince":       config.ContactStateProvince,
		"TechStateProvinceChoice": config.ContactStateProvinceChoice,
		"TechPostalCode":          config.ContactPostalCode,
		"TechCountry":             config.ContactCountry,
		"TechPhone":               config.ContactPhone,
		"TechPhoneExt":            config.ContactPhoneExt,
		"TechFax":                 config.ContactFax,
		"TechEmailAddress":        config.ContactEmailAddress,

		// Admin Contact Information
		"AdminOrganizationName":    config.ContactOrganizationName,
		"AdminJobTitle":            config.ContactJobTitle,
		"AdminFirstName":           "Alice",
		"AdminLastName":            "Smith",
		"AdminAddress1":            config.ContactAddress1,
		"AdminAddress2":            config.ContactAddress2,
		"AdminCity":                config.ContactCity,
		"AdminStateProvince":       config.ContactStateProvince,
		"AdminStateProvinceChoice": config.ContactStateProvinceChoice,
		"AdminPostalCode":          config.ContactPostalCode,
		"AdminCountry":             config.ContactCountry,
		"AdminPhone":               config.ContactPhone,
		"AdminPhoneExt":            config.ContactPhoneExt,
		"AdminFax":                 config.ContactFax,
		"AdminEmailAddress":        config.ContactEmailAddress,

		// Optional parameters for other contacts (if needed)
		"AuxBillingOrganizationName":    "",
		"AuxBillingJobTitle":            "",
		"AuxBillingFirstName":           "James",
		"AuxBillingLastName":            "Anderson",
		"AuxBillingAddress1":            config.ContactAddress1,
		"AuxBillingAddress2":            config.ContactAddress2,
		"AuxBillingCity":                config.ContactCity,
		"AuxBillingStateProvince":       config.ContactStateProvince,
		"AuxBillingStateProvinceChoice": config.ContactStateProvinceChoice,
		"AuxBillingPostalCode":          config.ContactPostalCode,
		"AuxBillingCountry":             config.ContactCountry,
		"AuxBillingPhone":               config.ContactPhone,
		"AuxBillingPhoneExt":            config.ContactPhoneExt,
		"AuxBillingFax":                 config.ContactFax,
		"AuxBillingEmailAddress":        "auxbilling@example.com",

		// Billing Contact Information
		"BillingFirstName":           "Billing",
		"BillingLastName":            "User",
		"BillingAddress1":            config.ContactAddress1,
		"BillingAddress2":            config.ContactAddress2,
		"BillingCity":                config.ContactCity,
		"BillingStateProvince":       config.ContactStateProvince,
		"BillingStateProvinceChoice": config.ContactStateProvinceChoice,
		"BillingPostalCode":          config.ContactPostalCode,
		"BillingCountry":             config.ContactCountry,
		"BillingPhone":               config.ContactPhone,
		"BillingPhoneExt":            config.ContactPhoneExt,
		"BillingFax":                 config.ContactFax,
		"BillingEmailAddress":        config.ContactEmailAddress,

		// Additional parameters for special TLDs
		"IdnCode": "",

		// Additional attributes for certain TLDs
		"Extended attributes": "",

		// Nameservers for the domain (if custom nameservers are required)
		"Nameservers": "",

		// Privacy settings for the domain
		"AddFreeWhoisguard": "no", // Default value: no, you can set this to "yes" if you want to add free domain privacy.
		"WGEnabled":         "no", // Default value: no, you can set this to "yes" if you want to enable free domain privacy.
	}

	resp, err := resty.New().R().SetFormData(params).Post(_url)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Error registering the domain: "+err.Error(), msg.Reference())
		return false
	}

	var apiResponse ns.ApiResponse
	err = xml.Unmarshal(resp.Body(), &apiResponse)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Error parsing the registration response: "+err.Error(), msg.Reference())
		return false
	}

	if apiResponse.Status != "OK" {
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Error in registration API response: "+apiResponse.Status, msg.Reference())
		return false
	}
	if apiResponse.CommandResponse.DomainCreateResult.Registered {
		content := fmt.Sprintf("Domain %s has been registered successfully!\n", domain)
		_, _ = session.ChannelMessageSendReply(msg.ChannelID, content, msg.Reference())
		return true
	} else {
		var builder strings.Builder
		builder.WriteString("Domain registration failed. Please check the response details.\n")
		if len(apiResponse.CommandResponse.DomainCreateResult.Errors) > 0 {
			builder.WriteString("Error details:")
			for _, err := range apiResponse.CommandResponse.DomainCreateResult.Errors {
				_, err := fmt.Fprintf(&builder, "Code: %d, Description: %s\n", err.Code, err.Description)
				if err != nil {
					return false
				}
			}
			_, _ = session.ChannelMessageSendReply(msg.ChannelID, builder.String(), msg.Reference())
		}
		return false
	}
}
