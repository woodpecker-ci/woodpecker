// Copyright (C) 2017 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpsig

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// ctxKeyIDType is the type used to retrieve the KeyId parametes extracted from the HTTP headers
// and set into the request.Context during call of verifier.Verify
type ctxKeyIDType struct{}

var ctxKeyIDKey = &ctxKeyIDType{}

// WithKeyID retrieves the KeyId parameter from the requests
func WithKeyID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyIDKey, id)
}

// KeyIDFromContext returns the request ID from the context.
// A zero ID is returned if there are no identifers in the
// current context.
func KeyIDFromContext(ctx context.Context) string {
	v := ctx.Value(ctxKeyIDKey)
	if v == nil {
		return ""
	}
	return v.(string)
}

// RequireSignature is a http middleware that ensure the incoming request have
// the required signature using verifier v
func RequireSignature(h http.Handler, v *Verifier, realm string) (
	out http.Handler) {

	var challengeParams []string
	if realm != "" {
		challengeParams = append(challengeParams,
			fmt.Sprintf("realm=%q", realm))
	}
	if headers := v.RequiredHeaders(); len(headers) > 0 {
		challengeParams = append(challengeParams,
			fmt.Sprintf("headers=%q", strings.Join(headers, " ")))
	}

	challenge := "Signature"
	if len(challengeParams) > 0 {
		challenge += fmt.Sprintf(" %s", strings.Join(challengeParams, ", "))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		keyID, err := v.Verify(req)
		if err != nil {
			w.Header()["WWW-Authenticate"] = []string{challenge}
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, err.Error())
			return
		}
		h.ServeHTTP(w, req.WithContext(WithKeyID(req.Context(), keyID)))
	})
}
