package utils

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Send makes an http request to the given endpoint, writing the input
// to the request body and unmarshaling the output from the response body.
func Send(ctx context.Context, method, path, signkey string, in, out interface{}) (statuscode int, err error) {
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

	// creates a new http request to the endpoint.
	req, err := http.NewRequestWithContext(ctx, method, uri.String(), buf)
	if err != nil {
		return 0, err
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// TODO: create global server key
	_, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	// Sign using the 'Signature' header
	err = SignHTTPRequest(privEd25519Key, "woodpecker-ci-plugins", req)
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

// Error represents a http error.
type Error struct {
	code int
	text string
}

// Code returns the http error code.
func (e *Error) Code() int {
	return e.code
}

// Error returns the error message in string format.
func (e *Error) Error() string {
	return e.text
}
