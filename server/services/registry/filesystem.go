// Copyright 2023 Woodpecker Authors
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

package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type filesystem struct {
	path string
}

func NewFilesystem(path string) ReadOnlyService {
	return &filesystem{path}
}

func parseDockerConfig(path string) ([]*model.Registry, error) {
	if path == "" {
		return nil, nil
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	configFile := configfile.ConfigFile{
		AuthConfigs: make(map[string]types.AuthConfig),
	}

	if err := json.NewDecoder(f).Decode(&configFile); err != nil {
		return nil, err
	}

	for registryHostname := range configFile.CredentialHelpers {
		newAuth, err := configFile.GetAuthConfig(registryHostname)
		if err == nil {
			configFile.AuthConfigs[registryHostname] = newAuth
		}
	}

	for addr, ac := range configFile.AuthConfigs {
		if ac.Auth != "" {
			ac.Username, ac.Password, err = decodeAuth(ac.Auth)
			if err != nil {
				return nil, err
			}
			ac.Auth = ""
			ac.ServerAddress = addr
			configFile.AuthConfigs[addr] = ac
		}
	}

	var auths []*model.Registry
	for key, auth := range configFile.AuthConfigs {
		auths = append(auths, &model.Registry{
			Address:  key,
			Username: auth.Username,
			Password: auth.Password,
		})
	}

	return auths, nil
}

func (f *filesystem) RegistryFind(*model.Repo, string) (*model.Registry, error) {
	return nil, nil
}

func (f *filesystem) RegistryList(_ *model.Repo, p *model.ListOptions) ([]*model.Registry, error) {
	regs, err := parseDockerConfig(f.path)
	if err != nil {
		return nil, err
	}
	return model.ApplyPagination(p, regs), nil
}

// decodeAuth decodes a base64 encoded string and returns username and password
func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}

	decLen := base64.StdEncoding.DecodedLen(len(authStr))
	decoded := make([]byte, decLen)
	authByte := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decoded, authByte)
	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", fmt.Errorf("something went wrong decoding auth config")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", fmt.Errorf("invalid auth configuration file")
	}
	password := strings.Trim(arr[1], "\x00")
	return arr[0], password, nil
}
