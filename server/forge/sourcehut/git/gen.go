package git

import (
	_ "git.sr.ht/~emersion/gqlclient"
)

//go:generate go run git.sr.ht/~emersion/gqlclient/cmd/gqlclientgen -s schema.graphqls -q queries.graphql -o gql.go -n git
