package commands

import (
	"NS/config"
	"NS/ns"
	"NS/util"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/go-resty/resty/v2"
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
		//password = options[4].StringValue()
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
	escapedDomain := url.QueryEscape(domain)
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
	_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.check&ClientIp=%s&DomainList=%s", apiUser, apiKey, userName, clientIP, domain)
	err = session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
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
			if RegisterDomain(c.Config, domain, msg, session) {
				// nameservers
				if len(nameservers) != 0 {
					sld, tld := parts[0], parts[1]
					_url := fmt.Sprintf("https://api.namecheap.com/xml.response?ApiUser=%s&ApiKey=%s&UserName=%s&Command=namecheap.domains.dns.setCustom&ClientIp=%s&SLD=%s&TLD=%s&NameServers=%s", apiUser, apiKey, userName, clientIP, sld, tld, nameservers)
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
					msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Name server changed successfully"), msg.Reference())
				}

				// addon
				addonURL := fmt.Sprintf("https://199.188.203.195:2083/json-api/cpanel?cpanel_jsonapi_func=addaddondomain&cpanel_jsonapi_module=AddonDomain&cpanel_jsonapi_version=2&newdomain=%s&subdomain=%s&dir=/home/swapped2/%s", escapedDomain, escapedDomain, escapedDomain)
				response, err := util.MakeRequest("GET", addonURL, "", nil)
				if err != nil {
					msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error making the request: %s", err.Error()), msg.Reference())
					return
				}
				var addDomainResponse ns.Response
				err = json.Unmarshal(response, &addDomainResponse)
				if err != nil {
					msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error unmarshalling JSON ```json\n%s```: %s", response, err.Error()), msg.Reference())
					return
				}
				if cpanelresult, ok := addDomainResponse.Cpanelresult.(map[string]interface{}); ok {
					success := false
					if data, ok := cpanelresult["data"].(map[string]interface{}); ok {
						if data["result"].(string) == "0" {
							msg, _ = session.ChannelMessageSendReply(msg.ChannelID, "Failed to add domain: "+data["reason"].(string), msg.Reference())
							return
						}
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, "Successfully added domain: "+domain, msg.Reference())
						success = true
					} else if datas, ok := cpanelresult["data"].([]interface{}); ok {
						for _, data := range datas {
							if data.(map[string]interface{})["result"].(float64) == 0 {
								msg, _ = session.ChannelMessageSendReply(msg.ChannelID, "Failed to add domain: "+data.(map[string]interface{})["reason"].(string), msg.Reference())
								return
							}
							msg, _ = session.ChannelMessageSendReply(msg.ChannelID, "Successfully added domain: "+domain, msg.Reference())
							success = true
						}
					} else {
						msg, _ = session.ChannelMessageSendReply(msg.ChannelID, "Error casting data to type []interface{} or map[string]interface{}", msg.Reference())
						return
					}
					// create email
					if success {
						emailURL := fmt.Sprintf("https://199.188.203.195:2083/json-api/cpanel?cpanel_jsonapi_func=addpop&cpanel_jsonapi_module=Email&cpanel_jsonapi_version=2&domain=%s&email=%s&password=%s", domain, localPart, password)
						response, err := util.MakeRequest("GET", emailURL, "", nil)
						if err != nil {
							return
						}
						var emailCreateResponse ns.Response
						err = json.Unmarshal(response, &emailCreateResponse)
						if err != nil {
							return
						}
						if cpanelresult, ok := emailCreateResponse.Cpanelresult.(map[string]interface{}); ok {
							emailSuccess := false
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
								if data["result"].(string) == "0" {
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
						} else {
							_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Error casting cpanelresult to type map[string]interface{}, ```json\n"+string(response)+"\n```", msg.Reference())
						}
					}

					// upload file
					if len(fileId) != 0 {
						attachment := applicationCommandData.Resolved.Attachments[fileId]
						fileName := attachment.Filename
						fileURL := attachment.URL
						res, err := http.DefaultClient.Get(fileURL)
						if err != nil {
							msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error during http.Get(): %s", err.Error()), msg.Reference())
							return
						}
						body := &bytes.Buffer{}
						writer := multipart.NewWriter(body)
						part, err := writer.CreateFormFile("file-1", fileName)
						if err != nil {
							msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error during writer.CreateFormFile(): %s", err.Error()), msg.Reference())
							return
						}
						_, err = io.Copy(part, res.Body)
						uploadURL := fmt.Sprintf("https://199.188.203.195:2083/execute/Fileman/upload_files?dir=/home/swapped2/%s", escapedDomain)
						response, err := util.MakeRequest("POST", uploadURL, writer.FormDataContentType(), body)
						if err != nil {
							msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error during util.MakeRequest(): %s", err.Error()), msg.Reference())
							return
						}
						var fileUploadResponse ns.FileUploadResponse
						err = json.Unmarshal(response, &fileUploadResponse)
						if err != nil {
							msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Error during json.Unmarshal(): %s", err.Error()), msg.Reference())
							return
						}
						if len(fileUploadResponse.Errors) > 0 {
							var contentBuilder strings.Builder
							for i, errors := range fileUploadResponse.Errors {
								contentBuilder.WriteString(fmt.Sprintf("Error `%d`: ```\n%s```\n", i, errors))
							}
							msg, _ = session.ChannelMessageSendReply(msg.ChannelID, contentBuilder.String(), msg.Reference())
							return
						}
						if data, ok := fileUploadResponse.Data.(map[string]interface{}); ok {
							if uploads, ok := data["uploads"].([]interface{}); ok {
								for _, upload := range uploads {
									if upload, ok := upload.(map[string]interface{}); ok {
										if upload["status"].(float64) == 0 {
											msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to upload file: %s", upload["reason"].(string)), msg.Reference())
										} else {
											msg, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Successfully uploaded file: %s", fileName), msg.Reference())
										}
									} else {
										msg, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `upload` to `map[string]interface{}`", msg.Reference())
									}
								}
							} else {
								msg, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `uploads` to `[]interface{}`", msg.Reference())
							}
						} else {
							msg, _ = session.ChannelMessageSendReply(msg.ChannelID, "Unable to cast `data` to `map[string]interface{}`", msg.Reference())
						}
					}

					// redirect
					redirectURL := "https://" + domain
					_url := fmt.Sprintf("https://%s:%s@199.188.203.195:2083/execute/Mime/add_redirect?domain=%s&redirect=%s&redirect_wildcard=0&redirect_www=0&src=/&type=permanent", "swapped2", "Mmady5113x", domain, redirectURL)
					// Create a new HTTP client that skips SSL certificate verification
					tr := &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					}
					client := &http.Client{Transport: tr}

					// Send the GET request
					resp, err := client.Get(_url)
					if err != nil {
						// content := fmt.Sprintf("Failed to add redirect for domain: %s. Error: %s", domain, err.Error())
						// _, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
						// 	Content: &content,
						// })
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to add redirect for domain: %s. Error: %s", domain, err.Error()), msg.Reference())
						return
					}
					defer func(Body io.ReadCloser) {
						err := Body.Close()
						if err != nil {
							fmt.Println("Error closing response body:", err.Error())
						}
					}(resp.Body)
					// Read and parse the response
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						// content := fmt.Sprintf("Failed to read response body: %s", err.Error())
						// _, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
						// 	Content: &content,
						// })
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to read response body: %s", err.Error()), msg.Reference())
						return
					}

					var response ns.RedirectCommandResponse
					fmt.Println("Raw Response:", string(body))
					err = json.Unmarshal(body, &response)
					if err != nil {
						// content := fmt.Sprintf("Failed to parse response: %s", err.Error())
						// _, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
						// 	Content: &content,
						// })
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Failed to parse response: %s", err.Error()), msg.Reference())
						return
					}

					// Handle the response based on the structure of the CommandResponse
					if response.Status == 1 { // Assuming a status of 1 indicates success
						successMessage := "Successfully added the redirect."
						if len(response.Messages) > 0 {
							successMessage += " " + strings.Join(response.Messages, " ")
						}
						// _, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
						// 	Content: &successMessage,
						// })
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, successMessage, msg.Reference())
					} else {
						errorMessage := "Failed to add the redirect."
						if len(response.Errors) > 0 {
							errorMessage += " " + strings.Join(response.Errors, " ")
						}
						// _, _ = session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
						// 	Content: &errorMessage,
						// })
						_, _ = session.ChannelMessageSendReply(msg.ChannelID, errorMessage, msg.Reference())
					}
				} else {
					_, _ = session.ChannelMessageSendReply(msg.ChannelID, "Error casting cpanelresult to type map[string]interface{}, ```json\n"+string(response)+"```", msg.Reference())
					return
				}
			}
		}
	}
}
