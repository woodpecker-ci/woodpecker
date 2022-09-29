// Copyright 2018 Drone.IO Inc.
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

package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mrjones/oauth"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

const (
	currentUserID    = "%s/plugins/servlet/applinks/whoami"
	pathUser         = "%s/rest/api/1.0/users/%s"
	pathRepo         = "%s/rest/api/1.0/projects/%s/repos/%s"
	pathRepos        = "%s/rest/api/1.0/repos?start=%s&limit=%s"
	pathHook         = "%s/rest/api/1.0/projects/%s/repos/%s/settings/hooks/%s"
	pathSource       = "%s/projects/%s/repos/%s/browse/%s?at=%s&raw"
	hookName         = "com.atlassian.stash.plugin.stash-web-post-receive-hooks-plugin:postReceiveHook"
	pathHookDetails  = "%s/rest/api/1.0/projects/%s/repos/%s/settings/hooks/%s"
	pathHookEnabled  = "%s/rest/api/1.0/projects/%s/repos/%s/settings/hooks/%s/enabled"
	pathHookSettings = "%s/rest/api/1.0/projects/%s/repos/%s/settings/hooks/%s/settings"
	pathStatus       = "%s/rest/build-status/1.0/commits/%s"
	pathBranches     = "%s/2.0/repositories/%s/%s/refs/branches"
)

type Client struct {
	client      *http.Client
	base        string
	accessToken string
	ctx         context.Context
}

func NewClientWithToken(ctx context.Context, url string, consumer *oauth.Consumer, AccessToken string) *Client {
	var token oauth.AccessToken
	token.Token = AccessToken
	client, err := consumer.MakeHttpClient(&token)
	if err != nil {
		log.Err(err).Msg("")
	}

	return &Client{
		client:      client,
		base:        url,
		accessToken: AccessToken,
		ctx:         ctx,
	}
}

func (c *Client) FindCurrentUser() (*User, error) {
	CurrentUserIDResponse, err := c.doGet(fmt.Sprintf(currentUserID, c.base))
	if CurrentUserIDResponse != nil {
		defer CurrentUserIDResponse.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	bits, err := io.ReadAll(CurrentUserIDResponse.Body)
	if err != nil {
		return nil, err
	}
	login := string(bits)

	CurrentUserResponse, err := c.doGet(fmt.Sprintf(pathUser, c.base, login))
	if CurrentUserResponse != nil {
		defer CurrentUserResponse.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	contents, err := io.ReadAll(CurrentUserResponse.Body)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(contents, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (c *Client) FindRepo(owner, name string) (*Repo, error) {
	urlString := fmt.Sprintf(pathRepo, c.base, owner, name)
	response, err := c.doGet(urlString)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Err(err).Msg("")
	}
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	repo := Repo{}
	err = json.Unmarshal(contents, &repo)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (c *Client) FindRepos() ([]*Repo, error) {
	return c.paginatedRepos(0)
}

func (c *Client) FindRepoPerms(owner, repo string) (*model.Perm, error) {
	perms := new(model.Perm)
	// If you don't have access return none right away
	_, err := c.FindRepo(owner, repo)
	if err != nil {
		return perms, err
	}
	// Must have admin to be able to list hooks. If have access the enable perms
	resp, err := c.doGet(fmt.Sprintf(pathHook, c.base, owner, repo, hookName))
	if resp != nil {
		defer resp.Body.Close()
	}
	if err == nil {
		perms.Push = true
		perms.Admin = true
	}
	perms.Pull = true
	return perms, nil
}

func (c *Client) FindFileForRepo(owner, repo, fileName, ref string) ([]byte, error) {
	response, err := c.doGet(fmt.Sprintf(pathSource, c.base, owner, repo, fileName, ref))
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		log.Err(err).Msg("")
	}
	if response.StatusCode == 404 {
		return nil, nil
	}
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Err(err).Msg("")
	}
	return responseBytes, nil
}

func (c *Client) CreateHook(owner, name, callBackLink string) error {
	hookDetails, err := c.GetHookDetails(owner, name)
	if err != nil {
		return err
	}
	var hooks []string
	if hookDetails.Enabled {
		hookSettings, err := c.GetHooks(owner, name)
		if err != nil {
			return err
		}
		hooks = hookSettingsToArray(hookSettings)
	}
	if !stringInSlice(callBackLink, hooks) {
		hooks = append(hooks, callBackLink)
	}

	putHookSettings := arrayToHookSettings(hooks)
	hookBytes, err := json.Marshal(putHookSettings)
	if err != nil {
		return err
	}
	return c.doPut(fmt.Sprintf(pathHookEnabled, c.base, owner, name, hookName), hookBytes)
}

func (c *Client) CreateStatus(revision string, status *PipelineStatus) error {
	uri := fmt.Sprintf(pathStatus, c.base, revision)
	return c.doPost(uri, status)
}

func (c *Client) DeleteHook(owner, name, link string) error {
	hookSettings, err := c.GetHooks(owner, name)
	if err != nil {
		return err
	}
	putHooks := filter(hookSettingsToArray(hookSettings), func(item string) bool {
		return !strings.Contains(item, link)
	})
	putHookSettings := arrayToHookSettings(putHooks)
	hookBytes, err := json.Marshal(putHookSettings)
	if err != nil {
		return err
	}
	return c.doPut(fmt.Sprintf(pathHookEnabled, c.base, owner, name, hookName), hookBytes)
}

func (c *Client) GetHookDetails(owner, name string) (*HookPluginDetails, error) {
	urlString := fmt.Sprintf(pathHookDetails, c.base, owner, name, hookName)
	response, err := c.doGet(urlString)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	hookDetails := HookPluginDetails{}
	err = json.NewDecoder(response.Body).Decode(&hookDetails)
	return &hookDetails, err
}

func (c *Client) GetHooks(owner, name string) (*HookSettings, error) {
	urlString := fmt.Sprintf(pathHookSettings, c.base, owner, name, hookName)
	response, err := c.doGet(urlString)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	hookSettings := HookSettings{}
	err = json.NewDecoder(response.Body).Decode(&hookSettings)
	return &hookSettings, err
}

// TODO: make these as as general do with the action

// Helper function to help create get
func (c *Client) doGet(url string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	return c.client.Do(request)
}

// Helper function to help create the hook
func (c *Client) doPut(url string, body []byte) error {
	request, err := http.NewRequestWithContext(c.ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return err
	}
	return nil
}

// Helper function to help create the hook
func (c *Client) doPost(url string, status *PipelineStatus) error {
	// write it to the body of the request.
	var buf io.ReadWriter
	if status != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(status)
		if err != nil {
			return err
		}
	}
	request, err := http.NewRequestWithContext(c.ctx, "POST", url, buf)
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if response != nil {
		defer response.Body.Close()
	}
	return err
}

// Helper function to get repos paginated
func (c *Client) paginatedRepos(start int) ([]*Repo, error) {
	limit := 1000
	requestURL := fmt.Sprintf(pathRepos, c.base, strconv.Itoa(start), strconv.Itoa(limit))
	response, err := c.doGet(requestURL)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	var repoResponse Repos
	err = json.NewDecoder(response.Body).Decode(&repoResponse)
	if err != nil {
		return nil, err
	}
	if !repoResponse.IsLastPage {
		reposList, err := c.paginatedRepos(start + limit)
		if err != nil {
			return nil, err
		}
		repoResponse.Values = append(repoResponse.Values, reposList...)
	}
	return repoResponse.Values, nil
}

func (c *Client) ListBranches(owner, name string) ([]*Branch, error) {
	uri := fmt.Sprintf(pathBranches, c.base, owner, name)
	response, err := c.doGet(uri)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	out := new(BranchResp)
	err = json.NewDecoder(response.Body).Decode(&out)
	return out.Values, err
}

func filter(vs []string, f func(string) bool) []string {
	var vsf []string
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// TODO: find a clean way of doing these next two methods- bitbucket server hooks only support 20 cb hooks
func arrayToHookSettings(hooks []string) HookSettings {
	hookSettings := HookSettings{}
	for loc, value := range hooks {
		switch loc {
		case 0:
			hookSettings.HookURL0 = value
		case 1:
			hookSettings.HookURL1 = value
		case 2:
			hookSettings.HookURL2 = value
		case 3:
			hookSettings.HookURL3 = value
		case 4:
			hookSettings.HookURL4 = value
		case 5:
			hookSettings.HookURL5 = value
		case 6:
			hookSettings.HookURL6 = value
		case 7:
			hookSettings.HookURL7 = value
		case 8:
			hookSettings.HookURL8 = value
		case 9:
			hookSettings.HookURL9 = value
		case 10:
			hookSettings.HookURL10 = value
		case 11:
			hookSettings.HookURL11 = value
		case 12:
			hookSettings.HookURL12 = value
		case 13:
			hookSettings.HookURL13 = value
		case 14:
			hookSettings.HookURL14 = value
		case 15:
			hookSettings.HookURL15 = value
		case 16:
			hookSettings.HookURL16 = value
		case 17:
			hookSettings.HookURL17 = value
		case 18:
			hookSettings.HookURL18 = value
		case 19:
			hookSettings.HookURL19 = value

			// Since there's only 19 hooks it will add to the latest if it doesn't exist :/
		default:
			hookSettings.HookURL19 = value
		}
	}
	return hookSettings
}

func hookSettingsToArray(hookSettings *HookSettings) []string {
	var hooks []string

	if hookSettings.HookURL0 != "" {
		hooks = append(hooks, hookSettings.HookURL0)
	}
	if hookSettings.HookURL1 != "" {
		hooks = append(hooks, hookSettings.HookURL1)
	}
	if hookSettings.HookURL2 != "" {
		hooks = append(hooks, hookSettings.HookURL2)
	}
	if hookSettings.HookURL3 != "" {
		hooks = append(hooks, hookSettings.HookURL3)
	}
	if hookSettings.HookURL4 != "" {
		hooks = append(hooks, hookSettings.HookURL4)
	}
	if hookSettings.HookURL5 != "" {
		hooks = append(hooks, hookSettings.HookURL5)
	}
	if hookSettings.HookURL6 != "" {
		hooks = append(hooks, hookSettings.HookURL6)
	}
	if hookSettings.HookURL7 != "" {
		hooks = append(hooks, hookSettings.HookURL7)
	}
	if hookSettings.HookURL8 != "" {
		hooks = append(hooks, hookSettings.HookURL8)
	}
	if hookSettings.HookURL9 != "" {
		hooks = append(hooks, hookSettings.HookURL9)
	}
	if hookSettings.HookURL10 != "" {
		hooks = append(hooks, hookSettings.HookURL10)
	}
	if hookSettings.HookURL11 != "" {
		hooks = append(hooks, hookSettings.HookURL11)
	}
	if hookSettings.HookURL12 != "" {
		hooks = append(hooks, hookSettings.HookURL12)
	}
	if hookSettings.HookURL13 != "" {
		hooks = append(hooks, hookSettings.HookURL13)
	}
	if hookSettings.HookURL14 != "" {
		hooks = append(hooks, hookSettings.HookURL14)
	}
	if hookSettings.HookURL15 != "" {
		hooks = append(hooks, hookSettings.HookURL15)
	}
	if hookSettings.HookURL16 != "" {
		hooks = append(hooks, hookSettings.HookURL16)
	}
	if hookSettings.HookURL17 != "" {
		hooks = append(hooks, hookSettings.HookURL17)
	}
	if hookSettings.HookURL18 != "" {
		hooks = append(hooks, hookSettings.HookURL18)
	}
	if hookSettings.HookURL19 != "" {
		hooks = append(hooks, hookSettings.HookURL19)
	}
	return hooks
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
