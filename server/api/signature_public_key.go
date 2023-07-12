// Copyright 2021 Woodpecker Authors
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

package api

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server"
)

// GetSignaturePublicKey
//
//	@Summary	Get server's signature public key
//	@Router		/signature/public-key [get]
//	@Produce	plain
//	@Success	200
//	@Tags		System
//	@Param		Authorization	header	string	true	"Insert your personal access token"	default(Bearer <personal access token>)
func GetSignaturePublicKey(c *gin.Context) {
	b, err := x509.MarshalPKIXPublicKey(server.Config.Services.SignaturePublicKey)
	if err != nil {
		log.Error().Err(err).Msg("can't marshal public key")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}

	c.String(http.StatusOK, "%s", pem.EncodeToMemory(block))
}
