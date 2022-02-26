package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/99designs/httpsignatures-go"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

type ConfigAPI struct {
	endpoint string
	secret   string
}

type requestStructure struct {
	Repo  *model.Repo  `json:"repo"`
	Build *model.Build `json:"build"`
}

func NewAPI(endpoint, secret string) ConfigAPI {
	return ConfigAPI{endpoint: endpoint, secret: secret}
}

func (cp *ConfigAPI) IsConfigured() bool {
	return cp.endpoint != ""
}

func (cp *ConfigAPI) FetchExternalConfig(ctx context.Context, repo *model.Repo, build *model.Build) (configData string, useOld bool, err error) {
	// Create request, needs repo
	response := struct {
		Data string `json:"data"`
	}{}

	status, err := sendRequest(ctx, "POST", cp.endpoint, cp.secret, requestStructure{Repo: repo, Build: build}, &response)
	if err != nil {
		return "", false, err
	}

	if status != 204 && status != 200 {
		return response.Data, status == 204, fmt.Errorf("Failed to fetch config")
	}

	return response.Data, status == 204, nil
}

// Send makes an http request to the given endpoint, writing the input
// to the request body and unmarshaling the output from the response body.
func sendRequest(ctx context.Context, method, path, signkey string, in, out interface{}) (statuscode int, err error) {
	uri, err := url.Parse(path)
	if err != nil {
		return 0, err
	}

	// if we are posting or putting data, we need to
	// write it to the body of the request.
	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		jsonerr := json.NewEncoder(buf).Encode(in)
		if jsonerr != nil {
			return 0, jsonerr
		}
	}

	fmt.Printf("%v", in)

	// creates a new http request to bitbucket.
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), buf)
	if err != nil {
		return 0, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Sign using the 'Signature' header
	err = httpsignatures.DefaultSha256Signer.SignRequest("hmac-key", signkey, req)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}
		fmt.Print(string(body))
		return resp.StatusCode, nil
	}

	// if a json response is expected, parse and return
	// the json response.
	if out != nil {
		return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
	}

	return resp.StatusCode, nil
}
