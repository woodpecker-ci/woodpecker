package frontend_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/compiler"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestParse(t *testing.T) {
	parsed, err := yaml.ParseString(sampleYaml)
	if err != nil {
		t.Error(err)
	}

	repoLink := ""
	defaultCloneImage := "clone-image"
	event := model.EventPush

	_secrets := []*model.Secret{
		{Name: "SURGE_TOKEN", Value: "123", Images: []string{"woodpeckerci/plugin-surge-preview-asda"}, Events: []model.WebhookEvent{model.EventPush}},
		{Name: "GITHUB_TOKEN_SURGE", Value: "456", Images: []string{"woodpeckerci/plugin-surge-preview-asda"}, Events: []model.WebhookEvent{model.EventPush}},
	}

	var secrets []compiler.Secret
	for _, sec := range _secrets {
		if !sec.Match(event) {
			continue
		}
		secrets = append(secrets, compiler.Secret{
			Name:  strings.ToUpper(sec.Name),
			Value: sec.Value,
			Match: sec.Images,
		})
	}

	config, err := compiler.New(
		// compiler.WithEnviron(environ),
		// compiler.WithEnviron(b.Envs),
		// compiler.WithEscalated(server.Config.Pipeline.Privileged...),
		// compiler.WithResourceLimit(server.Config.Pipeline.Limits.MemSwapLimit, server.Config.Pipeline.Limits.MemLimit, server.Config.Pipeline.Limits.ShmSize, server.Config.Pipeline.Limits.CPUQuota, server.Config.Pipeline.Limits.CPUShares, server.Config.Pipeline.Limits.CPUSet),
		// compiler.WithVolumes(server.Config.Pipeline.Volumes...),
		// compiler.WithNetworks(server.Config.Pipeline.Networks...),
		compiler.WithLocal(false),
		// compiler.WithOption(
		// 	compiler.WithNetrc(
		// 		b.Netrc.Login,
		// 		b.Netrc.Password,
		// 		b.Netrc.Machine,
		// 	),
		// 	b.Repo.IsSCMPrivate || server.Config.Pipeline.AuthenticatePublicRepos,
		// ),
		compiler.WithDefaultCloneImage(defaultCloneImage),
		// compiler.WithRegistry(registries...),
		compiler.WithSecret(secrets...),
		compiler.WithPrefix("test"),
		compiler.WithProxy(),
		compiler.WithWorkspaceFromURL("/woodpecker", repoLink),
		// compiler.WithMetadata(metadata),
	).Compile(parsed)

	assert.NoError(t, err)

	assert.Len(t, config.Stages, 4)
	assert.Len(t, config.Stages[3].Steps, 1)
	assert.Contains(t, config.Stages[3].Steps[0].Alias, "deploy-preview")
	assert.Contains(t, config.Stages[3].Steps[0].Environment, "PLUGIN_SURGE_TOKEN")
	assert.Equal(t, config.Stages[3].Steps[0].Environment["PLUGIN_SURGE_TOKEN"], "123")
}

var sampleYaml = `
pipeline:
  test:
    image: golang
    commands:
      - go install
      - go test
  build:
    image: golang
    network_mode: container:name
    commands:
      - go build
    when:
      event: push
  notify:
    image: slack
    channel: dev
    when:
      event: failure
  deploy-preview:
    image: woodpeckerci/plugin-surge-preview:next
    settings:
      path: "docs/build/"
      surge_token:
        from_secret: SURGE_TOKEN
      forge_type: github
      forge_url: "https://github.com"
      forge_repo_token:
        from_secret: GITHUB_TOKEN_SURGE

services:
  database:
    image: mysql

networks:
  custom:
    driver: overlay

volumes:
  custom:
    driver: blockbridge

depends_on:
  - lint
  - test
`

var simpleYamlAnchors = `
vars:
  image: &image plugins/slack
pipeline:
  notify_success:
    image: *image
`

var sampleVarYaml = `
_slack: &SLACK
  image: plugins/slack
pipeline:
  notify_fail: *SLACK
  notify_success:
    << : *SLACK
    when:
      event: success
`
