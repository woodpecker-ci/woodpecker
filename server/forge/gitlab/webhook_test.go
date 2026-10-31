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

package gitlab

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
)

func TestVerifyWebhook(t *testing.T) {
	// test vector of TestWebhookSign in the Standard Webhooks reference library,
	// https://github.com/standard-webhooks/standard-webhooks/blob/main/libraries/go/webhook_test.go
	key, err := base64.StdEncoding.DecodeString("MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw")
	require.NoError(t, err)
	const (
		id        = "msg_p5jXN8AQM9LWM0D4loKWxJek"
		timestamp = "1614265330"
		body      = `{"test": 2432232314}`
		signature = "v1,g0hM9SsE+OTPJTGt/tmIKtSyZlE3uFJELVlNIOLJ1OE="
	)
	sent := time.Unix(1614265330, 0)

	for _, tc := range []struct {
		name   string
		header map[string]string
		body   string
		now    time.Time
		valid  bool
	}{
		{name: "valid", valid: true},
		{name: "one of several signatures matches", header: map[string]string{"webhook-signature": "v2,unknown v1,bad " + signature}, valid: true},
		{name: "five minutes late", now: sent.Add(5 * time.Minute), valid: true},
		{name: "five minutes early", now: sent.Add(-5 * time.Minute), valid: true},
		{name: "changed body", body: body + " "},
		{name: "changed id", header: map[string]string{"webhook-id": "msg_other"}},
		{name: "changed timestamp", header: map[string]string{"webhook-timestamp": "1614265331"}},
		{name: "too late", now: sent.Add(5*time.Minute + time.Second)},
		{name: "too early", now: sent.Add(-5*time.Minute - time.Second)},
		{name: "missing signature", header: map[string]string{"webhook-signature": ""}},
		{name: "missing id", header: map[string]string{"webhook-id": ""}},
		{name: "invalid timestamp", header: map[string]string{"webhook-timestamp": "yesterday"}},
		{name: "other version only", header: map[string]string{"webhook-signature": strings.Replace(signature, "v1,", "v2,", 1)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			header := http.Header{}
			header.Set("webhook-id", id)
			header.Set("webhook-timestamp", timestamp)
			header.Set("webhook-signature", signature)
			for name, value := range tc.header {
				header.Set(name, value)
			}
			payload := body
			if tc.body != "" {
				payload = tc.body
			}
			now := sent
			if !tc.now.IsZero() {
				now = tc.now
			}

			err := verifyWebhook(header, []byte(payload), key, now)
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, types.ErrInvalidWebhookSignature)
			}
		})
	}
}
