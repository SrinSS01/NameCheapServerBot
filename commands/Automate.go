package commands

import (
	"NS/config"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type AutomateCommand struct {
	Command *discordgo.ApplicationCommand
	Config  *config.Config
}

var Automate = AutomateCommand{
	Command: &discordgo.ApplicationCommand{
		Name:        "automate",
		Description: "automate a task",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "domain",
				Description: "The domain to use",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "email",
				Description: "The email to use",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionAttachment,
				Name:        "file",
				Description: "The file to upload",
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "nameservers",
				Description: "The name servers to set",
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "password",
				Description: "The password to use",
			},
		},
	},
}

var autoMateRegex = regexp.MustCompile("^(?P<domain>\\w+(?:\\.\\w+)+) +(?P<localPart>[a-z0-9!#$%&'*+/=?^_{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")(?: +(?P<ns>(\\w+(?:\\.\\w+)+)(?:(?:, *| +)(\\w+(?:\\.\\w+)+))*))?(?: +(?P<passwrd>\\S+))?$")

func (c *AutomateCommand) ExecuteDash(session *discordgo.Session, messageCreate *discordgo.MessageCreate, args string) {
	matches := autoMateRegex.FindStringSubmatch(strings.TrimSpace(args))
	if len(matches) == 0 {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, "Enter valid arguments", messageCreate.Reference())
		return
	}
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	password := c.Config.DefaultPassword
	nameservers := c.Config.DefaultNameServers
	var err error
	domain := matches[autoMateRegex.SubexpIndex("domain")]
	localPart := matches[autoMateRegex.SubexpIndex("localPart")]
	ns := matches[autoMateRegex.SubexpIndex("ns")]
	passwrd := matches[autoMateRegex.SubexpIndex("passwrd")]
	attachments := messageCreate.Attachments
	parts := strings.SplitN(domain, ".", 2)

	if ns != "" {
		allStrings := NsR.FindAllString(ns, -1)
		nameservers = strings.Join(allStrings, ",")
	}
	if passwrd != "" {
		password = passwrd
	}

	apiResponse, err := RequestDomainCheck(apiUser, apiKey, userName, clientIP, domain)
	if err != nil {
		_, _ = session.ChannelMessageSendReply(messageCreate.ChannelID, err.Error(), messageCreate.Reference())
		return
	}
	/*_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.check&ClientIp=%s&DomainList=%s", apiUser, apiKey, userName, clientIP, domain)
	err =
	if err != nil {
		return
	}
	resp, err := resty.New().R().Get(_url)
	if err != nil {
		content := fmt.Sprintf("Error making the request: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	var apiResponse ns.ApiResponse
	resBody := resp.Body()
	err = xml.Unmarshal(resBody, &apiResponse)
	if err != nil {
		content := fmt.Sprintf("Error parsing the response: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	if apiResponse.Status != "OK" {
		content := fmt.Sprintf("Error in API response: \n```\n%s\n```", string(resBody))
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}*/
	for _, result := range apiResponse.CommandResponse.DomainCheckData {
		availability := "available"
		if !result.Available {
			availability = "unavailable"
		}
		msg, _ := session.ChannelMessageSendReply(messageCreate.ChannelID, fmt.Sprintf("Domain: %s, Availability: %s", result.Domain, availability), messageCreate.Reference())

		if result.Available {
			msg, err := session.ChannelMessageSendReply(msg.ChannelID, "Registering domain...", msg.Reference())
			if err != nil {
				return
			}
			if RequestRegisterDomain(c.Config, domain, msg, session) {
				// nameservers
				if len(nameservers) != 0 {
					sld, tld := parts[0], parts[1]
					res, err := RequestNameServerChange(apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
					if err != nil {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
						return
					}
					msg, _ = session.ChannelMessageSendReply(msg.ChannelID, res, msg.Reference())
					/*_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.dns.setCustom&ClientIp=%s&SLD=%s&TLD=%s&NameServers=%s", apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
					resp, err := resty.New().R().Get(_url)
					if err != nil {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error making the request: %s", err.Error()), msg.Reference())
						return
					}
					err = xml.Unmarshal(resp.Body(), &apiResponse)
					if err != nil {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error parsing the response: %s", err.Error()), msg.Reference())
						return
					}
					if apiResponse.Status == "ERROR" {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error in API response: %s", apiResponse.Status), msg.Reference())
						return
					}
					msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Name server changed successfully"), msg.Reference())*/
				}

				// addon
				res, err := RequestAddDomain(domain)
				if err != nil {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
					return
				}
				msg, _ = session.ChannelMessageSendReply(msg.ChannelID, res, msg.Reference())

				// create email
				createEmailResponse, err := RequestEmailCreate(domain, localPart, password)
				if err != nil {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
					return
				}
				emailSuccess := false
				if cpanelresult, ok := createEmailResponse.Cpanelresult.(map[string]interface{}); ok {
					if datas, _ok := cpanelresult["data"].([]interface{}); _ok {
						for _, data := range datas {
							if data.(map[string]interface{})["result"].(float64) == 0 {
								_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to create email account: %s", data.(map[string]interface{})["reason"].(string)), msg.Reference())
								emailSuccess = false
							} else {
								_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Successfully created email account: %s", localPart+"@"+domain), msg.Reference())
								emailSuccess = true
							}
						}
					} else if data, __ok := cpanelresult["data"].(map[string]interface{}); __ok {
						if data["result"].(float64) == 0 {
							_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to create email account: %s", data["reason"].(string)), msg.Reference())
							emailSuccess = false
						} else {
							_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Successfully created email account: %s", localPart+"@"+domain), msg.Reference())
							emailSuccess = true
						}
					} else {
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error casting data to type `map[string]interface{}` or `[]interface{}`, ```json\n%s\n```", cpanelresult["data"]), msg.Reference())
						emailSuccess = false
					}
				} else {
					marshal, _ := json.Marshal(createEmailResponse)
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Error casting cpanelresult to type map[string]interface{}, ```json\n"+string(marshal)+"\n```", msg.Reference())
					emailSuccess = false
				}
				if emailSuccess {
					if IsSSLInstalled(domain) {
						// SSL is installed
						_, err := session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("SSL is installed on %s", domain), msg.Reference())
						if err != nil {
							fmt.Println("Failed to respond to interaction:", err)
						}
					} else {
						// SSL is not installed
						_, err := session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("SSL is not installed on %s", domain), msg.Reference())
						if err != nil {
							fmt.Println("Failed to respond to interaction:", err)
						}
					}
				}

				// file upload
				if len(attachments) != 0 {
					for _, attachment := range attachments {
						response, err := RequestFileUpload(attachment, domain)
						if err != nil {
							_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
							return
						}
						if data, ok := response.Data.(map[string]interface{}); ok {
							if uploads, ok := data["uploads"].([]interface{}); ok {
								for _, upload := range uploads {
									if upload, ok := upload.(map[string]interface{}); ok {
										if upload["status"].(float64) == 0 {
											_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to upload file: %s", upload["reason"].(string)), msg.Reference())
										} else {
											_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Successfully uploaded file.", msg.Reference())
										}
									} else {
										_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `upload` to `map[string]interface{}`", msg.Reference())
									}
								}
							} else {
								_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `uploads` to `[]interface{}`", msg.Reference())
							}
						} else {
							_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `data` to `map[string]interface{}`", msg.Reference())
						}
					}
				}

				// redirect
				authUser := c.Config.BasicAuth.Username
				authPass := c.Config.BasicAuth.Password
				response, err := RequestAddRedirect(authUser, authPass, domain)
				if err != nil {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
					return
				}

				// Handle the response based on the structure of the CommandResponse
				if response.Status == 1 { // Assuming a status of 1 indicates success
					builder := strings.Builder{}
					builder.WriteString("Successfully added the redirect.")
					if len(response.Messages) > 0 {
						builder.WriteByte(' ')
						builder.WriteString(strings.Join(response.Messages, " "))
					}
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, builder.String(), msg.Reference())
				} else {
					errorMessage := strings.Builder{}
					errorMessage.WriteString("Failed to add the redirect.")
					if len(response.Errors) > 0 {
						errorMessage.WriteByte(' ')
						errorMessage.WriteString(strings.Join(response.Errors, " "))
					}
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, errorMessage.String(), msg.Reference())
				}
			}
		}
	}
}

func (c *AutomateCommand) Execute(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
	var err error
	apiUser := c.Config.ApiUser
	apiKey := c.Config.ApiKey
	userName := c.Config.UserName
	clientIP := c.Config.ClientIP
	applicationCommandData := interaction.ApplicationCommandData()
	options := applicationCommandData.Options
	domain := options[0].StringValue()
	email := options[1].StringValue()
	var fileId string
	nameservers := c.Config.DefaultNameServers
	password := c.Config.DefaultPassword
	parts := strings.SplitN(domain, ".", 2)

	if len(options) > 2 {
		for _, option := range options[2:] {
			switch option.Name {
			case "file":
				fileId = option.Value.(string)
			case "nameservers":
				nameservers = option.StringValue()
			case "password":
				password = option.StringValue()
			}
		}
	}
	if len(strings.TrimSpace(email)) == 0 {
		_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid email and password",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	if len(strings.TrimSpace(domain)) == 0 || len(parts) != 2 {
		_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid domain",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	localPart := email
	if len(strings.TrimSpace(password)) == 0 {
		_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please enter a valid password",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	_ = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	apiResponse, err := RequestDomainCheck(apiUser, apiKey, userName, clientIP, domain)
	if err != nil {
		content := err.Error()
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	/*_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.check&ClientIp=%s&DomainList=%s", apiUser, apiKey, userName, clientIP, domain)
	err =
	if err != nil {
		return
	}
	resp, err := resty.New().R().Get(_url)
	if err != nil {
		content := fmt.Sprintf("Error making the request: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	var apiResponse ns.ApiResponse
	resBody := resp.Body()
	err = xml.Unmarshal(resBody, &apiResponse)
	if err != nil {
		content := fmt.Sprintf("Error parsing the response: %s", err.Error())
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}
	if apiResponse.Status != "OK" {
		content := fmt.Sprintf("Error in API response: \n```\n%s\n```", string(resBody))
		_, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}*/
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
			if RequestRegisterDomain(c.Config, domain, msg, session) {
				// nameservers
				if len(nameservers) != 0 {
					sld, tld := parts[0], parts[1]
					res, err := RequestNameServerChange(apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
					if err != nil {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
						return
					}
					msg, _ = session.ChannelMessageSendReply(msg.ChannelID, res, msg.Reference())
					/*_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.dns.setCustom&ClientIp=%s&SLD=%s&TLD=%s&NameServers=%s", apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
					resp, err := resty.New().R().Get(_url)
					if err != nil {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error making the request: %s", err.Error()), msg.Reference())
						return
					}
					err = xml.Unmarshal(resp.Body(), &apiResponse)
					if err != nil {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error parsing the response: %s", err.Error()), msg.Reference())
						return
					}
					if apiResponse.Status == "ERROR" {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error in API response: %s", apiResponse.Status), msg.Reference())
						return
					}
					msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Name server changed successfully"), msg.Reference())*/
				}

				// addon
				res, err := RequestAddDomain(domain)
				if err != nil {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
					return
				}
				msg, _ = session.ChannelMessageSendReply(msg.ChannelID, res, msg.Reference())

				// create email
				createEmailResponse, err := RequestEmailCreate(domain, localPart, password)
				if err != nil {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
					return
				}
				emailSuccess := false
				if cpanelresult, ok := createEmailResponse.Cpanelresult.(map[string]interface{}); ok {
					if datas, _ok := cpanelresult["data"].([]interface{}); _ok {
						for _, data := range datas {
							if data.(map[string]interface{})["result"].(float64) == 0 {
								_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to create email account: %s", data.(map[string]interface{})["reason"].(string)), msg.Reference())
								emailSuccess = false
							} else {
								_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Successfully created email account: %s", localPart+"@"+domain), msg.Reference())
								emailSuccess = true
							}
						}
					} else if data, __ok := cpanelresult["data"].(map[string]interface{}); __ok {
						if data["result"].(float64) == 0 {
							_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to create email account: %s", data["reason"].(string)), msg.Reference())
							emailSuccess = false
						} else {
							_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Successfully created email account: %s", localPart+"@"+domain), msg.Reference())
							emailSuccess = true
						}
					} else {
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error casting data to type `map[string]interface{}` or `[]interface{}`, ```json\n%s\n```", cpanelresult["data"]), msg.Reference())
						emailSuccess = false
					}
				} else {
					marshal, _ := json.Marshal(createEmailResponse)
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Error casting cpanelresult to type map[string]interface{}, ```json\n"+string(marshal)+"\n```", msg.Reference())
					emailSuccess = false
				}
				if emailSuccess {
					if IsSSLInstalled(domain) {
						// SSL is installed
						_, err := session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("SSL is installed on %s", domain), msg.Reference())
						if err != nil {
							fmt.Println("Failed to respond to interaction:", err)
						}
					} else {
						// SSL is not installed
						_, err := session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("SSL is not installed on %s", domain), msg.Reference())
						if err != nil {
							fmt.Println("Failed to respond to interaction:", err)
						}
					}
				}

				// file upload
				if len(fileId) != 0 {
					attachment := applicationCommandData.Resolved.Attachments[fileId]
					response, err := RequestFileUpload(attachment, domain)
					if err != nil {
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
						return
					}
					if data, ok := response.Data.(map[string]interface{}); ok {
						if uploads, ok := data["uploads"].([]interface{}); ok {
							for _, upload := range uploads {
								if upload, ok := upload.(map[string]interface{}); ok {
									if upload["status"].(float64) == 0 {
										_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to upload file: %s", upload["reason"].(string)), msg.Reference())
									} else {
										_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Successfully uploaded file.", msg.Reference())
									}
								} else {
									_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `upload` to `map[string]interface{}`", msg.Reference())
								}
							}
						} else {
							_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `uploads` to `[]interface{}`", msg.Reference())
						}
					} else {
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `data` to `map[string]interface{}`", msg.Reference())
					}
				}

				// redirect
				authUser := c.Config.BasicAuth.Username
				authPass := c.Config.BasicAuth.Password
				response, err := RequestAddRedirect(authUser, authPass, domain)
				if err != nil {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, err.Error(), msg.Reference())
					return
				}

				// Handle the response based on the structure of the CommandResponse
				if response.Status == 1 { // Assuming a status of 1 indicates success
					builder := strings.Builder{}
					builder.WriteString("Successfully added the redirect.")
					if len(response.Messages) > 0 {
						builder.WriteByte(' ')
						builder.WriteString(strings.Join(response.Messages, " "))
					}
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, builder.String(), msg.Reference())
				} else {
					errorMessage := strings.Builder{}
					errorMessage.WriteString("Failed to add the redirect.")
					if len(response.Errors) > 0 {
						errorMessage.WriteByte(' ')
						errorMessage.WriteString(strings.Join(response.Errors, " "))
					}
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, errorMessage.String(), msg.Reference())
				}
			}
		}
	}
}
