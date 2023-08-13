package util

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

func MakeRequest(method, url, contentType string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth("XXXXXXX", "XXXXXX")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return nil, err
	}

	return responseBody, nil
}
