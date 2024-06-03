package woodpeckergo

import (
	"net/http"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	apiClient "go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/client"
)

// New returns a client at the specified url.
func New(uri string) (*apiClient.WoodpeckerCIAPI, error) {
	return NewWithClient(uri, http.DefaultClient)
}

// NewWithClient returns a client at the specified url.
func NewWithClient(_uri string, httpClient *http.Client) (*apiClient.WoodpeckerCIAPI, error) {
	uri, err := url.Parse(_uri)
	if err != nil {
		return nil, err
	}

	transport := httptransport.NewWithClient(uri.Host, uri.Path, []string{"https", "http"}, httpClient)

	return apiClient.New(transport, strfmt.Default), nil
}
