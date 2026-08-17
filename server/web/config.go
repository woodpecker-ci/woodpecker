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

package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware/session"
	"go.woodpecker-ci.org/woodpecker/v3/shared/token"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

func Config(c *gin.Context) {
	user := session.User(c)

	var csrf string
	if user != nil {
		t := token.New(token.CsrfToken)
		t.Set("user-id", strconv.FormatInt(user.ID, 10))
		csrf, _ = t.Sign(user.Hash)
	}

	var configPaths []string

	extensionsClean := make([]string, len(server.Config.Pipeline.ConfigExtensions))
	for i, e := range server.Config.Pipeline.ConfigExtensions {
		extensionsClean[i] = strings.TrimPrefix(e, ".")
	}

	extensions := strings.Join(extensionsClean, ",")
	for _, p := range server.Config.Pipeline.ConfigPaths {
		if strings.HasSuffix(p, "/") {
			// it's a directory -> add extensions
			configPaths = append(configPaths, fmt.Sprintf("%s*.{%s}", p, extensions))
		} else {
			configPaths = append(configPaths, p)
		}
	}

	configData := map[string]any{
		"user":                        user,
		"csrf":                        csrf,
		"version":                     version.String(),
		"skip_version_check":          server.Config.WebUI.SkipVersionCheck,
		"root_path":                   server.Config.Server.RootPath,
		"enable_swagger":              server.Config.WebUI.EnableSwagger,
		"user_registered_agents":      !server.Config.Agent.DisableUserRegisteredAgentRegistration,
		"max_pipeline_log_line_count": server.Config.WebUI.MaxPipelineLogLineCount,
		"default_config_paths":        configPaths,
	}

	// default func map with json parser.
	funcMap := template.FuncMap{
		"json": func(v any) string {
			a, err := json.Marshal(v)
			if err != nil {
				log.Error().Err(err).Msg("could not marshal JSON")
				return ""
			}
			return string(a)
		},
	}

	c.Header("Content-Type", "text/javascript; charset=utf-8")
	tmpl := template.Must(template.New("").Funcs(funcMap).Parse(configTemplate))

	if err := tmpl.Execute(c.Writer, configData); err != nil {
		log.Error().Err(err).Msg("could not execute template")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

const configTemplate = `
window.WOODPECKER_USER = {{ json .user }};
window.WOODPECKER_CSRF = "{{ .csrf }}";
window.WOODPECKER_VERSION = "{{ .version }}";
window.WOODPECKER_ROOT_PATH = "{{ .root_path }}";
window.WOODPECKER_ENABLE_SWAGGER = {{ .enable_swagger }};
window.WOODPECKER_SKIP_VERSION_CHECK = {{ .skip_version_check }}
window.WOODPECKER_USER_REGISTERED_AGENTS = {{ .user_registered_agents }}
window.WOODPECKER_MAX_PIPELINE_LOG_LINE_COUNT = {{ .max_pipeline_log_line_count }}
window.WOODPECKER_DEFAULT_CONFIG_PATHS = {{ json .default_config_paths }}
`
