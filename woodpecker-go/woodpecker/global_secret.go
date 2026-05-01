// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package woodpecker

import (
	"fmt"
	"net/url"
)

const (
	pathGlobalSecrets = "%s/api/secrets"
	pathGlobalSecret  = "%s/api/secrets/%s"
)

// GlobalSecret returns an global secret by name.
func (c *client) GlobalSecret(secret string) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, secret)
	err := c.get(uri, out)
	return out, err
}

// GlobalSecretList returns a list of all global secrets.
func (c *client) GlobalSecretList(opt SecretListOptions) ([]*Secret, error) {
	var out []*Secret
	uri, _ := url.Parse(fmt.Sprintf(pathGlobalSecrets, c.addr))
	uri.RawQuery = opt.getURLQuery().Encode()
	err := c.get(uri.String(), &out)
	return out, err
}

// GlobalSecretCreate creates a global secret.
func (c *client) GlobalSecretCreate(in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecrets, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

// GlobalSecretUpdate updates a global secret.
func (c *client) GlobalSecretUpdate(in *Secret) (*Secret, error) {
	out := new(Secret)
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, in.Name)
	err := c.patch(uri, in, out)
	return out, err
}

// GlobalSecretDelete deletes a global secret.
func (c *client) GlobalSecretDelete(secret string) error {
	uri := fmt.Sprintf(pathGlobalSecret, c.addr, secret)
	return c.delete(uri)
}
