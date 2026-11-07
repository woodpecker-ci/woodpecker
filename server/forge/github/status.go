// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/go-github/v93/github"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
)

// asStatusError tells the server whether a failed status request may pass if
// sent again (a 5xx, a rate limit, a dropped connection), and when.
func asStatusError(resp *github.Response, err error) error {
	var rateLimit *github.RateLimitError
	var abuse *github.AbuseRateLimitError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &rateLimit):
		return &types.StatusError{Err: err, RetryAfter: max(time.Until(rateLimit.Rate.Reset.Time), 0)}
	case errors.As(err, &abuse):
		if abuse.RetryAfter != nil {
			return &types.StatusError{Err: err, RetryAfter: *abuse.RetryAfter}
		}
		return &types.StatusError{Err: err, RetryAfter: time.Minute}
	case resp == nil || resp.Response == nil:
		return err // no answer at all: the connection broke or timed out
	case resp.StatusCode >= http.StatusInternalServerError, resp.StatusCode == http.StatusTooManyRequests:
		return err
	default:
		return &types.StatusError{Err: err, Final: true}
	}
}
