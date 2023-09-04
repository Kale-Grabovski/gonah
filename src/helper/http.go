package helper

import (
	"io"
	"net/http"
)

type HttpHelper struct{}

func (s *HttpHelper) GetWithBasicAuth(host, login, password string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", host, nil)
	req.SetBasicAuth(login, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

func (s *HttpHelper) Get(uri string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", uri, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}
