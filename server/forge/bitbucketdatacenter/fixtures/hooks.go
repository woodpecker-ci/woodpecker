package fixtures

import _ "embed"

//go:embed HookPullRequestOpenedFromFork.json
var HookPullFork string

//go:embed HookPush.json
var HookPush string

//go:embed HookPullRequestMerged.json
var HookPullMerged string

//go:embed HookPullRequestOpened.json
var HookPull string
