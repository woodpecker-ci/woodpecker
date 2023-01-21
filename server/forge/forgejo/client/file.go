// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/url"
)

func (f *Forgejo) GetFile(owner, repo, ref, filepath string) ([]byte, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}

	filepath = pathEscapeSegments(filepath)
	ref = pathEscapeSegments(ref)
	content, resp, err := f.getResponse("GET", fmt.Sprintf("/repos/%s/%s/raw/%s?ref=%s", owner, repo, filepath, url.QueryEscape(ref)), nil, nil)
	return content, resp, err
}
