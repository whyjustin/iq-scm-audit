package http

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type HttpClient struct {
	Username string
	Password string
	Token string
}

func (client *HttpClient) HttpGet(url string) []byte {
	return client.httpRequest("GET", "application/json", nil, url)
}

func (client *HttpClient) HttpPost(url string, body interface{}) []byte {
	jsonBytes, unmarshallError := json.Marshal(body)

	if unmarshallError != nil {
		log.Fatal(unmarshallError)
	}

	return client.httpRequest("POST", "application/json", bytes.NewBuffer(jsonBytes), url)
}

func (client *HttpClient) HttpPostXml(url string, body interface{}) []byte {
	xmlBytes, unmarshallError := xml.Marshal(body)
	if unmarshallError != nil {
		log.Fatal(unmarshallError)
	}
	return client.httpRequest("POST", "application/xml", bytes.NewBuffer(xmlBytes), url)
}

func (client *HttpClient) httpRequest(verb string, contentType string, body io.Reader, url string) []byte {
	var httpClient *http.Client
	if len(client.Token) > 0 {
		src := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: client.Token},
		)
		httpClient = oauth2.NewClient(context.Background(), src)
	} else {
		httpClient = &http.Client {}
	}
	request, requestError := http.NewRequest(verb, url, body)

	if requestError != nil {
		log.Fatal(requestError)
	}

	if len(client.Username) > 0 && len(client.Password) > 0 {
		request.SetBasicAuth(client.Username, client.Password)
	}
	request.Header.Set("Content-Type", contentType)

	response, requestError := httpClient.Do(request)

	if requestError != nil {
		log.Fatal(requestError)
	}

	defer response.Body.Close()

	responseBytes, requestError := ioutil.ReadAll(response.Body)

	if requestError != nil {
		log.Fatal(requestError)
	}

	return responseBytes
}
