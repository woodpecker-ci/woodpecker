package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/types"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type filesystem struct {
	path string
}

func Filesystem(path string) model.ReadOnlyRegistryService {
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

func (b *filesystem) RegistryFind(*model.Repo, string) (*model.Registry, error) {
	return nil, nil
}

func (b *filesystem) RegistryList(*model.Repo) ([]*model.Registry, error) {
	return parseDockerConfig(b.path)
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
		return "", "", fmt.Errorf("Something went wrong decoding auth config")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", fmt.Errorf("Invalid auth configuration file")
	}
	password := strings.Trim(arr[1], "\x00")
	return arr[0], password, nil
}
