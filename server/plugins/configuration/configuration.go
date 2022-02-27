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

type ConfigService struct {
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

type responseStructure struct {
	Configs []config `json:"pipelines"`
}

func NewAPI(endpoint, secret string) ConfigService {
	return ConfigService{endpoint: endpoint, secret: secret}
}

func (cp *ConfigService) IsConfigured() bool {
	return cp.endpoint != ""
}

func (cp *ConfigService) FetchExternalConfig(ctx context.Context, repo *model.Repo, build *model.Build, currentFileMeta []*remote.FileMeta) (configData []*remote.FileMeta, useOld bool, err error) {
	currentConfigs := make([]*config, len(currentFileMeta))
	for i, pipe := range currentFileMeta {
		currentConfigs[i] = &config{Name: pipe.Name, Data: string(pipe.Data)}
	}

	response, status, err := sendRequest(ctx, "POST", cp.endpoint, cp.secret, requestStructure{Repo: repo, Build: build, Configuration: currentConfigs})
	if err != nil {
		return nil, false, fmt.Errorf("Failed to fetch config via http (%d) %w", status, err)
	}

	var newFileMeta []*remote.FileMeta
	if response != nil {
		newFileMeta = make([]*remote.FileMeta, len(response.Configs))
		for i, pipe := range response.Configs {
			newFileMeta[i] = &remote.FileMeta{Name: pipe.Name, Data: []byte(pipe.Data)}
		}
	} else {
		newFileMeta = make([]*remote.FileMeta, 0)
	}

	return newFileMeta, status == 204, nil
}

func sendRequest(ctx context.Context, method, path, signkey string, in interface{}) (response *responseStructure, statuscode int, err error) {
	uri, err := url.Parse(path)
	if err != nil {
		return nil, 0, err
	}

	// if we are posting or putting data, we need to
	// write it to the body of the request.
	var buf io.ReadWriter
	if in != nil {
		buf = new(bytes.Buffer)
		jsonerr := json.NewEncoder(buf).Encode(in)
		if jsonerr != nil {
			return nil, 0, jsonerr
		}
	}

	// creates a new http request to bitbucket.
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), buf)
	if err != nil {
		return nil, 0, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Sign using the 'Signature' header
	err = httpsignatures.DefaultSha256Signer.SignRequest("hmac-key", signkey, req)
	if err != nil {
		return nil, 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, err
		}

		return nil, resp.StatusCode, fmt.Errorf("Response: %s", string(body))
	}

	if resp.StatusCode == 204 {
		return nil, resp.StatusCode, nil
	}

	// if no other errors parse and return the json response.
	decodedResponse := new(responseStructure)
	err = json.NewDecoder(resp.Body).Decode(decodedResponse)
	return decodedResponse, resp.StatusCode, err
}
