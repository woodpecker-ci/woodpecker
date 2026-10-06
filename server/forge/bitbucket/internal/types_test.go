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

package internal

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorMessage(t *testing.T) {
	err := Error{Status: http.StatusTooManyRequests}
	assert.Equal(t, "bitbucket responded with HTTP 429 Too Many Requests", err.Error())

	err.Body.Message = "Resource not found"
	assert.Equal(t, "Resource not found", err.Error())
}
