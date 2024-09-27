// Copyright 2022 Woodpecker Authors
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

package woodpecker

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LogLevel(t *testing.T) {
	logLevel := "warn"
	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var ll LogLevel
			if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&ll)) {
				return
			}
			logLevel = ll.Level
		}

		_, err := fmt.Fprintf(w, `{
			"log-level": "%s"
	}`, logLevel)
		assert.NoError(t, err)
	}

	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	client := NewClient(ts.URL, http.DefaultClient)

	curLvl, err := client.LogLevel()
	assert.NoError(t, err)
	assert.True(t, strings.EqualFold(curLvl.Level, logLevel))

	newLvl, err := client.SetLogLevel(&LogLevel{Level: "trace"})
	assert.NoError(t, err)
	assert.True(t, strings.EqualFold(newLvl.Level, logLevel))
}
