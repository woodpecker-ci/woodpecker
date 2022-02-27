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
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

type ConfigAPI struct {
	endpoint string
	secret   string
}

// Same as remote.FileMeta but with json tags and string data
type config struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type requestStructure struct {
	Repo          *model.Repo  `json:"repo"`
	Build         *model.Build `json:"build"`
	Configuration []*config    `json:"config"`
}

func NewAPI(endpoint, secret string) ConfigAPI {
	return ConfigAPI{endpoint: endpoint, secret: secret}
}

func (cp *ConfigAPI) IsConfigured() bool {
	return cp.endpoint != ""
}

func (cp *ConfigAPI) FetchExternalConfig(ctx context.Context, repo *model.Repo, build *model.Build, currentConfig []*remote.FileMeta) (configData []*remote.FileMeta, useOld bool, err error) {
	// Create request, needs repo
	response := struct {
		Pipelines []config `json:"pipelines"`
	}{}

	currentYamls := make([]*config, len(currentConfig))
	for i, pipe := range currentConfig {
		currentYamls[i] = &config{Name: pipe.Name, Data: string(pipe.Data)}
	}

	status, err := sendRequest(ctx, "POST", cp.endpoint, cp.secret, requestStructure{Repo: repo, Build: build, Configuration: currentYamls}, &response)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to fetch config via http %w", err)
	}

	yamls := make([]*remote.FileMeta, len(response.Pipelines))
	for i, pipe := range response.Pipelines {
		yamls[i] = &remote.FileMeta{Name: pipe.Name, Data: []byte(pipe.Data)}
	}

	return yamls, status == 204, nil
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

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, err
		}

		return resp.StatusCode, fmt.Errorf("Response: %s", string(body))
	}

	// if no other errors parse and return the json response.
	if out != nil {
		return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
	}

	return resp.StatusCode, nil
}
