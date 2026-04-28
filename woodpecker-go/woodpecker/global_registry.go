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
	pathGlobalRegistries = "%s/api/registries"
	pathGlobalRegistry   = "%s/api/registries/%s"
)

// GlobalRegistry returns an global registry by name.
func (c *client) GlobalRegistry(registry string) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathGlobalRegistry, c.addr, registry)
	err := c.get(uri, out)
	return out, err
}

// GlobalRegistryList returns a list of all global registries.
func (c *client) GlobalRegistryList(opt RegistryListOptions) ([]*Registry, error) {
	var out []*Registry
	uri, _ := url.Parse(fmt.Sprintf(pathGlobalRegistries, c.addr))
	uri.RawQuery = opt.getURLQuery().Encode()
	err := c.get(uri.String(), &out)
	return out, err
}

// GlobalRegistryCreate creates a global registry.
func (c *client) GlobalRegistryCreate(in *Registry) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathGlobalRegistries, c.addr)
	err := c.post(uri, in, out)
	return out, err
}

// GlobalRegistryUpdate updates a global registry.
func (c *client) GlobalRegistryUpdate(in *Registry) (*Registry, error) {
	out := new(Registry)
	uri := fmt.Sprintf(pathGlobalRegistry, c.addr, in.Address)
	err := c.patch(uri, in, out)
	return out, err
}

// GlobalRegistryDelete deletes a global registry.
func (c *client) GlobalRegistryDelete(registry string) error {
	uri := fmt.Sprintf(pathGlobalRegistry, c.addr, registry)
	return c.delete(uri)
}
