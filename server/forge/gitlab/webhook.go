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

// cspell:words whsec

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
)

// webhookTolerance is how far webhook-timestamp may be off the server time,
// the value of the Standard Webhooks reference library.
const webhookTolerance = 5 * time.Minute

// webhookSigningKey derives the 32 byte key GitLab signs webhooks with from the repository secret.
func webhookSigningKey(secret string) []byte {
	key := sha256.Sum256([]byte(secret))
	return key[:]
}

// webhookSigningToken is the signing token of a GitLab hook, in the whsec_ format of Standard Webhooks.
func webhookSigningToken(secret string) string {
	return "whsec_" + base64.StdEncoding.EncodeToString(webhookSigningKey(secret))
}

// verifyWebhook checks the Standard Webhooks signature GitLab puts on the message id, timestamp and body.
func verifyWebhook(header http.Header, body, key []byte, now time.Time) error {
	id, timestamp := header.Get("webhook-id"), header.Get("webhook-timestamp")
	sent, err := strconv.ParseInt(timestamp, 10, 64)
	if id == "" || err != nil || now.Sub(time.Unix(sent, 0)).Abs() > webhookTolerance {
		return forge_types.ErrInvalidWebhookSignature
	}

	mac := hmac.New(sha256.New, key)
	_, _ = fmt.Fprintf(mac, "%s.%s.", id, timestamp)
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)

	for signature := range strings.FieldsSeq(header.Get("webhook-signature")) {
		version, value, _ := strings.Cut(signature, ",")
		received, err := base64.StdEncoding.DecodeString(value)
		if version == "v1" && err == nil && hmac.Equal(expected, received) {
			return nil
		}
	}
	return forge_types.ErrInvalidWebhookSignature
}
