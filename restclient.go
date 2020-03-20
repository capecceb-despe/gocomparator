package main

import (
	"io/ioutil"
	"net/http"
)

type RestClient struct {
	client  *http.Client
	xclient string
	token   string
}

func NewRestClient(xclient string, token string) *RestClient {
	return &RestClient{
		client:  &http.Client{},
		xclient: xclient,
		token:   token,
	}
}

func (c *RestClient) get(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Client", c.xclient)
	req.Header.Add("Authorization", c.token)

	resp, err2 := c.client.Do(req)
	if err2 != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
