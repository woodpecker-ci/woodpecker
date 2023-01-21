// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

func (f *Forgejo) GetSemVer() (string, error) {
	v := struct {
		Version string `json:"version"`
	}{}
	_, err := f.getParsedResponse("GET", "/forgejo/v1/version", nil, nil, &v)
	return v.Version, err
}
