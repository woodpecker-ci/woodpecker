// Copyright 2024 Woodpecker Authors
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

package model

type ForgeType string

const (
	ForgeTypeGithub              ForgeType = "github"
	ForgeTypeGitlab              ForgeType = "gitlab"
	ForgeTypeGitea               ForgeType = "gitea"
	ForgeTypeForgejo             ForgeType = "forgejo"
	ForgeTypeBitbucket           ForgeType = "bitbucket"
	ForgeTypeBitbucketDatacenter ForgeType = "bitbucket-dc"
	ForgeTypeAddon               ForgeType = "addon"
)

// Additional-option keys of the github forge.
const (
	ForgeGithubOptionMergeRef           = "merge-ref"
	ForgeGithubOptionPublicOnly         = "public-only"
	ForgeGithubOptionAppID              = "app-id"
	ForgeGithubOptionAppPrivateKey      = "app-private-key"
	ForgeGithubOptionAppCloneTokenScope = "app-clone-token-scope"
)

// Additional-option keys of the bitbucket datacenter forge.
const (
	ForgeBitbucketDCOptionGitUsername = "git-username"
	ForgeBitbucketDCOptionGitPassword = "git-password"
	ForgeBitbucketDCOptionAdminScope  = "oauth-enable-project-admin-scope"
)

// Additional-option keys of the addon forge.
const (
	ForgeAddonOptionExecutable = "executable"
)

// secretForgeOptions lists the write-only additional-option keys per forge
// type. Their values never leave the server: API responses replace them with
// a "<key>-set" marker, and updates that omit or empty them keep the stored
// value.
var secretForgeOptions = map[ForgeType][]string{
	ForgeTypeGithub: {ForgeGithubOptionAppPrivateKey},
}

// SecretForgeOptions returns the write-only additional-option keys of the
// given forge type.
func SecretForgeOptions(forgeType ForgeType) []string {
	return secretForgeOptions[forgeType]
}

// SecretForgeOptionSetMarker returns the marker key that replaces a redacted
// secret option in API responses, so clients can tell a value is stored.
func SecretForgeOptionSetMarker(key string) string {
	return key + "-set"
}

type Forge struct {
	ID                int64          `json:"id"                           xorm:"pk autoincr 'id'"`
	Type              ForgeType      `json:"type"                         xorm:"VARCHAR(250)"`
	URL               string         `json:"url"                          xorm:"VARCHAR(500) 'url'"`
	OAuthClientID     string         `json:"client,omitempty"             xorm:"VARCHAR(250) 'oauth_client_id'"`
	OAuthClientSecret string         `json:"-"                            xorm:"VARCHAR(250) 'oauth_client_secret'"` // do not expose client secret
	SkipVerify        bool           `json:"skip_verify,omitempty"        xorm:"bool"`
	OAuthHost         string         `json:"oauth_host,omitempty"         xorm:"VARCHAR(250) 'oauth_host'"` // public url for oauth if different from url
	AdditionalOptions map[string]any `json:"additional_options,omitempty" xorm:"json"`
} //	@name	Forge

// TableName returns the database table name for xorm.
func (Forge) TableName() string {
	return "forges"
}

// PublicCopy returns a copy of the forge without sensitive information and technical details.
func (f *Forge) PublicCopy() *Forge {
	forge := &Forge{
		ID:   f.ID,
		Type: f.Type,
		URL:  f.URL,
	}

	return forge
}

// RedactSecrets removes secret values from the forge's additional options so
// the forge can be returned to API clients. Like OAuthClientSecret, secret
// options are write-only: clients keep the stored value by omitting the
// option on update. A redacted secret is replaced by an "<option>-set"
// marker so clients can tell whether a value is stored.
func (f *Forge) RedactSecrets() {
	if f == nil || f.AdditionalOptions == nil {
		return
	}
	for _, key := range SecretForgeOptions(f.Type) {
		if value, _ := f.AdditionalOptions[key].(string); value != "" {
			f.AdditionalOptions[SecretForgeOptionSetMarker(key)] = true
		}
		delete(f.AdditionalOptions, key)
	}
}

// ForgeWithOAuthClientSecret allows to update the client secret.
type ForgeWithOAuthClientSecret struct {
	Forge
	OAuthClientSecret string `json:"oauth_client_secret"`
} //	@name	ForgeWithOAuthClientSecret
