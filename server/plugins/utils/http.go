package utils

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-fed/httpsig"
)

// Send makes an http request to the given endpoint, writing the input
// to the request body and un-marshaling the output from the response body.
func Send(ctx context.Context, method, path string, privateKey crypto.PrivateKey, in, out interface{}) (int, error) {
	if !strings.HasSuffix(path, "/") {
		path += "/" // TODO(anbraten): remove after https://github.com/go-fed/httpsig/pull/27 got merged
	}

	uri, err := url.Parse(path)
	if err != nil {
		return 0, err
	}

	// if we are posting or putting data, we need to write it to the body of the request.
	var payload io.Reader
	var body []byte
	if in != nil {
		var err error
		body, err = json.Marshal(in)
		if err != nil {
			return 0, err
		}
		payload = bytes.NewReader(body)
	}

	// creates a new http request to the endpoint.
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), payload)
	if err != nil {
		return 0, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	err = SignHTTPRequest(privateKey, req, body)
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

		return resp.StatusCode, fmt.Errorf("Response: %s", string(body))
	}

	// if no other errors parse and return the json response.
	err = json.NewDecoder(resp.Body).Decode(out)
	return resp.StatusCode, err
}

func SignHTTPRequest(privateKey crypto.PrivateKey, req *http.Request, body []byte) error {
	pubKeyID := "woodpecker-ci-plugins"

	prefs := []httpsig.Algorithm{httpsig.ED25519}
	headers := []string{httpsig.RequestTarget, "date"}
	if body != nil {
		headers = append(headers, "digest", "content-type")
	}
	signer, _, err := httpsig.NewSigner(prefs, httpsig.DigestSha256, headers, httpsig.Signature, 0)
	if err != nil {
		return err
	}

	req.Header.Add("date", time.Now().UTC().Format(http.TimeFormat))

	err = signer.SignRequest(privateKey, pubKeyID, req, body)
	return err
}
