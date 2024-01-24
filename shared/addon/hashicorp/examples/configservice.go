package main

import (
	"fmt"
	"time"

	forgetypes "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/plugins/config"
	"go.woodpecker-ci.org/woodpecker/v2/shared/addon/hashicorp"
	"go.woodpecker-ci.org/woodpecker/v2/shared/addon/hashicorp/configservice"
)

type ExtensionImpl struct {
}

func (g *ExtensionImpl) FetchConfig(repo *model.Repo, pipeline *model.Pipeline, currentFileMeta []*forgetypes.FileMeta, netrc *model.Netrc, timeout time.Duration) (configData []*forgetypes.FileMeta, useOld bool, err error) {
	fmt.Println("hello world from hashicorp addon")
	return currentFileMeta[:len(currentFileMeta)-1], false, nil
}

func main() {
	hashicorp.Serve[config.Extension](configservice.Addon, &ExtensionImpl{})
}
