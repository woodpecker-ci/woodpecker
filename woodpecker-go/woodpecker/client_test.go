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
)

func Test_QueueInfo(t *testing.T) {
	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"pending": null,
			"running": [
					{
							"id": "4696",
							"data": "",
							"labels": {
									"platform": "linux/amd64",
									"repo": "woodpecker-ci/woodpecker"
							},
							"Dependencies": [],
							"DepStatus": {},
							"RunOn": null
					}
			],
			"stats": {
					"worker_count": 3,
					"pending_count": 0,
					"waiting_on_deps_count": 0,
					"running_count": 1,
					"completed_count": 0
			},
			"Paused": false
	}`)
	}

	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	client := NewClient(ts.URL, http.DefaultClient)

	info, err := client.QueueInfo()
	if info.Stats.Workers != 3 {
		t.Errorf("Unexpected worker count: %v, %v", info, err)
	}
}

func Test_LogLevel(t *testing.T) {
	logLevel := "warn"
	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var ll LogLevel
			if err := json.NewDecoder(r.Body).Decode(&ll); err != nil {
				t.Logf("could not decode json: %v\n", err)
				t.FailNow()
			}
			logLevel = ll.Level
		}

		fmt.Fprintf(w, `{
			"log-level": "%s"
	}`, logLevel)
	}

	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	client := NewClient(ts.URL, http.DefaultClient)

	curLvl, err := client.LogLevel()
	if err != nil {
		t.Logf("could not get current log level: %v", err)
		t.FailNow()
	}

	if !strings.EqualFold(curLvl.Level, logLevel) {
		t.Logf("log level is not correct\n\tExpected: %s\n\t     Got: %s\n", logLevel, curLvl.Level)
		t.FailNow()
	}

	newLvl, err := client.SetLogLevel(&LogLevel{Level: "trace"})
	if err != nil {
		t.Logf("could not set log level: %v", err)
		t.FailNow()
	}

	if !strings.EqualFold(newLvl.Level, logLevel) {
		t.Logf("log level is not correct\n\tExpected: %s\n\t     Got: %s\n", logLevel, newLvl.Level)
		t.FailNow()
	}
}
