package agent_api

import (
	"bytes"
	"errors"
	"gopkg.in/resty.v1"
	"io/ioutil"
	"log"
	"net/http"
)

func httpCaller(in *HttpRequest) ([]byte, int, http.Header, error) {

	client := new(http.Client)
	r := ioutil.NopCloser(bytes.NewReader([]byte(in.Body)))
	request, err := http.NewRequest(in.RequestType, in.Url, r)
	if err != nil {
		log.Println(err)
		return nil, 0, http.Header{}, err
	}
	///request.Header.Add("Accept-Encoding", "gzip")

	for _, header := range in.Headers {
		request.Header.Set(header.Key, header.Value)
	}

	response, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return nil, 0, http.Header{}, err
	}
	defer response.Body.Close()

	out, _ := ioutil.ReadAll(response.Body)

	return out, response.StatusCode, response.Header, nil
}

func httpCallerX(in *HttpRequest) ([]byte, int, http.Header, error) {

	httpClient := resty.New()

	var resp *resty.Response
	var err error

	request := httpClient.R()
	request.SetBody(in.Body)

	for _, header := range in.Headers {
		request.SetHeader(header.Key, header.Value)
	}

	switch in.RequestType {
	case resty.MethodPost:
		resp, err = request.Post(in.Url)
	case resty.MethodGet:
		resp, err = request.Get(in.Url)
	case resty.MethodDelete:
		resp, err = request.Delete(in.Url)
	case resty.MethodPut:
		resp, err = request.Put(in.Url)
	case resty.MethodPatch:
		resp, err = request.Patch(in.Url)

	}

	if err != nil {
		return []byte{}, 0, http.Header{}, err
	}

	if resp == nil {
		return []byte{}, 0, http.Header{}, errors.New("response is nil")
	}

	log.Println(resp.StatusCode())
	log.Println(resp.Header().Get("Content-Length"))
	log.Println(string(resp.Body()))

	return resp.Body(), resp.StatusCode(), resp.Header(), nil

}
